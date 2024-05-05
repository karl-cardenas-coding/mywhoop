package export

import (
	"log/slog"
	"os"
	"path"
)

// ExportDataToFile exports the user data to a file
// The file path is optional, if not provided, the file will be created in the data folder in the current directory
// The file will be named user.json
// Example usage: user.ExportDataToFile("data/user.json")
func (f *FileExport) Export(data []byte) error {

	currentDir, err := os.Getwd()
	if err != nil {
		slog.Error("unable to get current directory", err)
		return err
	}

	if f.FilePath == "" {
		f.FilePath = path.Join(currentDir, "data")
	}

	if f.FileType == "" {
		f.FileType = "json"
	}

	if f.FileName == "" {
		f.FileName = "user.json"
	}

	// write the data to a file in the data folder in the current directory
	err = WriteToFile(f.FilePath, data)
	if err != nil {
		slog.Error("unable to write to file", err)
		return err
	}

	return nil

}

// WriteToFile writes data to a file
func WriteToFile(filePath string, data []byte) error {

	// check if the path folder exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// create the data folder
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			slog.Error("unable to create data folder", err)
			return err
		}
	}

	// Remove the file if it exists
	if _, err := os.Stat(path.Join(filePath, "user.json")); err == nil {
		err := os.Remove(path.Join(filePath, "user.json"))
		if err != nil {
			slog.Error("unable to remove file", err)
			return err
		}
	}

	f, err := os.Create(path.Join(filePath, "user.json"))
	if err != nil {
		slog.Error("unable to create file", err)
		return err
	}

	defer f.Close()

	dataPretty := string(data)

	_, err = f.WriteString(dataPretty)
	if err != nil {
		slog.Error("unable to write to file", err)
		return err
	}

	return nil
}
