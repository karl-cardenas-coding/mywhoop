package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server mode.",
	Long:  "Start myWhoop in server mode and download data from Whoop API on a regular basis.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

// login authenticates with Whoop API and gets an access token
func server(ctx context.Context) error {
	slog.Info("Server mode enabled")
	InitLogger()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		err := internal.StartServer(ctx, Configuration, GlobalHTTPClient)
		if err != nil {
			slog.Error("unable to start server", "error", err)
			os.Exit(1)
		}

	}()

	sig := <-sigs
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		slog.Info("Server shutdown signal received")
		slog.Info("Cleaning up server resources")

		slog.Info("Server shutdown complete")
		os.Exit(0)
	}

	return nil
}
