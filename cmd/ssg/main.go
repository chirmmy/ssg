package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chirmmy/ssg/internal/build"
	"github.com/chirmmy/ssg/internal/config"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "ssg",
		Short: "ssg is a static site generator",
		Long:  `ssg is a static site generator that generates static HTML files from markdown files.`,
	}

	var cfgPath string
	var root string

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site",
		Long:  `Build the static site from markdown files.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			b := build.NewBuilder(cfg, root)

			return b.Build(context.Background())
		},
	}
	buildCmd.Flags().StringVar(&cfgPath, "config", "site.toml", "Path to the configuration file")
	buildCmd.Flags().StringVar(&root, "root", ".", "root path")

	rootCmd.AddCommand(buildCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
