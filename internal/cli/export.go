package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/client"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export documents and resources",
	}
	cmd.AddCommand(
		newExportMdCmd(),
		newExportMdZipCmd(),
		newExportResourcesCmd(),
	)
	return cmd
}

func newExportMdCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "md <id> [--out <local-file>]",
		Short: "Export a document's Markdown content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := apiClient.Call(cmd.Context(), "/api/export/exportMdContent", map[string]any{"id": args[0]})
			if errors.Is(err, client.ErrDryRun) {
				return nil
			}
			if err != nil {
				return err
			}
			if out == "" {
				return render(resp)
			}
			var d struct {
				HPath   string `json:"hPath"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(resp.Data, &d); err != nil {
				return err
			}
			if err := os.WriteFile(out, []byte(d.Content), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "wrote %s (%d bytes) to %s\n", d.HPath, len(d.Content), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write Markdown to this local file")
	return cmd
}

func newExportMdZipCmd() *cobra.Command {
	var ids []string
	cmd := &cobra.Command{
		Use:   "md-zip --ids a,b",
		Short: "Export documents as a Markdown zip (returns the zip path)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/export/exportMds", map[string]any{"ids": ids})
		},
	}
	cmd.Flags().StringSliceVar(&ids, "ids", nil, "document IDs (comma-separated)")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

func newExportResourcesCmd() *cobra.Command {
	var (
		paths []string
		name  string
	)
	cmd := &cobra.Command{
		Use:   "resources --paths a,b [--name <zip-name>]",
		Short: "Export workspace resources as a zip (returns the zip path)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body := map[string]any{"paths": paths}
			optString(body, "name", name)
			return emit(cmd, "/api/export/exportResources", body)
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&paths, "paths", nil, "workspace-relative paths (comma-separated)")
	f.StringVar(&name, "name", "", "output zip name")
	_ = cmd.MarkFlagRequired("paths")
	return cmd
}
