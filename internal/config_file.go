// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v2"
)

func CheckConfigFile(filePath string) bool {

	// Check if config file is in  $HOME/.mywhoop.yaml
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("unable to get user home directory", "error", err)
			return false
		}
		filePath = path.Join(home + DEFAULT_CONFIG_FILE)

		// Check if specified config file exists in $HOME directory
		_, err = os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Info("config file not found in ~/.mywhoop.yaml", "config", err)
			}
			return false
		}
	}

	// Check if specified config file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Info("config file not found", "config", err)
		return false
	}

	return true

}

// GenerateLambdaDeleteList is a function that takes a file path as input and returns a list of Lambdas to be deleted
func GenerateConfigStruct(filePath string) (ConfigurationData, error) {

	fileType, err := determineFileType(filePath)
	if err != nil {
		return ConfigurationData{}, err
	}

	if fileType != "yaml" {
		return ConfigurationData{}, errors.New("invalid file type provided. Must be of type json, yaml or yml")
	}

	configuration, err := readConfigFileYaml(filePath)
	if err != nil {
		return configuration, err
	}

	err = validateConfiguration(configuration)
	if err != nil {
		slog.Info("invalid configuration", "error", err)
		return configuration, err
	}

	return configuration, nil
}

// readConfigFileYaml is a function that takes a file path as input and returns a list of Lambdas to be deleted. A YAML file is expected.
func readConfigFileYaml(file string) (ConfigurationData, error) {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		slog.Error("unable to read the input file", "error", err)
		return ConfigurationData{}, errors.New("unable to read the input file")
	}

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return ConfigurationData{}, errors.New("unable to read the input file")
	}

	config := ConfigurationData{}

	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return ConfigurationData{}, errors.New("unable to unmarshall the YAML file")
	}

	// Set debug values to all upper case.
	config.Debug = strings.ToUpper(config.Debug)

	return config, err

}

// determineFileType validates the existence of an input file and ensures its prefix is json | yaml | yml
func determineFileType(file string) (string, error) {
	f, err := os.Stat(file)
	if err != nil {
		return "none", errors.New("unable to read the input file")
	}
	var fileType string

	switch {
	case strings.HasSuffix(f.Name(), "yaml"):
		fileType = "yaml"

	case strings.HasSuffix(f.Name(), "json"):
		fileType = "json"

	case strings.HasSuffix(f.Name(), "yml"):
		fileType = "yaml"

	default:
		fileType = "none"
		err = errors.New("invalid file type provided. Must be of type json, yaml or yml")
	}

	return fileType, err
}

// validateConfiguration is a function that validates the configuration data
func validateConfiguration(config ConfigurationData) error {

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(config)
	if err != nil {
		slog.Info("Invalid configuration file provided", "", "")
		for _, err := range err.(validator.ValidationErrors) {
			slog.Info(
				"The following field failed validation",
				"field", err.Field(),
				"error", err.ActualTag())
			slog.Debug("Additional Context: ", "params", err)

		}
		slog.Debug("Configuration Received", "config", config)

		return err
	}

	return nil
}
