package sigv4

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

func NewHTTPClient(config *aws.Config, service string, base *http.Client) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	return &http.Client{
		Transport: &transport{
			config:  config,
			service: service,
			base:    base,
		},
		CheckRedirect: base.CheckRedirect,
		Jar:           base.Jar,
		Timeout:       base.Timeout,
	}
}

type transport struct {
	config  *aws.Config
	service string
	base    *http.Client
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ph, err := payloadHash(req)
	if err != nil {
		return nil, signErr(fmt.Errorf("failed to calculate payload hash: %w", err))
	}

	ctx := req.Context()

	c, err := t.config.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, signErr(fmt.Errorf("failed to retrieve credentials: %w", err))
	}

	if err := v4.NewSigner().SignHTTP(ctx, c, req, ph, t.service, t.config.Region, time.Now()); err != nil {
		return nil, signErr(fmt.Errorf("failed to sign request: %w", err))
	}

	return t.base.Do(req)
}

func payloadHash(req *http.Request) (string, error) {
	if req.Body == nil {
		const emptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		return emptyHash, nil
	}

	body := req.Body
	defer body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(body); err != nil {
		return "", signErr(fmt.Errorf("failed to read request body: %w", err))
	}

	req.Body = io.NopCloser(buf)
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(buf.Bytes())), nil
	}

	h := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(h[:]), nil
}

func signErr(cause error) error {
	return fmt.Errorf("signature version 4 signing error: %w", cause)
}
