package grafana

import (
	"log"
	"testing"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("GRAFANA_AGG_TEMPLATE", "/home/linclaus/work/golang/src/github.com/linclaus/stock-daemon/template/Aggregation.tmpl")
}

func TestAggregation(t *testing.T) {
	gd := MakeAggerationGrafanaDashboardResource("aggregation", "股市汇总", "vector(1)", "vector(1)", "vector(1)", "vector(1)")
	log.Println(gd)
	if err := CreateOrUpdateGrafanaDashboard(gd); err != nil {
		log.Println(err.Error())
	}
}
