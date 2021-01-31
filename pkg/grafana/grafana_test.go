package grafana

import (
	"log"
	"testing"

	"github.com/linclaus/stock-daemon/pkg/kubernetes"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("GRAFANA_AGG_TEMPLATE", "/home/linclaus/work/golang/src/github.com/linclaus/stock-daemon/template/Aggregation.tmpl")
	viper.SetDefault("K8S_MASTER", "http://127.0.0.1:8001")
	kubernetes.Init()
}

func TestAggregation(t *testing.T) {
	gd := MakeAggerationGrafanaDashboardResource("aggregation", "股市汇总", "vector(1)", "vector(1)", "vector(1)", "vector(1)")
	log.Println(gd)
	if err := CreateOrUpdateGrafanaDashboard(gd); err != nil {
		log.Println(err.Error())
	}
	DeleteGrafanaDashboard("default", "aggregation")
}
