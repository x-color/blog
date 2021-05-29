package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/x-color/blog/tools/blog/blog"
)

func runBuildCmd(cmd *cobra.Command, args []string) error {
	title := args[0]

	configPath := fmt.Sprintf("config/zenn/%s.yaml", title)
	content, err := blog.BuildZennArticle(configPath)
	if err != nil {
		return err
	}

	articlePath := fmt.Sprintf("articles/%s.md", title)
	err = os.WriteFile(articlePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return err
}

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <title>",
		Short: "Build an article to articles directory for Zenn",
		Args:  cobra.ExactArgs(1),
		RunE:  runBuildCmd,
	}

	return cmd
}
