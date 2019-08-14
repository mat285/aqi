package server

import (
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/web"
	"github.com/mat285/slack/slack"
)

// Handler is a handler for the requests
type Handler func(*slack.SlashCommandRequest) (*slack.Message, error)

// Server is a slack server
type Server struct {
	App     *web.App
	Config  *Config
	Slack   *slack.Slack
	Handler Handler
}

// New returns a new slack server
func New(config *Config) *Server {
	s := &Server{
		Config: config,
		Slack:  slack.New([]byte(config.SlackSignatureSecret)),
	}
	return s
}

// WithHandler sets the handler on the server
func (s *Server) WithHandler(f Handler) *Server {
	s.Handler = f
	return s
}

// WithConfig sets the config on the server
func (s *Server) WithConfig(config *Config) *Server {
	s.Config = config
	return s
}

// Start gracefully starts the server starts the server and blocks until it exits
func (s *Server) Start() error {
	s.createApp()
	return graceful.Shutdown(s.App)
}

func (s *Server) createApp() {
	s.App = web.NewFromConfig(&s.Config.Config)
	s.App.POST("/", s.handle)
	s.App.GET("/healthz", s.healthz)
}

func (s *Server) handle(r *web.Ctx) web.Result {
	scr, err := s.Slack.VerifyRequest(r.Request())
	if err != nil {
		return r.JSON().NotAuthorized()
	}

	handler := s.Handler
	if handler == nil {
		handler = s.defaultHandler
	}

	if s.Config.AcknowledgeOnVerify {
		go s.handleAsync(scr, handler)
		return r.JSON().OK()
	}

	responseMessage, responseError := handler(scr)
	if responseError != nil {
		s.App.Logger().Error(err)
		return r.JSON().Result(s.errorMessage())
	}
	if responseMessage != nil {
		return r.JSON().Result(responseMessage)
	}
	return r.JSON().OK()
}

func (s *Server) handleAsync(scr *slack.SlashCommandRequest, handler Handler) {
	message, err := handler(scr)
	if err != nil {
		s.App.Logger().Error(err)
		err = slack.Notify(scr.ResponseURL, s.errorMessage())
		if err != nil {
			s.App.Logger().Error(err)
		}
		return
	}
	if message != nil {
		err = slack.Notify(scr.ResponseURL, message)
		if err != nil {
			s.App.Logger().Error(err)
		}
		return
	}
}

func (s *Server) defaultHandler(_ *slack.SlashCommandRequest) (*slack.Message, error) {
	return nil, nil
}

func (s *Server) errorMessage() *slack.Message {
	return &slack.Message{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         "Oops! Something went wrong with processing your request, please try again",
	}
}

func (s *Server) healthz(r *web.Ctx) web.Result {
	return r.JSON().Result(&Status{Ready: true})
}
