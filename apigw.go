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

type Handler func(context.Context, events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)

func New(lgr *logrus.Logger) *Service {
	return &Service{
		logger:   lgr,
		handlers: make(map[string]Handler),
	}
}

func (s *Service) addHandler(key string, handler Handler) {
	if _, ok := s.handlers[key]; ok {
		s.logger.WithField("key", key).Fatal("handler already registered for key")
	}

	s.handlers[key] = handler
}

func (s *Service) AddHandler(method, path string, handler Handler) {

	key := strings.Join([]string{method, path}, " ")

	s.addHandler(key, handler)

}

func (s *Service) HandleRoutes(ctx context.Context, input events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	// rk stands for routeKey
	rk := fmt.Sprintf("%s %s", input.HTTPMethod, input.Resource)

	if _, ok := s.handlers[rk]; !ok {
		return RespondJSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Route Not Found for %s", rk)}, nil)
	}

	return s.handlers[rk](ctx, input)
}
