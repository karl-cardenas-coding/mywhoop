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
	yaml "gopkg.in/yaml.v3"
)

func CheckConfigFile(filePath string) (bool, string) {

	// Check if config file is in  $HOME/.mywhoop.yaml
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("unable to get user home directory", "error", err)
			return false, ""
		}
		filePath = path.Join(home, DEFAULT_CONFIG_FILE)

		// Check if specified config file exists in $HOME directory
		_, err = os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Debug("config file not found in ~/.mywhoop.yaml", "config", err)
			}
			return false, ""
		}
		return true, filePath

	}

	// Check if specified config file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Info("config file not found", "config", err)
		return false, ""
	}

	return true, filePath

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
		slog.Error("file not found", "file", file)
		return ConfigurationData{}, errors.New("unable to read the input file")
	}

	fileContent, err := os.ReadFile(file)
	if err != nil {
		slog.Error("unable to read the content of the file", "file", file)
		return ConfigurationData{}, err
	}

	config := ConfigurationData{}

	dc := yaml.NewDecoder(strings.NewReader(string(fileContent)))
	dc.KnownFields(true)

	if err := dc.Decode(&config); err != nil {
		return ConfigurationData{}, errors.New("unable to decode the YAML file. Ensure the file is in the correct format and that all fields are correct")
	}

	// Set debug values to all upper case.
	config.Debug = strings.ToUpper(config.Debug)

	return config, err

}

// determineFileType validates the existence of an input file and ensures its prefix is json | yaml | yml
// If the file prefix is yml then it is converted to yaml.
func determineFileType(file string) (string, error) {

	switch {
	case strings.HasSuffix(file, "yaml"):
		return "yaml", nil

	case strings.HasSuffix(file, "json"):
		return "json", nil

	case strings.HasSuffix(file, "yml"):
		return "yaml", nil

	default:
		return "", errors.New("invalid file type provided. Must be of type json, yaml or yml")
	}
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
