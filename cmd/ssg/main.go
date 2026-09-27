package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/chirmmy/ssg/internal/build"
	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/dev"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "ssg",
		Short: "ssg is a static site generator",
		Long:  `ssg is a static site generator that generates static HTML files from markdown files.`,
	}

	// go run ./cmd/ssg build
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
			b := build.NewBuilder(cfg, root, false)

			return b.Build(context.Background())
		},
	}
	buildCmd.Flags().StringVar(&cfgPath, "config", "site.toml", "Path to the configuration file")
	buildCmd.Flags().StringVar(&root, "root", ".", "root path")

	// go run ./cmd/ssg dev
	var (
		devPort int
		devHost string
		devOpen bool
	)

	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Start the dev server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgPath)
			if err != nil {
				return err
			}
			cfg.Build.Drafts = true // dev 模式包含草稿

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			buildFn := func(ctx context.Context) error {
				return build.NewBuilder(cfg, root, true).Build(ctx)
			}

			// 首次构建：失败不退出，让用户看到错误页并修复
			if err := buildFn(ctx); err != nil {
				fmt.Fprintln(os.Stderr, "initial build failed:", err)
			}

			addr := fmt.Sprintf("%s:%d", devHost, devPort)
			assetsDir := filepath.Join(root, "assets")
			srv := dev.NewServer(root, cfg.Build.OutDir, assetsDir, addr, buildFn, devOpen)

			if err := buildFn(ctx); err != nil {
				srv.SetError(err)
			}

			watcher, err := dev.NewWatcher(root)
			if err != nil {
				return err
			}
			defer watcher.Close()
			if err := watcher.Start(ctx); err != nil {
				return err
			}

			reloader := dev.NewReloader(srv, watcher, buildFn)
			go reloader.Run(ctx)

			return srv.Start(ctx)
		},
	}
	devCmd.Flags().StringVar(&cfgPath, "config", "site.toml", "配置文件")
	devCmd.Flags().StringVar(&root, "root", ".", "项目根目录")
	devCmd.Flags().IntVar(&devPort, "port", 4321, "端口")
	devCmd.Flags().StringVar(&devHost, "host", "127.0.0.1", "监听地址")
	devCmd.Flags().BoolVar(&devOpen, "open", true, "启动后自动打开浏览器")

	rootCmd.AddCommand(buildCmd, devCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
