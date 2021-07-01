package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
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
}

var cmd = &cobra.Command{
	Use:  "awscurl",
	Args: cobra.ExactArgs(1),
	Run:  run,
}

func init() {
	cmd.Flags().StringVarP(&opt.Data, "data", "d", "", "Optional request body to send with the request")
	cmd.Flags().StringSliceVarP(&opt.Headers, "header", "H", nil, "Optional headers to include with the request")
	cmd.Flags().StringVarP(&opt.Method, "request", "X", "GET", "HTTP method [default: GET]")
	cmd.Flags().StringVarP(&opt.Service, "service", "s", "execute-api", "AWS service name [default: execute-api]")
}

func run(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, opt.Method, args[0], requestBody(opt.Data))
	if err != nil {
		log.Fatalf("failed to create HTTP request: %v", err)
	}

	setHeaders(req, opt.Headers)

	resp, err := sigv4.NewHTTPClient(&cfg, opt.Service, nil).Do(req)
	if err != nil {
		log.Fatalf("failed to HTTP request: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read HTTP response body: %v", err)
	}

	fmt.Print(string(b))
}

func requestBody(data string) io.Reader {
	if data == "" {
		return nil
	}
	return bytes.NewReader([]byte(data))
}

func setHeaders(req *http.Request, headers []string) {
	for _, h := range headers {
		a := strings.SplitN(h, ":", 2)
		if len(a) != 2 {
			continue
		}
		req.Header.Add(strings.TrimSpace(a[0]), strings.TrimSpace(a[1]))
	}
}
