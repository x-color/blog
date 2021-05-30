package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/x-color/blog/tools/blog/blog"
)

func runBuildCmd(cmd *cobra.Command, args []string) error {
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	configFiles, err := filepath.Glob(filepath.Join(cd, "config/zenn", "*"))
	if err != nil {
		return err
	}

	for _, configPath := range configFiles {
		content, err := blog.BuildZennArticle(configPath)
		if err != nil {
			return err
		}

		title := strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath))
		articlePath := fmt.Sprintf("articles/%s.md", title)
		err = os.WriteFile(articlePath, []byte(content), 0644)
		if err != nil {
			return err
		}
		cmd.Println(title)
	}

	return nil
}

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build articles to articles directory for Zenn",
		Args:  cobra.NoArgs,
		RunE:  runBuildCmd,
	}

	return cmd
}
