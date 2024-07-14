package apigw

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type transport struct {
	original http.RoundTripper

	config aws.Config
	signer *v4.Signer
}

func NewTransport(original http.RoundTripper, config aws.Config, signer *v4.Signer) http.RoundTripper {

	if original == nil {
		original = http.DefaultTransport
	}

	// func(options *v4.SignerOptions) {
	// 	options.LogSigning = true
	// 	options.Logger = logging.NewStandardLogger(os.Stdout)
	// }

	if signer == nil {
		signer = v4.NewSigner()
	}

	return &transport{
		original: original,
		config:   config,
		signer:   signer,
	}
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {

	var hashed string = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if r.Body != nil && r.ContentLength > 0 {

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		payloadSha := sha256.Sum256(data)

		hashed = hex.EncodeToString(payloadSha[:])

		fmt.Println("Hashed Body: ", hashed)
		r.Body = io.NopCloser(bytes.NewBuffer(data))

	}

	creds, err := t.config.Credentials.Retrieve(r.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	err = t.signer.SignHTTP(r.Context(), creds, r, hashed, "execute-api", t.config.Region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return t.original.RoundTrip(r)

}
