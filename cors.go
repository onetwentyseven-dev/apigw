package apigw

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-http-utils/headers"
)

type CorsOpts struct {
	Origins, Methods, Headers []string
}

var DefaultCorsOpt = &CorsOpts{
	Methods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
	Headers: []string{"Accept", "Content-Type", "User-Agent"},
	Origins: []string{"*"},
}

func Cors(opts *CorsOpts) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
			hders := buildCorsHeaders(opts)

			if req.RequestContext.HTTPMethod == http.MethodOptions {
				return &events.APIGatewayProxyResponse{
					StatusCode: http.StatusNoContent,
					Headers:    hders,
				}, nil
			}

			results, err := next(ctx, req)
			if err != nil {
				return results, err
			}

			if results.Headers == nil {
				results.Headers = make(map[string]string)
			}

			for h, v := range hders {
				results.Headers[h] = v
			}

			return results, nil
		}
	}
}

func buildCorsHeaders(opts *CorsOpts) map[string]string {
	hdrs := make(map[string]string)

	if len(opts.Headers) > 0 {
		hdrs[headers.AccessControlAllowHeaders] = strings.Join(opts.Headers, ",")
	}

	if len(opts.Methods) > 0 {
		hdrs[headers.AccessControlAllowMethods] = strings.Join(opts.Methods, ",")
	}
	if len(opts.Origins) > 0 {
		hdrs[headers.AccessControlAllowOrigin] = strings.Join(opts.Origins, ",")
	}

	return hdrs
}
