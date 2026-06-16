package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newProxyCmd() *cobra.Command {
	var (
		method          string
		headers         []string
		payload         string
		payloadFile     string
		contentType     string
		payloadEncoding string
		respEncoding    string
		timeout         int
	)
	cmd := &cobra.Command{
		Use:   "proxy <url>",
		Short: "Forward an HTTP request through SiYuan (server-side fetch)",
		Long: `Send an HTTP request from the SiYuan kernel via /api/network/forwardProxy.

Private/loopback addresses are blocked by the server (SSRF protection).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{
				"url":              args[0],
				"method":           strings.ToUpper(method),
				"timeout":          timeout,
				"contentType":      contentType,
				"payloadEncoding":  payloadEncoding,
				"responseEncoding": respEncoding,
			}

			hdrs := make([]any, 0, len(headers))
			for _, h := range headers {
				k, v, ok := strings.Cut(h, ":")
				if !ok {
					return fmt.Errorf("invalid --header %q (want Key: Value)", h)
				}
				hdrs = append(hdrs, map[string]string{strings.TrimSpace(k): strings.TrimSpace(v)})
			}
			if len(hdrs) > 0 {
				body["headers"] = hdrs
			}

			raw := payload
			if payloadFile != "" {
				b, err := readInput(payloadFile)
				if err != nil {
					return err
				}
				raw = string(b)
			}
			if raw != "" {
				if payloadEncoding == "" || payloadEncoding == "json" {
					body["payload"] = parseValue(raw)
				} else {
					body["payload"] = raw
				}
			}
			return emit(cmd, "/api/network/forwardProxy", body)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&method, "method", "X", "GET", "HTTP method")
	f.StringArrayVarP(&headers, "header", "H", nil, "request header 'Key: Value' (repeatable)")
	f.StringVar(&payload, "payload", "", "request body")
	f.StringVar(&payloadFile, "payload-file", "", "read request body from a file ('-' for stdin)")
	f.StringVar(&contentType, "content-type", "application/json", "request Content-Type")
	f.StringVar(&payloadEncoding, "payload-encoding", "json", "payload encoding: json|text|base64|base64-url|base32|hex")
	f.StringVar(&respEncoding, "response-encoding", "text", "response body encoding")
	f.IntVar(&timeout, "request-timeout", 7000, "upstream request timeout in milliseconds")
	return cmd
}
