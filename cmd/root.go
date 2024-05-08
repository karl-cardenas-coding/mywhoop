// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

var (
	// VersionString is the version of the CLI
	VersionString string = "0.0.0"
	// Credentials file containing a Whoop authentication token. Can also be set through ENV variable or configuration file.
	CredentialsFile string
	// cfgFile is the myWhoop configuration file
	cfgFile string
	// Exporter is the exporter to use for storing data
	Exporter string
	// ConfigurationData is the configuration data
	Configuration internal.ConfigurationData
	// GlobalHTTPClient is the HTTP client used for all requests
	GlobalHTTPClient *http.Client
	// UserAgent is the value to use for the User-Agent header
	UserAgent string
	// Debug is a flag to enable debug output
	VerbosityLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "mywhoop",
	Short:            "Interact with the Whoop API and assume ownership of your Whoop data.",
	Long:             `A tool for interacting with the Whoop API and assuming ownership of your Whoop data.`,
	TraverseChildren: true,
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
func Execute(ctx context.Context, args []string, stdout, stderr *os.File) error {

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
	rootCmd.PersistentFlags().StringVarP(&VerbosityLevel, "debug", "d", "", "Enable debug output. Use the values DEBUG, INFO, WARN, ERROR, Default is INFO.")
	rootCmd.PersistentFlags().StringVarP(&Exporter, "exporter", "e", "", "Specify an exporter to use. Supporter exporters are file, and s3. Default is file.")
	rootCmd.PersistentFlags().StringVar(&CredentialsFile, "credentials", "", "File path to the Whoop credentials file that contains a valid authentication token.")

	UserAgent = fmt.Sprintf("mywhoop/%s", VersionString)

}

// InitLogger initializes the logger
func InitLogger() error {
	outputLvl := strings.ToUpper(VerbosityLevel)
	slog.SetDefault(logger(outputLvl))

	envConfigVars, err := internal.ExtractEnvVariables()
	if err != nil {
		return err
	}

	slog.Debug("Environment Configuration",
		slog.Group("Verbosity", slog.String("Level", outputLvl)),
		slog.Group("Config", slog.String("File", cfgFile)),
	)

	ok := internal.CheckConfigFile(cfgFile)
	if ok {
		slog.Info("config file found", "config", cfgFile)
		config, err := internal.GenerateConfigStruct(cfgFile)
		if err != nil {
			slog.Error("unable to generate configuration struct", "error", err)
			os.Exit(1)
		}
		Configuration = config
	}

	// Merge the configuration data from the environment variables
	Configuration.Credentials = envConfigVars.Credentials

	// Prioritize CLI flags

	if Exporter != "" {
		Configuration.Export.Method = Exporter
	}

	if CredentialsFile != "" {
		Configuration.Credentials.CredentialsFile = CredentialsFile
	}

	if outputLvl != "" {
		Configuration.Debug = outputLvl
	}

	if Configuration.Credentials.CredentialsFile == "" {
		Configuration.Credentials.CredentialsFile = internal.DEFAULT_CREDENTIALS_FILE
	}

	return nil

}

// Logger returns a new logger
func logger(verbosity string) *slog.Logger {

	var opts *slog.HandlerOptions

	switch verbosity {
	case "DEBUG":
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	case "INFO":
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	case "WARN":
		opts = &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}
	case "ERROR":
		opts = &slog.HandlerOptions{
			Level: slog.LevelError,
		}
	default:
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
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
