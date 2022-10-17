package apigw

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type Handler func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Route struct {
	Method, Path string
}

func (r *Route) routeKey() string {
	return fmt.Sprintf("%s %s", r.Method, r.Path)
}

func (s *Service) HandleRoutes(routes map[Route]Handler) Handler {
	mapped := make(map[string]Handler)
	for route, handler := range routes {
		mapped[route.routeKey()] = handler
	}

	return func(ctx context.Context, input events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if handler, ok := mapped[input.Resource]; ok {
			return handler(ctx, input)
		}

		resp := map[string]string{"error": fmt.Sprintf("Route Not Found for %s", input.Resource)}

		return s.RespondJSON(http.StatusNotFound, resp, nil)
	}

}
