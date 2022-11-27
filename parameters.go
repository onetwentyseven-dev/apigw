package apigw

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

// UUIDPathParameter returns a parsed uuid.UUID from the google/uuid package for the given key from a
// APIGatewayV2HTTPRequest structs PathParameters map
func UUIDPathParameter(key string, event *events.APIGatewayV2HTTPRequest) (uuid.UUID, error) {

	parsedUUID, err := uuid.Parse(event.PathParameters[key])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse %s to valid uuid: %w", key, err)
	}

	return parsedUUID, nil

}
