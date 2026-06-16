package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// readInput reads from a file path, or stdin when path == "-".
func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// parseValue interprets s as JSON when possible, otherwise as a plain string.
func parseValue(s string) any {
	var v any
	if json.Unmarshal([]byte(s), &v) == nil {
		return v
	}
	return s
}

// buildBody assembles a request body for the generic `api` command from a data
// string, a data file, and key=value params (params take precedence).
func buildBody(data, dataFile string, params []string, rawBody bool) ([]byte, error) {
	if rawBody {
		if dataFile != "" {
			return readInput(dataFile)
		}
		return []byte(data), nil
	}

	obj := map[string]any{}
	have := false

	if dataFile != "" {
		b, err := readInput(dataFile)
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(string(b))) > 0 {
			if err := json.Unmarshal(b, &obj); err != nil {
				return nil, fmt.Errorf("parse --data-file as a JSON object: %w", err)
			}
			have = true
		}
	}

	if strings.TrimSpace(data) != "" {
		m := map[string]any{}
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			return nil, fmt.Errorf("parse --data as a JSON object: %w", err)
		}
		for k, v := range m {
			obj[k] = v
		}
		have = true
	}

	for _, p := range params {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --param %q (want key=value)", p)
		}
		obj[k] = parseValue(v)
		have = true
	}

	if !have {
		return []byte("{}"), nil
	}
	return json.Marshal(obj)
}

// jsonUnmarshalString decodes a JSON value (typically a bare string) into dst.
func jsonUnmarshalString(data json.RawMessage, dst *string) error {
	return json.Unmarshal(data, dst)
}

// readJSONObject parses a JSON object from a data string or a file ('-' = stdin).
// Exactly one source must be provided.
func readJSONObject(data, file string) (map[string]any, error) {
	var raw []byte
	switch {
	case data != "" && file != "":
		return nil, fmt.Errorf("provide only one of --data or --data-file")
	case data != "":
		raw = []byte(data)
	case file != "":
		b, err := readInput(file)
		if err != nil {
			return nil, err
		}
		raw = b
	default:
		return nil, fmt.Errorf("provide a JSON object via --data or --data-file")
	}
	obj := map[string]any{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("parse JSON object: %w", err)
	}
	return obj, nil
}

// readJSONValue parses an arbitrary JSON value from a data string or file.
func readJSONValue(data, file string) (any, error) {
	var raw []byte
	switch {
	case data != "" && file != "":
		return nil, fmt.Errorf("provide only one of --data or --data-file")
	case data != "":
		raw = []byte(data)
	case file != "":
		b, err := readInput(file)
		if err != nil {
			return nil, err
		}
		raw = b
	default:
		return nil, fmt.Errorf("provide JSON via --data or --data-file")
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return v, nil
}

// attrsFromPairs converts key=value pairs into a map. Keys listed in remove are
// set to nil (which deletes the attribute on the server).
func attrsFromPairs(pairs, remove []string) (map[string]any, error) {
	attrs := map[string]any{}
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --attr %q (want key=value)", p)
		}
		attrs[k] = v
	}
	for _, k := range remove {
		attrs[strings.TrimSpace(k)] = nil
	}
	if len(attrs) == 0 {
		return nil, fmt.Errorf("no attributes given (use --attr key=value or --remove key)")
	}
	return attrs, nil
}

// contentFlags holds the standard block-content flags shared by block commands.
type contentFlags struct {
	markdown    string
	dom         string
	data        string
	dataType    string
	contentFile string
}

func (cf *contentFlags) register(f *pflag.FlagSet) {
	f.StringVar(&cf.markdown, "markdown", "", "block content as Markdown/Kramdown")
	f.StringVar(&cf.dom, "dom", "", "block content as DOM/HTML")
	f.StringVar(&cf.data, "data", "", "raw block content (use with --data-type)")
	f.StringVar(&cf.dataType, "data-type", "", "data type: markdown|dom (default markdown)")
	f.StringVar(&cf.contentFile, "content-file", "", "read block content from a file ('-' for stdin)")
}

// resolve returns the (data, dataType) pair, enforcing that exactly one content
// source is provided.
func (cf *contentFlags) resolve(cmd *cobra.Command) (data, dataType string, err error) {
	ch := cmd.Flags().Changed
	dataType = "markdown"
	count := 0

	if cf.contentFile != "" {
		b, e := readInput(cf.contentFile)
		if e != nil {
			return "", "", e
		}
		data = string(b)
		if cf.dataType != "" {
			dataType = cf.dataType
		}
		count++
	}
	if ch("markdown") {
		data, dataType = cf.markdown, "markdown"
		count++
	}
	if ch("dom") {
		data, dataType = cf.dom, "dom"
		count++
	}
	if ch("data") {
		data = cf.data
		if cf.dataType != "" {
			dataType = cf.dataType
		}
		count++
	}

	switch {
	case count == 0:
		return "", "", fmt.Errorf("provide content via --markdown, --dom, --data (+--data-type), or --content-file")
	case count > 1:
		return "", "", fmt.Errorf("provide only one of --markdown, --dom, --data, --content-file")
	}
	if dataType != "markdown" && dataType != "dom" {
		return "", "", fmt.Errorf("invalid --data-type %q (want markdown or dom)", dataType)
	}
	return data, dataType, nil
}

// optString adds a string field to m when non-empty.
func optString(m map[string]any, key, val string) {
	if val != "" {
		m[key] = val
	}
}
