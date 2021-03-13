package prometheus

import (
	"fmt"
	"testing"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/spf13/viper"
)

func init() {
	viper.Set("PROMETHEUS_HOST", "http://192.168.1.27:30090")
	viper.Set("HISTORY_WAVE_THRESHOLD", "20")
	viper.Set("HISTORY_REBOUND_THRESHOLD", "10")
	viper.Set("SMOOTH_WAVE_THRESHOLD", "10")
	viper.Set("SMOOTH_REBOUND_THRESHOLD", "20")
}

func TestGetStockHistoryDataFrame(t *testing.T) {
	LoadHistoryStockAggregation("90d")
	df := GetStockHistoryDataFrame(90)
	fmt.Println(df.Describe())
}

func TestGetAggregate10IncreaseExpr(t *testing.T) {
	GetAggregate10IncreaseExpr("3d")
}

func Test1(t *testing.T) {
	df := dataframe.LoadRecords(
		[][]string{
			[]string{"A", "B", "C", "D"},
			[]string{"a", "4", "5.1", "true"},
			[]string{"b", "4", "6.0", "true"},
			[]string{"c", "3", "6.0", "false"},
			[]string{"a", "2", "7.1", "false"},
		},
	)
	fmt.Println(df)
	df = df.Capply(s)
	fmt.Println(df)
}
func s(s series.Series) series.Series {
	fmt.Println(s)
	fmt.Println(s.Min())
	fmt.Println(s.Max())
	fmt.Println(s.Len())
	fmt.Println(s.Val(s.Len()))
	s.Append("1")
	return s
}
