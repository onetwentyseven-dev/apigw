package apigw

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-http-utils/headers"
)

func RestrictContentType(ct string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			contentType, ok := event.Headers[strings.ToLower(headers.ContentType)]
			if ok && contentType != ct {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnsupportedMediaType,
				}, nil
			}

			return next(ctx, event)

		}
	}
}
