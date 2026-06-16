package cli

import "github.com/spf13/cobra"

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tpl"},
		Short:   "Render templates",
	}

	var (
		path    string
		id      string
		preview bool
	)
	render := &cobra.Command{
		Use:   "render --path <workspace path> --id <block id>",
		Short: "Render a template file in the context of a block",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body := map[string]any{"path": path, "id": id}
			if preview {
				body["preview"] = true
			}
			return emit(cmd, "/api/template/render", body)
		},
	}
	rf := render.Flags()
	rf.StringVar(&path, "path", "", "template file path inside the workspace (required)")
	rf.StringVar(&id, "id", "", "context block ID (required)")
	rf.BoolVar(&preview, "preview", false, "render in preview mode")
	_ = render.MarkFlagRequired("path")
	_ = render.MarkFlagRequired("id")

	sprig := &cobra.Command{
		Use:   "sprig <template-string>",
		Short: "Render a Sprig template string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/template/renderSprig", map[string]any{"template": args[0]})
		},
	}

	cmd.AddCommand(render, sprig)
	return cmd
}
