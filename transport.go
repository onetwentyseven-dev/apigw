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
	// signer v4.Signer
}

func NewTransport(original http.RoundTripper, config aws.Config) http.RoundTripper {

	if original == nil {
		original = http.DefaultTransport
	}

	return &transport{
		original: original,
		config:   config,
		// signer:   signer,
	}
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {

	cloned := r.Clone(r.Context())

	if cloned.Body == nil && cloned.ContentLength == 0 {
		cloned.Body = io.NopCloser(bytes.NewBufferString(""))
	}

	signer := v4.NewSigner()

	hash := sha256.New()

	_, err := io.Copy(hash, cloned.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to copy body to hash: %w", err)
	}

	hashed := hex.EncodeToString(hash.Sum(nil))

	creds, err := t.config.Credentials.Retrieve(r.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	err = signer.SignHTTP(cloned.Context(), creds, r, hashed, "execute-api", t.config.Region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return t.original.RoundTrip(r)

}
