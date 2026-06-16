package cli

import "github.com/spf13/cobra"

func newAttrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attr",
		Aliases: []string{"attrs"},
		Short:   "Get and set block attributes",
	}

	get := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a block's attributes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/attr/getBlockAttrs", map[string]any{"id": args[0]})
		},
	}

	var (
		pairs  []string
		remove []string
		data   string
		file   string
	)
	set := &cobra.Command{
		Use:   "set <id> (--attr k=v ... | --remove k ... | --data '{...}')",
		Short: "Set block attributes (custom attrs start with 'custom-')",
		Long: `Set attributes on a block.

Use --attr key=value (repeatable) to set values, --remove key to delete them,
or --data with a full JSON object (a null value deletes that attribute).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var attrs map[string]any
			var err error
			if data != "" || file != "" {
				attrs, err = readJSONObject(data, file)
			} else {
				attrs, err = attrsFromPairs(pairs, remove)
			}
			if err != nil {
				return err
			}
			return emit(cmd, "/api/attr/setBlockAttrs", map[string]any{"id": args[0], "attrs": attrs})
		},
	}
	sf := set.Flags()
	sf.StringArrayVar(&pairs, "attr", nil, "attribute key=value (repeatable)")
	sf.StringArrayVar(&remove, "remove", nil, "attribute key to delete (repeatable)")
	sf.StringVarP(&data, "data", "d", "", "attributes as a JSON object")
	sf.StringVar(&file, "data-file", "", "read attributes JSON from a file ('-' for stdin)")

	var ids []string
	batchGet := &cobra.Command{
		Use:   "batch-get --ids a,b,c",
		Short: "Get attributes for multiple blocks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/attr/batchGetBlockAttrs", map[string]any{"ids": ids})
		},
	}
	batchGet.Flags().StringSliceVar(&ids, "ids", nil, "block IDs (comma-separated)")
	_ = batchGet.MarkFlagRequired("ids")

	cmd.AddCommand(get, set, batchGet)
	return cmd
}
