package cmd

import (
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
	// AuthToken is the Whoop API token
	AuthToken string
	// VersionString is the version of the CLI
	VersionString string = "0.0.0"
	// cfgFile is the myWhoop configuration file
	cfgFile string
	// Exporter is the exporter to use for storing data
	Exporter string
	// ConfigurationData is the configuration data
	Configuration internal.ConfigurationData
	// GlobalHTTPClient is the HTTP client used for all requests
	GlobalHTTPClient *http.Client
	// RefreshToken is the Whoop API refresh token
	RefreshToken string
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
func Execute() error {

	authToken := os.Getenv("WHOOP_TOKEN")
	// if authToken == "" {
	// 	return fmt.Errorf("WHOOP_TOKEN environment variable not set")
	// }
	AuthToken = authToken

	refreshToken := os.Getenv("WHOOP_REFRESH_TOKEN")
	RefreshToken = refreshToken

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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	UserAgent = fmt.Sprintf("mywhoop/%s", VersionString)

}

// InitLogger initializes the logger
func InitLogger() {
	slog.SetDefault(logger(VerbosityLevel))
	slog.Debug("Environment Configuration",
		slog.Group("Verbosity", slog.String("Level", VerbosityLevel)),
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

}

// Logger returns a new logger
func logger(verbosity string) *slog.Logger {

	value := strings.ToUpper(verbosity)

	var opts *slog.HandlerOptions

	switch value {
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
