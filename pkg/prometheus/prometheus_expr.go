package prometheus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type PrometheusVectorResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]interface{} `json:"metric"`
		} `json:"result"`
	} `json:"data"`
}

var codeMap = map[string][]string{}

func GetAggregateIncreaseExpr() string {
	codes := queryPrometheusForCodes("9.9<(max_over_time(stock_increase_gauge[10m])<10.1)%20and%20idelta(stock_increase_gauge[5m])==0")
	if len(codes) == 0 {
		codes = codeMap["ie"]
	}
	if len(codes) > 0 {
		codeMap["ie"] = codes
		cstr := strings.Join(codes, "|")
		expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
		return expr
	}

	return ""
}

func GetAggregateDecreaseExpr() string {
	codes := queryPrometheusForCodes("-10.1<(min_over_time(stock_increase_gauge[10m])<-9.9)%20and%20idelta(stock_increase_gauge[5m])==0")
	if len(codes) == 0 {
		codes = codeMap["de"]
	}
	if len(codes) > 0 {
		codeMap["de"] = codes
		cstr := strings.Join(codes, "|")
		expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
		return expr
	}
	return ""
}

func GetAggregate10IncreaseExpr() string {
	codes := queryPrometheusForCodes("-10.1<min_over_time(stock_increase_gauge[10m])<-9.9%20and%20stock_increase_gauge>-9")
	if len(codes) == 0 {
		codes = codeMap["10ie"]
	}
	if len(codes) > 0 {
		codeMap["10ie"] = codes
		cstr := strings.Join(codes, "|")
		expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
		return expr
	}
	return ""
}

func GetAggregate10DecreaseExpr() string {
	codes := queryPrometheusForCodes("10.1>max_over_time(stock_increase_gauge[10m])>9.9%20and%20stock_increase_gauge<9")
	if len(codes) == 0 {
		codes = codeMap["10de"]
	}
	if len(codes) > 0 {
		codeMap["10de"] = codes
		cstr := strings.Join(codes, "|")
		expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
		return expr
	}
	return ""
}

func GetHistoryIncreaseExpr() string {
	codes := codeMap["ics"]
	cstr := strings.Join(codes, "|")
	expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
	return expr
}

func GetHistoryDecreaseExpr() string {
	codes := codeMap["dcs"]
	cstr := strings.Join(codes, "|")
	expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
	return expr
}

func GetHistorySmoothExpr() string {
	codes := codeMap["scs"]
	cstr := strings.Join(codes, "|")
	expr := fmt.Sprintf("stock_increase_gauge{code=~\"%s\"}", cstr)
	return expr
}

func queryPrometheusForCodes(queryString string) []string {
	resp, err := http.Get(fmt.Sprintf(viper.GetString("PROMETHEUS_HOST")+"/api/v1/query?query=%s", queryString))
	var codes []string
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				log.Println(err)
				return codes
			}
			var r PrometheusVectorResponse
			json.Unmarshal(body, &r)
			for _, rst := range r.Data.Result {
				codes = append(codes, rst.Metric["code"].(string))
			}
		}
	}
	return codes
}
