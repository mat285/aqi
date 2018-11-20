package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blend/go-sdk/env"
	logger "github.com/blend/go-sdk/logger"
	web "github.com/blend/go-sdk/web"
	"github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/slack"
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
	body, err := r.PostBody()
	if err != nil {
		return r.JSON().InternalError(err)
	}
	r.Request().Body = ioutil.NopCloser(bytes.NewReader(body))
	user := web.StringValue(r.Param(slack.ParamUserIDKey))
	text := web.StringValue(r.Param(slack.ParamTextKey))

	err = verify(r) // verify the request came from slack
	if err != nil {
		r.Logger().Error(err)
		return r.JSON().NotAuthorized()
	}

	if util.IsBlocked(user) && !strings.Contains(text, "please") {
		return r.JSON().Result(util.BlockedSlackMessage())
	}
	aqi, err := util.FetchAQI(conf, util.LocationRequestFromText(text), r.Logger())
	if err != nil {
		return r.JSON().InternalError(err)
	}
	if strings.Contains(text, "cigarettes") {
		return r.JSON().Result(util.CigarettesSlackMessage(aqi))
	}
	return r.JSON().Result(util.AQISlackMessage(aqi))
}

func verify(r *web.Ctx) error {
	timestamp, err := r.HeaderValue(slack.TimestampHeaderParam)
	if err != nil {
		return err
	}
	body, err := r.PostBody()
	if err != nil {
		return err
	}
	sig, err := r.HeaderValue(slack.SignatureHeaderParam)
	if err != nil {
		return err
	}
	return slack.VerifyRequest(timestamp, string(body), string(sig), env.Env().String(slack.EnvVarSignatureSecret))
}
