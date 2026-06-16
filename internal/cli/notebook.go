package cli

import "github.com/spf13/cobra"

func newNotebookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notebook",
		Aliases: []string{"nb"},
		Short:   "Manage notebooks (list/open/close/create/rename/remove/info/conf)",
	}

	var flashcard bool
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List notebooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/notebook/lsNotebooks", map[string]any{"flashcard": flashcard})
		},
	}
	ls.Flags().BoolVar(&flashcard, "flashcard", false, "only notebooks containing flashcards")

	open := &cobra.Command{
		Use:   "open <notebook-id>",
		Short: "Open (mount) a notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/openNotebook", map[string]any{"notebook": args[0]})
		},
	}

	closeCmd := &cobra.Command{
		Use:   "close <notebook-id>",
		Short: "Close (unmount) a notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/closeNotebook", map[string]any{"notebook": args[0]})
		},
	}

	create := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a notebook (returns the new notebook)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/createNotebook", map[string]any{"name": args[0]})
		},
	}

	remove := &cobra.Command{
		Use:   "remove <notebook-id>",
		Short: "Remove a notebook (destructive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/removeNotebook", map[string]any{"notebook": args[0]})
		},
	}

	var newName string
	rename := &cobra.Command{
		Use:   "rename <notebook-id> --name <name>",
		Short: "Rename a notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/renameNotebook", map[string]any{"notebook": args[0], "name": newName})
		},
	}
	rename.Flags().StringVar(&newName, "name", "", "new notebook name (required)")
	_ = rename.MarkFlagRequired("name")

	info := &cobra.Command{
		Use:   "info <notebook-id>",
		Short: "Get notebook info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/getNotebookInfo", map[string]any{"notebook": args[0]})
		},
	}

	confGet := &cobra.Command{
		Use:   "conf-get <notebook-id>",
		Short: "Get notebook configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/notebook/getNotebookConf", map[string]any{"notebook": args[0]})
		},
	}

	var confData, confFile string
	confSet := &cobra.Command{
		Use:   "conf-set <notebook-id> --data '{...}'",
		Short: "Set notebook configuration (conf object)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := readJSONObject(confData, confFile)
			if err != nil {
				return err
			}
			return emit(cmd, "/api/notebook/setNotebookConf", map[string]any{"notebook": args[0], "conf": conf})
		},
	}
	confSet.Flags().StringVarP(&confData, "data", "d", "", "conf JSON object")
	confSet.Flags().StringVar(&confFile, "data-file", "", "read conf JSON from a file ('-' for stdin)")

	cmd.AddCommand(ls, open, closeCmd, create, remove, rename, info, confGet, confSet)
	return cmd
}
