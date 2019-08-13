package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blend/go-sdk/env"
	logger "github.com/blend/go-sdk/logger"
	web "github.com/blend/go-sdk/web"
	"github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/util"
	"github.com/mat285/slack"
)

var (
	conf    *config.Config
	slacker *slack.Slack
)

const errMessage = "Oops! Something's not quite right"

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

	slacker = slack.New(env.Env().Bytes(slack.EnvVarSignatureSecret))

	file := env.Env().String("BLOCKED_USERS_FILE")
	_, err = os.Stat(file)
	if err == nil {
		err = util.BlockUsersFromFile(file)
		if err != nil {
			log.Error(err)
		}
	}

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
	sr, err := slacker.VerifyRequest(r.Request())
	if err != nil {
		r.Logger().Error(err)
		return r.JSON().NotAuthorized()
	}
	user := sr.UserID
	text := sr.Text

	if util.IsBlocked(user) && !strings.Contains(text, "please") {
		return r.JSON().Result(util.BlockedSlackMessage())
	}
	req := util.LocationRequestFromText(text)
	if req == nil {
		return r.JSON().Result(errMessage)
	}
	aqi, err := util.FetchAQI(conf, req, r.Logger())
	if err != nil {
		r.Logger().Error(err)
		return r.JSON().Result(errMessage)
	}
	if strings.Contains(text, "cigarettes") {
		return r.JSON().Result(util.CigarettesSlackMessage(aqi, req.City))
	}
	return r.JSON().Result(util.AQISlackMessage(aqi, req.City))
}
