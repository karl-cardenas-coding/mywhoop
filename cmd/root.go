package cmd

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	// VersionString is the version of the CLI
	VersionString string = "0.0.0"
	// cfgFile is the myWhoop configuration file
	cfgFile string
	// GlobalHTTPClient is the HTTP client used for all requests
	GlobalHTTPClient *http.Client
	// UserAgent is the value to use for the User-Agent header
	UserAgent string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mywhoop",
	Short: "Interact with the Whoop API and assume ownership of your Whoop data.",
	Long:  `A tool for interacting with the Whoop API and assuming ownership of your Whoop data.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Help()
		if err != nil {
			slog.Info("Error running help command")
			return err
		}

		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	GlobalHTTPClient = createHTTPClient()
	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func init() {

	// Global Flag - Config File
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "myWhoop config file - default is $HOME/.mywhoop.yaml")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	UserAgent = fmt.Sprintf("mywhoop/%s", VersionString)

}

// createHTTPClient creates an HTTP client with TLS
func createHTTPClient() *http.Client {

	// Setup client header to use TLS 1.2
	tr := &http.Transport{
		// Reads PROXY configuration from environment variables
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Needed due to custom client being leveraged, otherwise HTTP2 will not be used.
	tr.ForceAttemptHTTP2 = true

	// Create the client
	return &http.Client{Transport: tr}
}
