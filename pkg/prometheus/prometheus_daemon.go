package prometheus

import (
	"log"
	"strings"
	"time"

	"github.com/go-gota/gota/series"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	codeIncreaseChan, codeDecreaseChan, codeSmoothChan chan string
)

func codeIncreaseChanConsumer() {
	ics := make([]string, 0)
	for code := range codeIncreaseChan {
		ics = append(ics, code)
	}
	if len(ics) != 0 {
		codeMap["ics"] = ics
	}
}
func codeDecreaseChanConsumer() {
	dcs := make([]string, 0)
	for code := range codeDecreaseChan {
		dcs = append(dcs, code)
	}
	if len(dcs) != 0 {
		codeMap["dcs"] = dcs
	}
}
func codeSmoothChanConsumer() {
	scs := make([]string, 0)
	for code := range codeSmoothChan {
		scs = append(scs, code)
	}
	if len(scs) != 0 {
		codeMap["scs"] = scs
	}
}

func LoadHistoryStockAggregation(duration string) {
	doJob(duration)
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(timezone))
	c.AddFunc("0 9,15 * * *", func() {
		logrus.Info("Start to compute aggregation")
		doJob(duration)
		logrus.Info("End to compute aggregation")
	})
	c.Start()
}

func doJob(duration string) {
	codeIncreaseChan = make(chan string, 100)
	codeDecreaseChan = make(chan string, 100)
	codeSmoothChan = make(chan string, 100)
	go codeIncreaseChanConsumer()
	go codeDecreaseChanConsumer()
	go codeSmoothChanConsumer()
	getHistoryStock(90)
	close(codeIncreaseChan)
	close(codeDecreaseChan)
	close(codeSmoothChan)
}

func getHistoryStock(days int) {
	todayStr := time.Now().Format("2006-01-02")
	today, _ := time.Parse("2006-01-02", todayStr)
	end := today.Add(time.Hour * 7)
	start := end.AddDate(0, 0, -1*days)
	logrus.Infof("Query Prometheus start: %s, end: %s", start.Format(timeFormat), end.Format(timeFormat))
	rsp, err := queryRange("stock_current_gauge{}", start.Format(timeFormat), end.Format(timeFormat), "1d")
	if err == nil || rsp.Status == "success" {
		for _, r := range rsp.Data.Result {
			s := series.New([]string{}, series.Float, r.Metric.Code)
			for _, v := range r.Values {
				if v[1] == "0" {
					continue
				}
				s.Append(v[1])
			}
			filterSeries(s)
		}
	} else {
		log.Println(err.Error())
	}
}

func filterSeries(s series.Series) {
	logrus.Infof("filter code: %s", s.Name)
	if strings.HasPrefix(s.Name, "sh000") || s.Len() < 5 {
		return
	}
	min := s.Min()
	max := s.Max()
	current := s.Val(s.Len() - 1).(float64)

	frontSetIndexes := make([]int, 0)
	for i := 0; i < s.Len()-5; i++ {
		if i < 0 {
			continue
		}
		frontSetIndexes = append(frontSetIndexes, i)
	}
	frontSet := s.Subset(frontSetIndexes)
	frontMin := frontSet.Min()
	frontMax := frontSet.Max()
	frontMean := frontSet.Mean()

	backSetIndexes := make([]int, 0)
	for i := s.Len() - 5; i < s.Len(); i++ {
		if i < 0 {
			continue
		}
		backSetIndexes = append(backSetIndexes, i)
	}
	backSet := s.Subset(backSetIndexes)
	backMin := backSet.Min()
	backMax := backSet.Max()
	backMean := backSet.Mean()
	if frontMean > backMean && frontMax == max && (frontMax-min)/frontMax*100 > viper.GetFloat64("HISTORY_WAVE_THRESHOLD") && (current-backMin)/backMin*100 > viper.GetFloat64("HISTORY_REBOUND_THRESHOLD") {
		logrus.Infof("Add code: %s to increase code set for history decrease: %.1f,mean: %.1f and current increase: %.1f,mean: %.1f", s.Name, (frontMax-min)/frontMax*100, frontMean, (current-backMin)/backMin*100, backMean)
		codeIncreaseChan <- s.Name
	}
	if frontMean < backMean && frontMin == min && (max-frontMin)/frontMin*100 > viper.GetFloat64("HISTORY_WAVE_THRESHOLD") && (backMax-current)/backMax*100 > viper.GetFloat64("HISTORY_REBOUND_THRESHOLD") {
		logrus.Infof("Add code: %s to decrease code set for history increase: %.1f,mean: %.1f and current decrease: %.1f,mean: %.1f", s.Name, (max-frontMin)/frontMin*100, frontMean, (backMax-current)/backMax*100, backMean)
		codeDecreaseChan <- s.Name
	}
	if frontMean < backMean && (frontMax-frontMin)/frontMin*100 < viper.GetFloat64("SMOOTH_WAVE_THRESHOLD") && (current-backMin)/backMin*100 > viper.GetFloat64("SMOOTH_REBOUND_THRESHOLD") {
		logrus.Infof("Add code: %s to smooth_increase code set for history wave: %.1f,mean: %.1f and current increase: %.1f,mean: %.1f", s.Name, (frontMax-frontMin)/frontMin*100, frontMean, (current-backMin)/backMin*100, backMean)
		codeSmoothChan <- s.Name
	}
}
