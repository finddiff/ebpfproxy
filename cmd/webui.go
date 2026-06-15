package cmd

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/daeuniverse/dae/pkg/webapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:embed webui-dist/*
var embeddedWebUI embed.FS

var (
	webuiPort       int
	webuiNoOpen     bool
	webuiConfigPath string

	webuiCmd = &cobra.Command{
		Use:   "webui",
		Short: "Start the web UI server for dae.",
		Long: `Start a web-based management interface for dae.
The web UI provides real-time monitoring of system status, proxy groups, routing rules,
connections, configuration management, and log viewing.`,
		Run: func(cmd *cobra.Command, args []string) {
			runWebUI()
		},
	}
)

func init() {
	rootCmd.AddCommand(webuiCmd)
	webuiCmd.Flags().IntVarP(&webuiPort, "port", "p", 2025, "Port for the web UI server")
	webuiCmd.Flags().StringVarP(&webuiConfigPath, "config", "c", "", "Path to dae config file")
}

func getWebuiFS() fs.FS {
	fsys, err := fs.Sub(embeddedWebUI, "webui-dist")
	if err != nil {
		return nil
	}
	return fsys
}

func runWebUI() {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	staticFS := getWebuiFS()
	if staticFS == nil {
		log.Warn("Failed to load embedded WebUI, using fallback")
	}

	server := webapi.NewServer(log, nil, webuiConfigPath, staticFS)

	if webuiConfigPath != "" {
		if data, err := os.ReadFile(webuiConfigPath); err == nil {
			server.SetConfigData(data)
		}
	}

	addr := fmt.Sprintf(":%d", webuiPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.StartBackgroundUpdates(ctx)

	go func() {
		log.Infof("WebUI server starting on http://0.0.0.0%s", addr)
		if err := server.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start web UI server")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Infof("Received signal %v, shutting down...", sig)

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Warn("Error during web UI shutdown")
	}
	log.Info("WebUI server stopped")
}
