package main

import (
	"os"
	"os/signal"
	"syscall"

	logger "github.com/blend/go-sdk/logger"
	web "github.com/blend/go-sdk/web"
	"github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/util"
)

var conf *config.Config

func main() {
	log := logger.All()

	wc, err := web.NewConfigFromEnv()
	if err != nil {
		log.SyncFatalExit(err)
	}
	c, err := config.NewFromEnv()
	if err != nil {
		log.SyncFatalExit(err)
	}
	conf = c
	app := web.NewFromConfig(wc).WithLogger(log)

	app.GET("/", handle)
	app.POST("/", handle)

	quit := make(chan os.Signal, 1)
	// trap ^C
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	go func() {
		<-quit
		log.SyncError(app.Shutdown())
	}()

	if err := web.StartWithGracefulShutdown(app); err != nil {
		log.SyncFatalExit(err)
	}
}

func handle(r *web.Ctx) web.Result {
	aqi, err := util.FetchAQI(conf, util.SanFranciscoAirVisualRequest(), r.Logger())
	if err != nil {
		return r.Text().InternalError(err)
	}
	return r.Text().Result(util.SlackMessageText(aqi))
}