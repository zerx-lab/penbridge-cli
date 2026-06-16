// Package cli wires up the penbridge command tree.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/client"
)

// Version is the CLI version (overridable at build time via -ldflags).
var Version = "0.1.0"

var (
	apiClient    *client.Client
	outputFormat string

	flagBaseURL  string
	flagToken    string
	flagUser     string
	flagPassword string
	flagConfig   string
	flagTimeout  int
	flagInsecure bool
	flagDryRun   bool
	flagVerbose  bool

	resolvedConfigPath string
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "penbridge",
		Short: "PenBridge - a CLI bridge to the SiYuan kernel API",
		Long: `PenBridge is a command-line bridge to the SiYuan (思源笔记) kernel API.

It exposes every open API endpoint so you (or an AI agent) can read and edit any
notebook, document, block, or paragraph in a safe, scriptable way.

Authentication:
  The kernel API (default ` + client.DefaultBaseURL + `) uses an API token found in
  SiYuan: Settings -> About -> API token. Provide it with --token, the
  PENBRIDGE_TOKEN env var, or 'penbridge config set --token <token>'.

  The publish service (default :6808) is READ-ONLY and uses Basic auth
  (--user/--password); write commands will fail against it.

Discovery:
  'penbridge api list' prints every known endpoint. 'penbridge api <path>' calls
  any endpoint directly. Typed subcommands (block, doc, notebook, ...) wrap the
  most common operations.

Safety:
  Use --dry-run to print a request without sending it. Use --verbose to log
  requests/responses to stderr.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&flagBaseURL, "base-url", "", "kernel API base URL (default "+client.DefaultBaseURL+")")
	pf.StringVar(&flagToken, "token", "", "API token (Settings -> About -> API token)")
	pf.StringVar(&flagUser, "user", "", "HTTP Basic auth username (publish service / proxy)")
	pf.StringVar(&flagPassword, "password", "", "HTTP Basic auth password")
	pf.StringVar(&flagConfig, "config", "", "config file path (default OS config dir/penbridge/config.json)")
	pf.IntVar(&flagTimeout, "timeout", 0, "request timeout in seconds (default 120)")
	pf.BoolVar(&flagInsecure, "insecure", false, "skip TLS certificate verification")
	pf.StringVarP(&outputFormat, "output", "o", "pretty", "output format: pretty|json|raw|envelope")
	pf.BoolVar(&flagDryRun, "dry-run", false, "print the request instead of sending it")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "log requests/responses to stderr")

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return setupClient(cmd)
	}

	root.AddCommand(
		newAPICmd(),
		newNotebookCmd(),
		newDocCmd(),
		newBlockCmd(),
		newAttrCmd(),
		newSQLCmd(),
		newFlushCmd(),
		newFileCmd(),
		newSearchCmd(),
		newExportCmd(),
		newTemplateCmd(),
		newNotifyCmd(),
		newSystemCmd(),
		newProxyCmd(),
		newTxCmd(),
		newConfigCmd(),
	)
	return root
}

func setupClient(cmd *cobra.Command) error {
	path := flagConfig
	if path == "" {
		if p, err := client.DefaultConfigPath(); err == nil {
			path = p
		}
	}
	resolvedConfigPath = path

	cfg, err := client.LoadConfig(path)
	if err != nil {
		return err
	}
	flags := cmd.Flags()
	if flags.Changed("base-url") {
		cfg.BaseURL = flagBaseURL
	}
	if flags.Changed("token") {
		cfg.Token = flagToken
	}
	if flags.Changed("user") {
		cfg.User = flagUser
	}
	if flags.Changed("password") {
		cfg.Password = flagPassword
	}
	if flags.Changed("timeout") {
		cfg.Timeout = flagTimeout
	}
	if flags.Changed("insecure") {
		cfg.Insecure = flagInsecure
	}

	apiClient = client.New(cfg)
	apiClient.DryRun = flagDryRun
	apiClient.Verbose = flagVerbose
	return nil
}

// Execute runs the root command and exits with a non-zero status on error.
func Execute() {
	err := newRootCmd().Execute()
	if err == nil || errors.Is(err, client.ErrDryRun) {
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
