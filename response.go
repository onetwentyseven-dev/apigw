package apigw

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
)

// Respond is a simple response with a status and body
func Respond(status int, body string, headers map[string]string, isBase64Encoded bool) (*events.APIGatewayProxyResponse, error) {

	if headers == nil {
		headers = map[string]string{}
	}

	e := &events.APIGatewayProxyResponse{
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
func RespondError(status int, body string, headers map[string]string, err error) (*events.APIGatewayProxyResponse, error) {
	return Respond(status, body, headers, false)
}

// RespondJSON returns a json-formatted response
func RespondJSON(status int, body interface{}, headers map[string]string) (*events.APIGatewayProxyResponse, error) {
	if headers == nil {
		headers = map[string]string{}
	}

	if body == nil {
		return Respond(status, "", headers, false)
	}

	headers["Content-Type"] = "application/json"

	data, err := json.Marshal(body)
	if err != nil {
		return RespondError(status, `{"error": "an internal error occurred"}`, headers, err)
	}

	return Respond(status, string(data), headers, false)
}

// RespondJSONError returns a json-formatted error response
func RespondJSONError(ctx context.Context, status int, msg string, headers map[string]string, err error) (*events.APIGatewayProxyResponse, error) {
	if err != nil {
		logrus.New().WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"status": status,
		}).Error(msg)
	}

	b := map[string]string{"error": msg}

	return RespondJSON(status, b, headers)

}
