package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newTxCmd() *cobra.Command {
	var (
		data string
		file string
	)
	cmd := &cobra.Command{
		Use:   "tx (--data '<json>' | --data-file -)",
		Short: "Run low-level transactions via /api/transactions (advanced editing)",
		Long: `Perform raw transactions against the kernel.

Input may be either:
  - a transactions array: [ {"doOperations":[...],"undoOperations":[...]} ]
    which is wrapped automatically with a generated reqId; or
  - a full request object: {"transactions":[...], "reqId":<ms>}

Each operation looks like:
  {"action":"update","id":"<block id>","data":"<DOM html>"}

Common actions: insert, update, delete, move, appendInsert, prependInsert,
foldHeading, unfoldHeading, setAttrs. Operation 'data' is typically block DOM.
This is an expert tool; prefer the 'block' commands for everyday edits.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			v, err := readJSONValue(data, file)
			if err != nil {
				return err
			}
			var body map[string]any
			switch val := v.(type) {
			case []any:
				body = map[string]any{"transactions": val, "reqId": time.Now().UnixMilli()}
			case map[string]any:
				body = val
				if _, ok := body["transactions"]; !ok {
					return fmt.Errorf("object input must contain a 'transactions' array")
				}
				if _, ok := body["reqId"]; !ok {
					body["reqId"] = time.Now().UnixMilli()
				}
			default:
				return fmt.Errorf("input must be a transactions array or a request object")
			}
			return emit(cmd, "/api/transactions", body)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&data, "data", "d", "", "transactions JSON (array or full object)")
	f.StringVar(&file, "data-file", "", "read transactions JSON from a file ('-' for stdin)")
	return cmd
}
