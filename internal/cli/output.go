package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/client"
)

// emit POSTs body to path and renders the response per the output format.
func emit(cmd *cobra.Command, path string, body any) error {
	return emitResp(apiClient.Call(cmd.Context(), path, body))
}

// emitResp renders a (response, error) pair, treating a dry-run as success.
func emitResp(resp *client.APIResponse, err error) error {
	if errors.Is(err, client.ErrDryRun) {
		return nil
	}
	if err != nil {
		return err
	}
	return render(resp)
}

func render(resp *client.APIResponse) error {
	if resp == nil {
		return nil
	}
	switch outputFormat {
	case "envelope":
		return printValue(resp, true)
	case "json":
		return printData(resp.Data, false)
	case "raw":
		return printRaw(resp.Data)
	case "pretty", "":
		return printData(resp.Data, true)
	default:
		return fmt.Errorf("unknown output format %q (use pretty|json|raw|envelope)", outputFormat)
	}
}

// printData prints the data payload. When indent is true, a bare JSON string is
// unquoted for readability; objects/arrays are pretty-printed.
func printData(data json.RawMessage, indent bool) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	if indent && trimmed[0] == '"' {
		var s string
		if json.Unmarshal(trimmed, &s) == nil {
			fmt.Println(s)
			return nil
		}
	}
	var buf bytes.Buffer
	if indent {
		if err := json.Indent(&buf, trimmed, "", "  "); err != nil {
			return err
		}
	} else {
		if err := json.Compact(&buf, trimmed); err != nil {
			return err
		}
	}
	fmt.Println(buf.String())
	return nil
}

func printRaw(data json.RawMessage) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if json.Unmarshal(trimmed, &s) == nil {
			fmt.Println(s)
			return nil
		}
	}
	return printData(data, false)
}

func printValue(v any, indent bool) error {
	var (
		out []byte
		err error
	)
	if indent {
		out, err = json.MarshalIndent(v, "", "  ")
	} else {
		out, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
