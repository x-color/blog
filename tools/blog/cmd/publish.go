package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/x-color/blog/tools/blog/blog"
)

func runPublishCmd(cmd *cobra.Command, args []string) error {
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	postFiles, err := filepath.Glob(filepath.Join(cd, "content/posts", "*"))
	if err != nil {
		return err
	}

	return blog.Publish(postFiles)
}

func newPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Update published parameter to 'true' if 'date' is passed",
		Args:  cobra.NoArgs,
		RunE:  runPublishCmd,
	}

	return cmd
}
