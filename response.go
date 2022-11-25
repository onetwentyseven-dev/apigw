package apigw

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

// Respond is a simple response with a status and body
func (s *Service) respond(status int, body string, headers map[string]string, isBase64Encoded bool) (*events.APIGatewayV2HTTPResponse, error) {

	if headers == nil {
		headers = map[string]string{}
	}

	e := &events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    headers,
	}

	if body != "" {
		e.Body = body
		e.IsBase64Encoded = isBase64Encoded
	}

	return e, nil
}

// RespondError returns a response while logging an error
func (s *Service) RespondError(status int, body string, headers map[string]string, err error) (*events.APIGatewayV2HTTPResponse, error) {
	return s.respond(status, body, headers, false)
}

// RespondJSON returns a json-formatted response
func (s *Service) RespondJSON(status int, body interface{}, headers map[string]string) (*events.APIGatewayV2HTTPResponse, error) {
	if headers == nil {
		headers = map[string]string{}
	}

	if body == nil {
		return s.respond(status, "", headers, false)
	}

	headers["Content-Type"] = "application/json"

	data, err := json.Marshal(body)
	if err != nil {
		return s.RespondError(status, `{"error": "an internal error occurred"}`, headers, err)
	}

	return s.respond(status, string(data), headers, false)
}

// RespondJSONError returns a json-formatted error response
func (s *Service) RespondJSONError(ctx context.Context, status int, msg string, headers map[string]string, err error) (*events.APIGatewayV2HTTPResponse, error) {
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"status": status,
		}).Error(msg)
	}

	b := map[string]string{"error": msg}

	return s.RespondJSON(status, b, headers)

}
