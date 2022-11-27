package apigw

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger   *logrus.Logger
	handlers map[string]Handler
}

type Handler func(context.Context, events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error)

func New(lgr *logrus.Logger) *Service {
	return &Service{
		logger:   lgr,
		handlers: make(map[string]Handler),
	}
}

func (s *Service) AddHandler(key string, handler Handler) {
	if _, ok := s.handlers[key]; ok {
		s.logger.WithField("key", key).Fatal("handler already registered for key")
	}

	s.handlers[key] = handler
}

func (s *Service) AddHandlerMethod(method, path string, handler Handler) {

	key := strings.Join([]string{method, path}, " ")

	s.AddHandler(key, handler)

}

func (s *Service) HandleRoutes(ctx context.Context, input events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	if _, ok := s.handlers[input.RouteKey]; !ok {
		return RespondJSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Route Not Found for %s", input.RouteKey)}, nil)
	}

	return s.handlers[input.RouteKey](ctx, input)
}
