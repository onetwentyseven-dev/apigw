package apigw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

type OpenIDConfiguration struct {
	JWKSURI string `json:"jwks_uri"`
}

func fetchOpenIDConfiguration(client *http.Client, tenant string) (*OpenIDConfiguration, error) {
	tenant = strings.TrimSuffix(tenant, "/")

	endpoint := fmt.Sprintf("%s/.well-known/openid-configuration", tenant)

	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to openid-configration endpoint: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to openid-configration endpoint: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code received from the openid-configration endpoint %s: %w", endpoint, err)
	}

	defer response.Body.Close()
	var openIDConfiguration = new(OpenIDConfiguration)

	err = json.NewDecoder(response.Body).Decode(openIDConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to decode openid-configration: %w", err)
	}

	return openIDConfiguration, nil

}

func fetchJWKS(client *http.Client, tenant string) (jwk.Set, error) {
	openIDConfiguration, err := fetchOpenIDConfiguration(client, tenant)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, openIDConfiguration.JWKSURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request to jwks-uri %s: %w", openIDConfiguration.JWKSURI, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to jwks-uri %s: %w", openIDConfiguration.JWKSURI, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code received from the jwks-uri %s: %w", openIDConfiguration.JWKSURI, err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jwks: %w", err)
	}

	set, err := jwk.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data to jwks: %w", err)
	}

	return set, nil

}

func Auth(client *http.Client, tenant, clientID, audience string) (Middleware, error) {

	jwks, err := fetchJWKS(client, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize jwks: %w", err)
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {

			authorization := event.Headers["authorization"]
			if authorization == "" {
				return &events.APIGatewayV2HTTPResponse{
					StatusCode: http.StatusUnauthorized,
				}, nil
			}

			token, err := jwt.ParseString(
				authorization,
				jwt.WithKeySet(jwks),
				jwt.WithIssuer(tenant),
				jwt.WithClock(jwt.ClockFunc(time.Now().UTC)),
				jwt.WithClaimValue("azp", audience),
			)
			if err != nil {
				return RespondError(http.StatusUnauthorized, "", nil, nil)
			}

			ctx = context.WithValue(ctx, UserContextKey, token)

			return next(ctx, event)

		}
	}, nil
}
