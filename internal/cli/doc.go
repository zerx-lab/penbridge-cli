package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Manage documents (create/rename/remove/move/get/list/tree/path)",
	}
	cmd.AddCommand(
		newDocCreateCmd(),
		newDocCreateDailyCmd(),
		newDocRenameCmd(),
		newDocRemoveCmd(),
		newDocMoveCmd(),
		newDocGetCmd(),
		newDocListCmd(),
		newDocTreeCmd(),
		newDocDuplicateCmd(),
		newDocHPathCmd(),
		newDocPathCmd(),
		newDocIDsCmd(),
	)
	return cmd
}

func newDocCreateCmd() *cobra.Command {
	var (
		notebook    string
		hpath       string
		markdown    string
		contentFile string
		id          string
		parentID    string
		tags        string
		withMath    bool
	)
	cmd := &cobra.Command{
		Use:   "create --notebook <id> --path <hpath> [--markdown ... | --content-file ...]",
		Short: "Create a document from Markdown (returns the new doc ID)",
		Long: `Create a document from Markdown content.

--path is the human-readable path (hPath), e.g. /Projects/Notes; the last
segment becomes the document title. Provide content with --markdown or
--content-file ('-' for stdin), which is best for long, multi-line content.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			md := markdown
			if contentFile != "" {
				b, err := readInput(contentFile)
				if err != nil {
					return err
				}
				md = string(b)
			}
			body := map[string]any{"notebook": notebook, "path": hpath, "markdown": md}
			optString(body, "id", id)
			optString(body, "parentID", parentID)
			optString(body, "tags", tags)
			if withMath {
				body["withMath"] = true
			}
			return emit(cmd, "/api/filetree/createDocWithMd", body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&notebook, "notebook", "", "notebook ID (required)")
	f.StringVar(&hpath, "path", "", "human-readable path, e.g. /Folder/Title (required)")
	f.StringVar(&markdown, "markdown", "", "document content as Markdown")
	f.StringVar(&contentFile, "content-file", "", "read Markdown from a file ('-' for stdin)")
	f.StringVar(&id, "id", "", "predefined document root block ID (optional)")
	f.StringVar(&parentID, "parent-id", "", "parent document ID (optional)")
	f.StringVar(&tags, "tags", "", "document tags (optional)")
	f.BoolVar(&withMath, "with-math", false, "parse inline math")
	_ = cmd.MarkFlagRequired("notebook")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func newDocCreateDailyCmd() *cobra.Command {
	var notebook string
	cmd := &cobra.Command{
		Use:   "create-daily --notebook <id>",
		Short: "Create (or open) today's daily note (returns its ID)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/filetree/createDailyNote", map[string]any{"notebook": notebook})
		},
	}
	cmd.Flags().StringVar(&notebook, "notebook", "", "notebook ID (required)")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}

func newDocRenameCmd() *cobra.Command {
	var (
		id       string
		notebook string
		path     string
		title    string
	)
	cmd := &cobra.Command{
		Use:   "rename (--id <id> | --notebook <id> --path <.sy path>) --title <title>",
		Short: "Rename a document by ID or by notebook+path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if id != "" {
				return emit(cmd, "/api/filetree/renameDocByID", map[string]any{"id": id, "title": title})
			}
			if notebook == "" || path == "" {
				return fmt.Errorf("provide either --id, or both --notebook and --path")
			}
			return emit(cmd, "/api/filetree/renameDoc", map[string]any{"notebook": notebook, "path": path, "title": title})
		},
	}
	f := cmd.Flags()
	f.StringVar(&id, "id", "", "document root block ID")
	f.StringVar(&notebook, "notebook", "", "notebook ID (with --path)")
	f.StringVar(&path, "path", "", "document .sy storage path (with --notebook)")
	f.StringVar(&title, "title", "", "new title (required)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func newDocRemoveCmd() *cobra.Command {
	var (
		id       string
		notebook string
		path     string
	)
	cmd := &cobra.Command{
		Use:   "remove (--id <id> | --notebook <id> --path <.sy path>)",
		Short: "Remove a document by ID or by notebook+path (destructive)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if id != "" {
				return emit(cmd, "/api/filetree/removeDocByID", map[string]any{"id": id})
			}
			if notebook == "" || path == "" {
				return fmt.Errorf("provide either --id, or both --notebook and --path")
			}
			return emit(cmd, "/api/filetree/removeDoc", map[string]any{"notebook": notebook, "path": path})
		},
	}
	f := cmd.Flags()
	f.StringVar(&id, "id", "", "document root block ID")
	f.StringVar(&notebook, "notebook", "", "notebook ID (with --path)")
	f.StringVar(&path, "path", "", "document .sy storage path (with --notebook)")
	return cmd
}

func newDocMoveCmd() *cobra.Command {
	var (
		fromIDs    []string
		toID       string
		fromPaths  []string
		toNotebook string
		toPath     string
	)
	cmd := &cobra.Command{
		Use:   "move (--from-ids a,b --to-id <id>) | (--from-paths ... --to-notebook <id> --to-path <.sy path>)",
		Short: "Move documents by ID or by path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(fromIDs) > 0 {
				if toID == "" {
					return fmt.Errorf("--to-id is required with --from-ids")
				}
				return emit(cmd, "/api/filetree/moveDocsByID", map[string]any{"fromIDs": fromIDs, "toID": toID})
			}
			if len(fromPaths) == 0 || toNotebook == "" || toPath == "" {
				return fmt.Errorf("provide --from-ids+--to-id, or --from-paths+--to-notebook+--to-path")
			}
			return emit(cmd, "/api/filetree/moveDocs", map[string]any{
				"fromPaths": fromPaths, "toNotebook": toNotebook, "toPath": toPath,
			})
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&fromIDs, "from-ids", nil, "source document IDs (comma-separated)")
	f.StringVar(&toID, "to-id", "", "target document or notebook ID")
	f.StringSliceVar(&fromPaths, "from-paths", nil, "source .sy paths (comma-separated)")
	f.StringVar(&toNotebook, "to-notebook", "", "target notebook ID")
	f.StringVar(&toPath, "to-path", "", "target .sy path")
	return cmd
}

func newDocGetCmd() *cobra.Command {
	var (
		mode int
		size int
	)
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a document's content (rendered block DOM)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"id": args[0], "mode": mode, "size": size}
			return emit(cmd, "/api/filetree/getDoc", body)
		},
	}
	f := cmd.Flags()
	f.IntVar(&mode, "mode", 0, "0=root only, 1=up, 2=down, 3=both, 4=tail")
	f.IntVar(&size, "size", 102400, "max blocks to load")
	return cmd
}

func newDocListCmd() *cobra.Command {
	var (
		notebook string
		path     string
		sort     int
	)
	cmd := &cobra.Command{
		Use:   "list --notebook <id> [--path /]",
		Short: "List documents under a path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body := map[string]any{"notebook": notebook, "path": path}
			if cmd.Flags().Changed("sort") {
				body["sort"] = sort
			}
			return emit(cmd, "/api/filetree/listDocsByPath", body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&notebook, "notebook", "", "notebook ID (required)")
	f.StringVar(&path, "path", "/", "parent .sy path (default /)")
	f.IntVar(&sort, "sort", 0, "sort mode")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}

func newDocTreeCmd() *cobra.Command {
	var (
		notebook string
		path     string
	)
	cmd := &cobra.Command{
		Use:   "tree --notebook <id> [--path /]",
		Short: "List the document tree (IDs and children) under a path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/filetree/listDocTree", map[string]any{"notebook": notebook, "path": path})
		},
	}
	f := cmd.Flags()
	f.StringVar(&notebook, "notebook", "", "notebook ID (required)")
	f.StringVar(&path, "path", "/", "parent .sy path (default /)")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}

func newDocDuplicateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "duplicate <id>",
		Short: "Duplicate a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/filetree/duplicateDoc", map[string]any{"id": args[0]})
		},
	}
}

func newDocHPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hpath <id>",
		Short: "Get the human-readable path for a block/document ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/filetree/getHPathByID", map[string]any{"id": args[0]})
		},
	}
}

func newDocPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path <id>",
		Short: "Get the notebook and .sy storage path for a document ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/filetree/getPathByID", map[string]any{"id": args[0]})
		},
	}
}

func newDocIDsCmd() *cobra.Command {
	var notebook string
	cmd := &cobra.Command{
		Use:   "ids --notebook <id> <hpath>",
		Short: "Get document IDs matching a human-readable path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/filetree/getIDsByHPath", map[string]any{"notebook": notebook, "path": args[0]})
		},
	}
	cmd.Flags().StringVar(&notebook, "notebook", "", "notebook ID (required)")
	_ = cmd.MarkFlagRequired("notebook")
	return cmd
}
