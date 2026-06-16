package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Read and edit blocks/paragraphs (insert/append/prepend/update/delete/move/fold/get)",
	}
	cmd.AddCommand(
		newBlockInsertCmd(),
		newBlockAppendCmd(),
		newBlockPrependCmd(),
		newBlockUpdateCmd(),
		newBlockDeleteCmd(),
		newBlockMoveCmd(),
		newBlockFoldCmd(true),
		newBlockFoldCmd(false),
		newBlockGetCmd(),       // kramdown (source)
		newBlockInfoCmd(),
		newBlockDOMCmd(),
		newBlockChildrenCmd(),
		newBlockExistCmd(),
		newBlockBreadcrumbCmd(),
		newBlockSiblingCmd(),
		newBlockRefsCmd(),
	)
	return cmd
}

func newBlockInsertCmd() *cobra.Command {
	var (
		cf         contentFlags
		parentID   string
		previousID string
		nextID     string
	)
	cmd := &cobra.Command{
		Use:   "insert [--previous-id <id> | --parent-id <id> | --next-id <id>] (--markdown ... | --dom ... | --content-file ...)",
		Short: "Insert a new block relative to another block",
		Long: `Insert a new block.

Position is chosen by exactly one of --previous-id (insert after it),
--next-id (insert before it), or --parent-id (insert as first child).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data, dataType, err := cf.resolve(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{"data": data, "dataType": dataType}
			optString(body, "parentID", parentID)
			optString(body, "previousID", previousID)
			optString(body, "nextID", nextID)
			return emit(cmd, "/api/block/insertBlock", body)
		},
	}
	cf.register(cmd.Flags())
	f := cmd.Flags()
	f.StringVar(&parentID, "parent-id", "", "insert as the first child of this block")
	f.StringVar(&previousID, "previous-id", "", "insert after this block")
	f.StringVar(&nextID, "next-id", "", "insert before this block")
	return cmd
}

func newBlockAppendCmd() *cobra.Command {
	var cf contentFlags
	cmd := &cobra.Command{
		Use:   "append <parent-id> (--markdown ... | --dom ... | --content-file ...)",
		Short: "Append a block as the last child of a parent (e.g. a document)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, dataType, err := cf.resolve(cmd)
			if err != nil {
				return err
			}
			return emit(cmd, "/api/block/appendBlock", map[string]any{
				"parentID": args[0], "data": data, "dataType": dataType,
			})
		},
	}
	cf.register(cmd.Flags())
	return cmd
}

func newBlockPrependCmd() *cobra.Command {
	var cf contentFlags
	cmd := &cobra.Command{
		Use:   "prepend <parent-id> (--markdown ... | --dom ... | --content-file ...)",
		Short: "Prepend a block as the first child of a parent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, dataType, err := cf.resolve(cmd)
			if err != nil {
				return err
			}
			return emit(cmd, "/api/block/prependBlock", map[string]any{
				"parentID": args[0], "data": data, "dataType": dataType,
			})
		},
	}
	cf.register(cmd.Flags())
	return cmd
}

func newBlockUpdateCmd() *cobra.Command {
	var cf contentFlags
	cmd := &cobra.Command{
		Use:   "update <id> (--markdown ... | --dom ... | --content-file ...)",
		Short: "Update (replace) a block's content",
		Long: `Replace the content of the block with the given id.

This rewrites the whole block, so include all of its content. To edit large or
multi-line content safely, use --content-file - and pipe Markdown via stdin.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, dataType, err := cf.resolve(cmd)
			if err != nil {
				return err
			}
			return emit(cmd, "/api/block/updateBlock", map[string]any{
				"id": args[0], "data": data, "dataType": dataType,
			})
		},
	}
	cf.register(cmd.Flags())
	return cmd
}

func newBlockDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a block (destructive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/deleteBlock", map[string]any{"id": args[0]})
		},
	}
}

func newBlockMoveCmd() *cobra.Command {
	var (
		parentID   string
		previousID string
	)
	cmd := &cobra.Command{
		Use:   "move <id> [--parent-id <id>] [--previous-id <id>]",
		Short: "Move a block under a parent and/or after a sibling",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if parentID == "" && previousID == "" {
				return fmt.Errorf("provide --parent-id and/or --previous-id")
			}
			body := map[string]any{"id": args[0]}
			optString(body, "parentID", parentID)
			optString(body, "previousID", previousID)
			return emit(cmd, "/api/block/moveBlock", body)
		},
	}
	f := cmd.Flags()
	f.StringVar(&parentID, "parent-id", "", "new parent block ID")
	f.StringVar(&previousID, "previous-id", "", "move after this sibling block ID")
	return cmd
}

func newBlockFoldCmd(fold bool) *cobra.Command {
	verb := "unfold"
	path := "/api/block/unfoldBlock"
	if fold {
		verb = "fold"
		path = "/api/block/foldBlock"
	}
	return &cobra.Command{
		Use:   verb + " <id>",
		Short: verb + " a block (e.g. a heading or list)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, path, map[string]any{"id": args[0]})
		},
	}
}

func newBlockGetCmd() *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a block's Markdown/Kramdown source (read before editing)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"id": args[0]}
			optString(body, "mode", mode)
			return emit(cmd, "/api/block/getBlockKramdown", body)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "kramdown mode: md or textmark (default md)")
	return cmd
}

func newBlockInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <id>",
		Short: "Get block info (box, path, rootID, rootTitle, ...)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getBlockInfo", map[string]any{"id": args[0]})
		},
	}
}

func newBlockDOMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dom <id>",
		Short: "Get a block's rendered DOM/HTML",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getBlockDOM", map[string]any{"id": args[0]})
		},
	}
}

func newBlockChildrenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "children <id>",
		Short: "List the direct child blocks of a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getChildBlocks", map[string]any{"id": args[0]})
		},
	}
}

func newBlockExistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exist <id>",
		Short: "Check whether a block exists (returns true/false)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/checkBlockExist", map[string]any{"id": args[0]})
		},
	}
}

func newBlockBreadcrumbCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "breadcrumb <id>",
		Short: "Get the breadcrumb path of a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getBlockBreadcrumb", map[string]any{"id": args[0]})
		},
	}
}

func newBlockSiblingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sibling <id>",
		Short: "Get a block's parent/previous/next sibling IDs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getBlockSiblingID", map[string]any{"id": args[0]})
		},
	}
}

func newBlockRefsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refs <id>",
		Short: "Get the reference (backlink) definition IDs for a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/block/getRefIDs", map[string]any{"id": args[0]})
		},
	}
}
