package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/linclaus/stock-daemon/pkg/grafana"
	"github.com/linclaus/stock-daemon/pkg/kubernetes"
	"github.com/linclaus/stock-daemon/pkg/prometheus"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

func main() {
	fmt.Println("Stock-Daemon Started")
	// Set Args
	flag.StringVar(&cfgFile, "file", "", "The config file.")

	flag.Parse()

	initConfig()
	kubernetes.Init()

	go prometheus.LoadHistoryStockAggregation(viper.GetString("HISTORY_AGGREGATION_DURATION"))

	ticker := time.NewTicker(time.Second * time.Duration(60))
	for {
		select {
		case <-ticker.C:
			ie := prometheus.GetAggregateIncreaseExpr()
			ie10 := prometheus.GetAggregate10IncreaseExpr(viper.GetString("AGGREGATION_DURATION"))
			de := prometheus.GetAggregateDecreaseExpr()
			de10 := prometheus.GetAggregate10DecreaseExpr(viper.GetString("AGGREGATION_DURATION"))
			ics := prometheus.GetHistoryIncreaseExpr()
			dcs := prometheus.GetHistoryDecreaseExpr()
			scs := prometheus.GetHistorySmoothExpr()
			gd := grafana.MakeAggerationGrafanaDashboardResource("aggregation", "股市汇总", ie, ie10, de, de10, ics, dcs, scs)
			if err := grafana.CreateOrUpdateGrafanaDashboard(gd); err != nil {
				log.Println(err.Error())
			}
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.AddConfigPath("/etc/")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
