package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/catalog"
)

func newAPICmd() *cobra.Command {
	var (
		data     string
		dataFile string
		params   []string
		method   string
		rawBody  bool
	)
	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Call any kernel API endpoint with a JSON body",
		Long: `Call any SiYuan kernel API endpoint directly.

The <path> may be given with or without the /api/ prefix:
  penbridge api block/insertBlock -d '{"parentID":"...","dataType":"markdown","data":"hi"}'
  penbridge api /api/system/version

Body sources are merged (low to high precedence): --data-file, --data, --param.
With no body the request sends '{}'. Use 'penbridge api list' to discover paths.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := buildBody(data, dataFile, params, rawBody)
			if err != nil {
				return err
			}
			return emitResp(apiClient.CallMethod(cmd.Context(), method, args[0], body))
		},
	}
	f := cmd.Flags()
	f.StringVarP(&data, "data", "d", "", "JSON body string")
	f.StringVar(&dataFile, "data-file", "", "read JSON body from a file ('-' for stdin)")
	f.StringArrayVarP(&params, "param", "p", nil, "body field key=value (repeatable; value parsed as JSON when valid)")
	f.StringVar(&method, "method", "POST", "HTTP method")
	f.BoolVar(&rawBody, "raw-body", false, "send --data/--data-file verbatim (no JSON parsing/merging)")

	cmd.AddCommand(newAPIListCmd())
	return cmd
}

func newAPIListCmd() *cobra.Command {
	var (
		search     string
		category   string
		writesOnly bool
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known kernel API endpoints (for discovery)",
		Long: `List the embedded catalog of SiYuan kernel API endpoints.

Flags shown after each path: W=write/mutating, A=admin role, D=deprecated.
Use --json for a machine-readable list (handy for agents).`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			eps := catalog.Filter(search, category, writesOnly)
			if asJSON {
				return printValue(eps, true)
			}
			if len(eps) == 0 {
				fmt.Fprintln(os.Stderr, "no endpoints match the given filters")
				return nil
			}
			byCat := map[string][]catalog.Endpoint{}
			for _, e := range eps {
				byCat[e.Category] = append(byCat[e.Category], e)
			}
			cats := make([]string, 0, len(byCat))
			for c := range byCat {
				cats = append(cats, c)
			}
			sort.Strings(cats)
			for _, c := range cats {
				fmt.Printf("# %s\n", c)
				for _, e := range byCat[c] {
					fmt.Printf("  %-50s %s\n", e.Path, endpointFlags(e))
				}
				fmt.Println()
			}
			fmt.Printf("%d endpoint(s). Flags: W=write A=admin D=deprecated\n", len(eps))
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&search, "search", "s", "", "filter by substring of path/name")
	f.StringVarP(&category, "category", "c", "", "filter by category (e.g. block, notebook, filetree)")
	f.BoolVar(&writesOnly, "writes", false, "only show write/mutating endpoints")
	f.BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func endpointFlags(e catalog.Endpoint) string {
	var b strings.Builder
	write := func(ok bool, ch string) {
		if ok {
			b.WriteString(ch)
		} else {
			b.WriteString("-")
		}
	}
	write(e.Write, "W")
	write(e.Admin, "A")
	write(e.Deprecated, "D")
	return b.String()
}
