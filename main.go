package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"

	"rxnew/awscurl/sigv4"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var opt struct {
	Data    string
	Headers []string
	Method  string
	Service string
	Retry   uint
	Fail    bool
}

var cmd = &cobra.Command{
	Use:  "awscurl [URL]",
	Args: cobra.ExactArgs(1),
	Run:  run,
}

func init() {
	cmd.Flags().StringVarP(&opt.Data, "data", "d", "", "Optional request body to send with the request")
	cmd.Flags().StringSliceVarP(&opt.Headers, "header", "H", nil, "Optional headers to include with the request")
	cmd.Flags().StringVarP(&opt.Method, "request", "X", "GET", "HTTP method [default: GET]")
	cmd.Flags().StringVarP(&opt.Service, "service", "s", "execute-api", "AWS service name [default: execute-api]")
	cmd.Flags().UintVar(&opt.Retry, "retry", 0, "Retry a specified number of times [default: 0]")
	cmd.Flags().BoolVarP(&opt.Fail, "fail", "f", false, "Fail silently (no output at all) on HTTP errors")
}

func run(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	cfg, err := loadConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	req, err := request(ctx, args[0])
	if err != nil {
		log.Fatalf("failed to create HTTP request: %v", err)
	}

	resp, err := sigv4.NewHTTPClient(cfg, opt.Service, newRetryableHTTPClient()).Do(req)
	if err != nil {
		log.Fatalf("failed to HTTP request: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read HTTP response body: %v", err)
	}

	if opt.Fail && resp.StatusCode >= 400 {
		fmt.Printf("awscurl: (22) The requested URL returned error: %d\n", resp.StatusCode)
		os.Exit(22)
	}

	fmt.Print(string(b))
}

func request(ctx context.Context, url string) (*http.Request, error) {
	b, err := requestBody(opt.Data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, opt.Method, url, b)
	if err != nil {
		return nil, err
	}

	for _, h := range opt.Headers {
		a := strings.SplitN(h, ":", 2)
		if len(a) != 2 {
			return nil, fmt.Errorf("invalid request header [%s]", h)
		}
		req.Header.Add(strings.TrimSpace(a[0]), strings.TrimSpace(a[1]))
	}

	return req, nil
}

func requestBody(data string) (io.Reader, error) {
	if data == "" {
		return nil, nil
	}

	if data[0] == '@' && len(data) > 1 {
		b, err := os.ReadFile(data[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return bytes.NewReader(removeNewline(b)), nil
	}

	return bytes.NewReader(removeNewline([]byte(data))), nil
}

func removeNewline(b []byte) []byte {
	return []byte(strings.NewReplacer("\r\n", "", "\r", "", "\n", "").Replace(string(b)))
}

func newRetryableHTTPClient() *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = int(opt.Retry)
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 10 * time.Minute
	retryClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		return resp, err
	}
	retryClient.Logger = nil
	return retryClient.StandardClient()
}
