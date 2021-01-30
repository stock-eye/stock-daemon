package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

var (
	cfgFile string
)

func main() {
	fmt.Println("Stock-Daemon Started")
	// Set Args
	flag.StringVar(&cfgFile, "file", "", "The config file.")

	flag.Parse()

	ticker := time.NewTicker(time.Second * time.Duration(60))
	for {
		select {
		case <-ticker.C:
			fmt.Println("heatbreak")
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
