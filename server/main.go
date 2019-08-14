package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blend/go-sdk/env"
	logger "github.com/blend/go-sdk/logger"
	web "github.com/blend/go-sdk/web"
	"github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/util"
	slackserver "github.com/mat285/slack/server"
	"github.com/mat285/slack/slack"
)

var (
	conf *config.Config
	log  *logger.Logger
)

const errMessage = "Oops! Something's not quite right"

func main() {
	log = logger.All()

	wc, err := web.NewConfigFromEnv()
	if err != nil {
		log.SyncFatalExit(err)
	}
	c, err := config.NewFromEnv()
	if err != nil {
		log.SyncFatalExit(err)
	}
	conf = c

	sc := &slackserver.Config{
		Config:               *wc,
		AcknowledgeOnVerify:  false,
		SlackSignatureSecret: env.Env().String(slack.EnvVarSignatureSecret),
	}

	file := env.Env().String("BLOCKED_USERS_FILE")
	_, err = os.Stat(file)
	if err == nil {
		err = util.BlockUsersFromFile(file)
		if err != nil {
			log.Error(err)
		}
	}

	serv := slackserver.New(sc).WithHandler(handle)
	err = serv.Start()
	if err != nil {
		log.SyncFatalExit(err)
	}
}

func handle(sr *slack.SlashCommandRequest) (*slack.Message, error) {
	user := sr.UserID
	text := sr.Text

	if util.IsBlocked(user) && !strings.Contains(text, "please") {
		return util.BlockedSlackMessage(), nil
	}
	req := util.LocationRequestFromText(text)
	if req == nil {
		return nil, fmt.Errorf(errMessage)
	}
	aqi, err := util.FetchAQI(conf, req, log)
	if err != nil {
		return nil, err
	}
	if strings.Contains(text, "cigarettes") {
		return util.CigarettesSlackMessage(aqi, req.City), nil
	}
	return util.AQISlackMessage(aqi, req.City), nil
}
