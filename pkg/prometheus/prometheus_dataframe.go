package prometheus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"log"

	"github.com/go-gota/gota/dataframe"
	"github.com/spf13/viper"
)

var (
	timeFormat = "2006-01-02T15:04:05Z"
)

type PrometheusQueryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Code string `json:"code"`
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func GetStockHistoryDataFrame(days int) dataframe.DataFrame {
	var stockDf dataframe.DataFrame
	todayStr := time.Now().Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)
	end := today.Add(time.Hour * 7)
	start := end.AddDate(0, 0, -1*days)
	rsp, err := queryRange("stock_current_gauge{}", start.Format(timeFormat), end.Format(timeFormat), "1d")
	if err == nil || rsp.Status == "success" {
		for i, r := range rsp.Data.Result {
			var df dataframe.DataFrame
			var records [][]string

			records = [][]string{
				[]string{r.Metric.Code},
			}

			for _, v := range r.Values {
				records = append(records, []string{v[1].(string)})
			}
			if i != 0 {
				cnt := len(records)
				for i := -1; i < stockDf.Nrow()-cnt; i++ {
					records = append(records, records[cnt-1])
				}
			}
			df = dataframe.LoadRecords(records)
			stockDf = stockDf.CBind(df)
		}
	} else {
		log.Println(err.Error())
	}
	return stockDf
}

func queryRange(query, start, end, step string) (*PrometheusQueryRangeResponse, error) {
	rsp, err := http.Get(fmt.Sprintf(viper.GetString("PROMETHEUS_HOST")+"/api/v1/query_range?query=%s&start=%s&end=%s&step=%s", query, start, end, step))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	prb, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	pr := PrometheusQueryRangeResponse{}
	err = json.Unmarshal(prb, &pr)
	return &pr, err
}
