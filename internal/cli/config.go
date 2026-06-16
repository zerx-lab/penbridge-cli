package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/client"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage saved connection settings (base URL, token, ...)",
	}
	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigSetCmd(),
		newConfigPathCmd(),
		newConfigTestCmd(),
	)
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the effective configuration (secrets masked)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return printValue(apiClient.Config().Redacted(), true)
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(resolvedConfigPath)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	var (
		baseURL  string
		token    string
		user     string
		password string
		timeout  int
		insecure bool
	)
	cmd := &cobra.Command{
		Use:   "set [--base-url ...] [--token ...] [--user ...] [--password ...] [--timeout ...] [--insecure]",
		Short: "Persist connection settings to the config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Start from whatever is currently saved on disk (not env/flag overrides).
			cfg, err := client.LoadConfig(resolvedConfigPath)
			if err != nil {
				return err
			}
			f := cmd.Flags()
			if f.Changed("base-url") {
				cfg.BaseURL = baseURL
			}
			if f.Changed("token") {
				cfg.Token = token
			}
			if f.Changed("user") {
				cfg.User = user
			}
			if f.Changed("password") {
				cfg.Password = password
			}
			if f.Changed("timeout") {
				cfg.Timeout = timeout
			}
			if f.Changed("insecure") {
				cfg.Insecure = insecure
			}
			if err := client.SaveConfig(resolvedConfigPath, cfg); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "saved config to %s\n", resolvedConfigPath)
			return printValue(cfg.Redacted(), true)
		},
	}
	f := cmd.Flags()
	// Local flags (shadow the persistent ones) so `config set --token x` is intuitive.
	f.StringVar(&baseURL, "base-url", "", "kernel API base URL")
	f.StringVar(&token, "token", "", "API token")
	f.StringVar(&user, "user", "", "Basic auth username")
	f.StringVar(&password, "password", "", "Basic auth password")
	f.IntVar(&timeout, "timeout", 0, "request timeout in seconds")
	f.BoolVar(&insecure, "insecure", false, "skip TLS verification")
	return cmd
}

func newConfigTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Verify connectivity and auth by calling system/version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := apiClient.Call(cmd.Context(), "/api/system/version", nil)
			if err != nil {
				return err
			}
			var ver string
			_ = jsonUnmarshalString(resp.Data, &ver)
			fmt.Fprintf(os.Stderr, "OK: connected to %s (kernel version %s)\n", apiClient.Config().BaseURL, ver)
			return nil
		},
	}
}
