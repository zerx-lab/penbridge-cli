package cli

import "github.com/spf13/cobra"

func newNotifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Push a notification message into the SiYuan UI",
	}

	newPush := func(use, short, path string) *cobra.Command {
		var timeout int
		c := &cobra.Command{
			Use:   use,
			Short: short,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return emit(cmd, path, map[string]any{"msg": args[0], "timeout": timeout})
			},
		}
		c.Flags().IntVar(&timeout, "timeout", 7000, "display duration in milliseconds")
		return c
	}

	cmd.AddCommand(
		newPush("msg <text>", "Push an info message", "/api/notification/pushMsg"),
		newPush("err <text>", "Push an error message", "/api/notification/pushErrMsg"),
	)
	return cmd
}
