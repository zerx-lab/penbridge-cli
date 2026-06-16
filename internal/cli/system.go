package cli

import "github.com/spf13/cobra"

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System info (version, time, boot progress, config)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Get the SiYuan kernel version",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return emit(cmd, "/api/system/version", nil)
			},
		},
		&cobra.Command{
			Use:   "time",
			Short: "Get the current server time (ms)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return emit(cmd, "/api/system/currentTime", nil)
			},
		},
		&cobra.Command{
			Use:   "boot",
			Short: "Get boot progress",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return emit(cmd, "/api/system/bootProgress", nil)
			},
		},
		&cobra.Command{
			Use:   "conf",
			Short: "Get the workspace configuration",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return emit(cmd, "/api/system/getConf", nil)
			},
		},
	)
	return cmd
}
