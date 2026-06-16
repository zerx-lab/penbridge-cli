package cli

import "github.com/spf13/cobra"

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search blocks and assets, or find/replace text",
	}
	cmd.AddCommand(
		newSearchFullTextCmd(),
		newSearchAssetCmd(),
		newSearchReplaceCmd(),
	)
	return cmd
}

func newSearchFullTextCmd() *cobra.Command {
	var (
		page     int
		pageSize int
		method   int
		orderBy  int
		groupBy  int
		paths    []string
	)
	cmd := &cobra.Command{
		Use:   "fulltext <query>",
		Short: "Full-text search blocks",
		Long: `Full-text search for blocks.

--method: 0=keyword, 1=query syntax, 2=SQL (admin), 3=regex
--order-by: 0..7 (see SiYuan docs)  --group-by: 0=none, 1=by document`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{
				"query":    args[0],
				"page":     page,
				"pageSize": pageSize,
				"method":   method,
				"orderBy":  orderBy,
				"groupBy":  groupBy,
			}
			if len(paths) > 0 {
				body["paths"] = paths
			}
			return emit(cmd, "/api/search/fullTextSearchBlock", body)
		},
	}
	f := cmd.Flags()
	f.IntVar(&page, "page", 1, "page number")
	f.IntVar(&pageSize, "page-size", 32, "results per page")
	f.IntVar(&method, "method", 0, "0=keyword,1=query,2=SQL,3=regex")
	f.IntVar(&orderBy, "order-by", 0, "ordering mode")
	f.IntVar(&groupBy, "group-by", 0, "0=none,1=by document")
	f.StringSliceVar(&paths, "paths", nil, "restrict to these notebook/.sy paths")
	return cmd
}

func newSearchAssetCmd() *cobra.Command {
	var exts []string
	cmd := &cobra.Command{
		Use:   "asset <keyword>",
		Short: "Search assets by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"k": args[0]}
			if len(exts) > 0 {
				body["exts"] = exts
			}
			return emit(cmd, "/api/search/searchAsset", body)
		},
	}
	cmd.Flags().StringSliceVar(&exts, "exts", nil, "filter by extensions (e.g. .pdf,.png)")
	return cmd
}

func newSearchReplaceCmd() *cobra.Command {
	var (
		find    string
		replace string
		ids     []string
	)
	cmd := &cobra.Command{
		Use:   "replace --find <text> --replace <text> --ids a,b",
		Short: "Find and replace text within specific blocks (destructive)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return emit(cmd, "/api/search/findReplace", map[string]any{
				"k": find, "r": replace, "ids": ids,
			})
		},
	}
	f := cmd.Flags()
	f.StringVar(&find, "find", "", "text to find (required)")
	f.StringVar(&replace, "replace", "", "replacement text")
	f.StringSliceVar(&ids, "ids", nil, "block IDs to operate on (required)")
	_ = cmd.MarkFlagRequired("find")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}
