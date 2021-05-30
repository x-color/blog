package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/x-color/blog/tools/blog/blog"
)

func runPostCmd(cmd *cobra.Command, args []string) error {
	title := args[0]

	configPath := fmt.Sprintf("config/qiita/%s.yaml", title)
	content, err := blog.BuildQiitaArticle(configPath)
	if err != nil {
		return err
	}

	if !content.Edited || !content.Private {
		cmd.Println("Skip")
		return nil
	}

	id, err := blog.PostArticleToQiita(content, os.Getenv("TOKEN"))
	if err != nil {
		return err
	}

	return blog.UpdateQiitaArticleConf(configPath, id, content.Hash)
}

func newPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post <title>",
		Short: "Post an article to Qiita",
		Args:  cobra.ExactArgs(1),
		RunE:  runPostCmd,
	}

	return cmd
}
