package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zerx-lab/penbridge-cli/internal/client"
)

func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Read/write workspace files and assets (get/put/remove/rename/readdir/copy)",
	}
	cmd.AddCommand(
		newFileGetCmd(),
		newFilePutCmd(),
		newFilePutDirCmd(),
		newFileRemoveCmd(),
		newFileRenameCmd(),
		newFileReadDirCmd(),
		newFileCopyCmd(),
	)
	return cmd
}

func newFileGetCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "get <workspace-path> [--out <local-file>]",
		Short: "Download a workspace file (raw bytes to stdout or --out)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, _, err := apiClient.GetFile(cmd.Context(), args[0])
			if errors.Is(err, client.ErrDryRun) {
				return nil
			}
			if err != nil {
				return err
			}
			if out != "" {
				if err := os.WriteFile(out, data, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(data), out)
				return nil
			}
			_, err = os.Stdout.Write(data)
			return err
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write to this local file instead of stdout")
	return cmd
}

func newFilePutCmd() *cobra.Command {
	var (
		local       string
		contentFile string
	)
	cmd := &cobra.Command{
		Use:   "put <workspace-path> (--file <local-file> | --content-file -)",
		Short: "Upload a file to a workspace path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsPath := args[0]
			var (
				name string
				err  error
			)
			switch {
			case local != "":
				f, e := os.Open(local)
				if e != nil {
					return e
				}
				defer f.Close()
				name = filepath.Base(local)
				return emitResp(apiClient.PutFile(cmd.Context(), wsPath, false, name, f))
			case contentFile != "":
				b, e := readInput(contentFile)
				if e != nil {
					return e
				}
				name = filepath.Base(strings.TrimLeft(wsPath, "/"))
				return emitResp(apiClient.PutFile(cmd.Context(), wsPath, false, name, strings.NewReader(string(b))))
			default:
				err = fmt.Errorf("provide --file <local-file> or --content-file -")
			}
			return err
		},
	}
	f := cmd.Flags()
	f.StringVar(&local, "file", "", "local file to upload")
	f.StringVar(&contentFile, "content-file", "", "read content from a file ('-' for stdin)")
	return cmd
}

func newFilePutDirCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "put-dir <workspace-path>",
		Short: "Create a directory in the workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emitResp(apiClient.PutFile(cmd.Context(), args[0], true, "", nil))
		},
	}
}

func newFileRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <workspace-path>",
		Short: "Remove a workspace file or directory (destructive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/file/removeFile", map[string]any{"path": args[0]})
		},
	}
}

func newFileRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <workspace-path> <new-path>",
		Short: "Rename/move a workspace file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/file/renameFile", map[string]any{"path": args[0], "newPath": args[1]})
		},
	}
}

func newFileReadDirCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "readdir <workspace-path>",
		Short: "List entries in a workspace directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/file/readDir", map[string]any{"path": args[0]})
		},
	}
}

func newFileCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy <src> <dest>",
		Short: "Copy a file (src is an asset path; dest is an absolute path)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return emit(cmd, "/api/file/copyFile", map[string]any{"src": args[0], "dest": args[1]})
		},
	}
}
