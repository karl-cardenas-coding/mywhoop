// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/internal"
	"github.com/spf13/cobra"
)

var (
	// Credentials file containing a Whoop authentication token. Can also be set through ENV variable or configuration file.
	CredentialsFile string
	// cfgFile is the myWhoop configuration file
	cfgFile string
	// ConfigurationData is the configuration data
	Configuration internal.ConfigurationData
	// UserAgent is the value to use for the User-Agent header
	UserAgent string
	// Debug is a flag to enable debug output
	VerbosityLevel string
	// VersionString is the version of the CLI
	VersionString string = "0.0.0"
	// StaticAssets is the embedded static assets
	GlobalStaticAssets embed.FS
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
func Execute(ctx context.Context, args []string, stdout, stderr *os.File, staticAssets embed.FS) error {

	GlobalStaticAssets = staticAssets

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
	rootCmd.PersistentFlags().StringVar(&CredentialsFile, "credentials", "", "File path to the Whoop credentials file that contains a valid authentication token.")

	UserAgent = fmt.Sprintf("mywhoop/%s", VersionString)

}

// InitLogger initializes the logger
func InitLogger(cfg *internal.ConfigurationData) error {

	envConfigVars, err := internal.ExtractEnvVariables()
	if err != nil {
		return err
	}

	ok, configFilePath := internal.CheckConfigFile(cfgFile)
	if ok {
		slog.Info("config file found", "config", configFilePath)
		config, err := internal.GenerateConfigStruct(configFilePath)
		if err != nil {
			slog.Error("unable to generate configuration struct", "error", err)
			os.Exit(1)
		}
		Configuration = config
	}

	// Merge the configuration data from the environment variables
	cfg.Credentials = envConfigVars.Credentials

	// Prioritize CLI flags

	if CredentialsFile != "" {
		slog.Info("User provided credentials file path", "path", CredentialsFile)
		cfg.Credentials.CredentialsFile = CredentialsFile
	}

	outputLvl := strings.ToUpper(VerbosityLevel)
	if outputLvl == "" {
		outputLvl = cfg.Debug
	}
	slog.SetDefault(logger(outputLvl))
	slog.Debug("Environment Configuration",
		slog.Group("Verbosity", slog.String("Level", outputLvl)),
		slog.Group("Config", slog.String("File", cfgFile)),
	)

	if cfg.Credentials.CredentialsFile == "" {
		slog.Info("No credentials file provided. Using default credentials file")
		cfg.Credentials.CredentialsFile = internal.DEFAULT_CREDENTIALS_FILE
	}
	slog.Debug("Configuration", "Config", cfg)

	return nil

}

// changeTimeFormat changes the timestamp of the logger.
func changeTimeFormat(groups []string, a slog.Attr) slog.Attr {

	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(time.Now().Format("2006/01/02 15:04:05"))
	}
	return a

}

// Logger returns a new logger
func logger(verbosity string) *slog.Logger {

	var opts *slog.HandlerOptions

	switch verbosity {
	case "DEBUG":
		opts = &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: changeTimeFormat,
		}
	case "INFO":
		opts = &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: changeTimeFormat,
		}
	case "WARN":
		opts = &slog.HandlerOptions{
			Level:       slog.LevelWarn,
			ReplaceAttr: changeTimeFormat,
		}
	case "ERROR":
		opts = &slog.HandlerOptions{
			Level:       slog.LevelError,
			ReplaceAttr: changeTimeFormat,
		}
	default:
		opts = &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: changeTimeFormat,
		}
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
