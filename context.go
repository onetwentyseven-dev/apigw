package apigw

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/jwt"
)

type contextKey struct{}

var (
	UserContextKey contextKey = struct{}{}
)

func TokenFromContext(ctx context.Context) (jwt.Token, error) {

	tokenIface := ctx.Value(UserContextKey)
	if tokenIface == nil {
		return nil, fmt.Errorf("ctx key is nil")
	}

	token, ok := tokenIface.(jwt.Token)
	if !ok {
		return nil, fmt.Errorf("ctx value is not expected type")
	}

	return token, nil

}
