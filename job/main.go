package main

import (
	logger "github.com/blend/go-sdk/logger"
	config "github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/util"
)

func main() {
	agent := logger.All()
	conf, err := config.NewFromEnv()
	if err != nil {
		agent.SyncFatalExit(err)
	}
	_, err = util.FetchAndSendAQIForConfig(conf, util.SanFranciscoAirVisualRequest(), agent)
	if err != nil {
		agent.SyncFatalExit(err)
	}
}
