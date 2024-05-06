package export

type Export interface {
	Setup() error
	Export(data []byte) error
	CleanUp() error
}

type FileExport struct {
	FilePath string `yaml:"filePath"`
	FileType string `yaml:"fileType"`
	FileName string `yaml:"fileName"`
}

type AWS_S3 struct {
	Region string `yaml:"region"`
	Bucket string `yaml:"bucket"`
}
