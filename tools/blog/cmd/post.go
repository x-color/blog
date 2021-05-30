package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/x-color/blog/tools/blog/blog"
)

func runPostCmd(cmd *cobra.Command, args []string) error {
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	configFiles, err := filepath.Glob(filepath.Join(cd, "config/qiita", "*"))
	if err != nil {
		return err
	}

	for _, configPath := range configFiles {
		content, err := blog.BuildQiitaArticle(configPath)
		if err != nil {
			return err
		}

		if !content.Edited || !content.Private {
			continue
		}

		id, err := blog.PostArticleToQiita(content, os.Getenv("TOKEN"))
		if err != nil {
			return err
		}
		cmd.Println(strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath)))

		if err = blog.UpdateQiitaArticleConf(configPath, id, content.Hash); err != nil {
			return err
		}

	}
	return nil
}

func newPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post articles to Qiita",
		Args:  cobra.NoArgs,
		RunE:  runPostCmd,
	}

	return cmd
}
