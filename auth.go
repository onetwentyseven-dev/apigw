package apigw

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

			authorization := event.Headers["authorization"]
			if authorization == "" {
				return &events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnauthorized,
				}, nil
			}

			if strings.HasPrefix(strings.ToLower(authorization), `bearer `) {
				authorization = strings.TrimPrefix(authorization, `bearer `)
				authorization = strings.TrimPrefix(authorization, `Bearer `)
			}

			token, err := jwt.ParseString(
				authorization,
				jwt.WithKeySet(jwks),
				jwt.WithClock(jwt.ClockFunc(time.Now().UTC)),
				jwt.WithIssuer(tenant),
				jwt.WithClaimValue("azp", clientID),
				jwt.WithAudience(audience),
				jwt.WithValidate(true),
			)
			if err != nil {
				return RespondJSONError(ctx, http.StatusUnauthorized, "", nil, err)
			}

			ctx = context.WithValue(ctx, UserContextKey, token)

			return next(ctx, event)

		}
	}, nil
}

func UserIDFromContext(ctx context.Context) (string, error) {

	token, err := TokenFromContext(ctx)
	if err != nil {
		return "", err
	}

	cliams := token.PrivateClaims()

	emailIface, ok := cliams["https://email"]
	if !ok {
		return "", fmt.Errorf("required key email missing from claims")
	}

	email, ok := emailIface.(string)
	if !ok {
		return "", fmt.Errorf("key email is %T expected string", emailIface)
	}

	md5Hash := md5.New()
	md5Hash.Write([]byte(email))
	hash := md5Hash.Sum(nil)

	return hex.EncodeToString(hash), nil

}
