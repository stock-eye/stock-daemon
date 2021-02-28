package prometheus

import (
	"log"
	"time"

	"github.com/go-gota/gota/series"
)

var (
	codeIncreaseChan, codeDecreaseChan, codeSmoothChan chan string
)

func codeIncreaseChanConsumer() {
	ics := make([]string, 0)
	for code := range codeIncreaseChan {
		ics = append(ics, code)
	}
	codeMap["ics"] = ics
}
func codeDecreaseChanConsumer() {
	dcs := make([]string, 0)
	for code := range codeDecreaseChan {
		dcs = append(dcs, code)
	}
	codeMap["dcs"] = dcs
}
func codeSmoothChanConsumer() {
	scs := make([]string, 0)
	for code := range codeSmoothChan {
		scs = append(scs, code)
	}
	codeMap["scs"] = scs
}

func LoadHistoryStockAggregation(duration string) {
	doJob(duration)
	ticker := time.NewTicker(time.Hour * 24)
	for {
		select {
		case <-ticker.C:
			doJob(duration)
		}
	}
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
	rsp, err := queryRange("stock_current_gauge{}", start.Format(timeFormat), end.Format(timeFormat), "1d")
	if err == nil || rsp.Status == "success" {
		for _, r := range rsp.Data.Result {
			s := series.New([]string{}, series.Float, r.Metric.Code)
			for _, v := range r.Values {
				s.Append(v[1])
			}
			validateSeries(s)
		}
	} else {
		log.Println(err.Error())
	}
}

func validateSeries(s series.Series) {
	min := s.Min()
	max := s.Max()
	current := s.Val(s.Len() - 1).(float64)
	subSetIndexes := make([]int, 0)
	for i := s.Len() - 10; i < s.Len(); i++ {
		if i < 0 {
			continue
		}
		subSetIndexes = append(subSetIndexes, i)
	}
	sub := s.Subset(subSetIndexes)
	subMin := sub.Min()
	subMax := sub.Max()
	if (max-min)/max*100 > 40 && (current-min)/min > 10 && max != subMax {
		codeDecreaseChan <- s.Name
	}
	if (max-min)/min*100 > 40 && (max-current)/max > 10 && min != subMin {
		codeIncreaseChan <- s.Name
	}
	if (max-min)/min*100 < 10 {
		codeSmoothChan <- s.Name
	}
}
