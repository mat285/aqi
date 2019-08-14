package server

import (
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/web"
	"github.com/mat285/slack/slack"
)

// HandleFunc is handles the requests
type HandleFunc func(*slack.SlashCommandRequest) (*slack.Message, error)

// Server is a slack server
type Server struct {
	App        *web.App
	Config     *Config
	Slack      *slack.Slack
	HandleFunc HandleFunc
}

// New returns a new slack server
func New(config *Config) *Server {
	app := web.NewFromConfig(&config.Config)
	s := &Server{
		App:    app,
		Config: config,
		Slack:  slack.New([]byte(config.SlackSignatureSecret)),
	}
	return s
}

// WithHandleFunc sets the handle func on the server
func (s *Server) WithHandleFunc(f HandleFunc) *Server {
	s.HandleFunc = f
	return s
}

// Start gracefully starts the server starts the server and blocks until it exits
func (s *Server) Start() error {
	s.App.POST("/", s.handle)
	s.App.GET("/healthz", s.healthz)
	return graceful.Shutdown(s.App)
}

func (s *Server) handle(r *web.Ctx) web.Result {
	scr, err := s.Slack.VerifyRequest(r.Request())
	if err != nil {
		return r.JSON().NotAuthorized()
	}

	handler := s.HandleFunc
	if handler == nil {
		handler = s.defaultHandleFunc
	}

	if s.Config.AcknowledgeOnVerify {
		go s.handleAsync(scr, handler)
		return r.JSON().OK()
	}

	responseMessage, responseError := handler(scr)
	if responseError != nil {
		return r.JSON().InternalError(responseError)
	}
	if responseMessage != nil {
		return r.JSON().Result(responseMessage)
	}
	return r.JSON().OK()
}

func (s *Server) handleAsync(scr *slack.SlashCommandRequest, handler HandleFunc) {
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
		err := slack.Notify(scr.ResponseURL, message)
		if err != nil {
			s.App.Logger().Error(err)
		}
		return
	}
}

func (s *Server) defaultHandleFunc(_ *slack.SlashCommandRequest) (*slack.Message, error) {
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
