package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSQLCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "sql [statement]",
		Short: "Run a read-only SQL query against the workspace database",
		Long: `Run a SQL query (SELECT) against SiYuan's SQLite database.

Useful tables: blocks (id, type, content, markdown, root_id, box, path, hpath,
parent_id, ...), attributes, spans, refs, file_annotation_refs.

Examples:
  penbridge sql "SELECT id, content FROM blocks WHERE type='p' LIMIT 5"
  penbridge sql --file query.sql`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var stmt string
			switch {
			case file != "":
				b, err := readInput(file)
				if err != nil {
					return err
				}
				stmt = string(b)
			case len(args) == 1:
				stmt = args[0]
			default:
				return fmt.Errorf("provide a SQL statement as an argument or via --file")
			}
			if strings.TrimSpace(stmt) == "" {
				return fmt.Errorf("empty SQL statement")
			}
			return emit(cmd, "/api/query/sql", map[string]any{"stmt": stmt})
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "read the SQL statement from a file ('-' for stdin)")
	return cmd
}

func newFlushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flush",
		Short: "Flush the transaction queue (persist pending writes before querying)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/sqlite/flushTransaction", nil)
		},
	}
}
