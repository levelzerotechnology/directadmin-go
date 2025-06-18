package directadmin

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileMetadata struct {
	AccessTime time.Time `json:"accessTime"`
	BirthTime  time.Time `json:"birthTime"`
	ChangeTime time.Time `json:"changeTime"`
	GID        int       `json:"gid"`
	Group      string    `json:"group"`
	Mode       string    `json:"mode"`
	ModifyTime time.Time `json:"modifyTime"`
	Name       string    `json:"name"`
	SizeBytes  int       `json:"sizeBytes"`
	Symlink    struct {
		Resolved string `json:"resolved"`
		Target   string `json:"target"`
	} `json:"symlink"`
	Type     string `json:"type"`
	UID      int    `json:"uid"`
	UnixMode int    `json:"unixMode"`
	User     string `json:"user"`
}

// CreateDirectory (user) creates the given path, including any missing parent directories.
func (c *UserContext) CreateDirectory(path string) error {
	body := map[string]string{
		"path": path,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "filemanager-actions/mkdir", body, nil); err != nil {
		return err
	}

	return nil
}

// DeleteFiles (user) deletes all the specified files for the session user.
func (c *UserContext) DeleteFiles(skipTrash bool, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}

	normalized := make([]string, len(files))

	for i, f := range files {
		if len(f) == 0 {
			return fmt.Errorf("empty file path provided for file %d", i+1)
		}

		if f[0] != '/' {
			f = "/" + f
		}

		normalized[i] = f
	}

	body := struct {
		Paths []string `json:"paths"`
		Trash bool     `json:"trash"`
	}{
		Paths: normalized,
		Trash: !skipTrash,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "filemanager-actions/remove", body, nil); err != nil {
		return err
	}

	return nil
}

// DownloadFile (user) downloads the given file path from the server.
func (c *UserContext) DownloadFile(filePath string) ([]byte, error) {
	return c.makeRequestNew(http.MethodGet, "filemanager/download?path="+filePath, nil, nil)
}

// DownloadFileToDisk (user) wraps DownloadFile and writes the output to the given path.
func (c *UserContext) DownloadFileToDisk(filePath string, outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("no file path provided")
	}

	response, err := c.DownloadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, response, 0o644)
}

// ExtractFile (user) unzips the given file path on the server.
func (c *UserContext) ExtractFile(destinationDir string, source string, mergeAndOverwrite bool) error {
	if destinationDir == "" || source == "" {
		return fmt.Errorf("no destination directory or source provided")
	}

	// Prepend / to the filePath if necessary.
	if destinationDir[0] != '/' {
		destinationDir = "/" + destinationDir
	}

	// Prepend / to the file if necessary.
	if source[0] != '/' {
		source = "/" + source
	}

	body := struct {
		DestinationDir    string `json:"destinationDir"`
		MergeAndOverwrite bool   `json:"mergeAndOverwrite"`
		Source            string `json:"source"`
	}{
		DestinationDir:    destinationDir,
		MergeAndOverwrite: mergeAndOverwrite,
		Source:            source,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "filemanager-actions/extract-archive", body, nil); err != nil {
		return err
	}

	return nil
}

// GetFileMetadata (user) retrieves file metadata for the given path.
func (c *UserContext) GetFileMetadata(filePath string) (*FileMetadata, error) {
	var response *FileMetadata

	if _, err := c.makeRequestNew(http.MethodGet, "filemanager/metadata?path="+filePath, nil, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// MovePath (user) moves the given file or directory to the new destination.
func (c *UserContext) MovePath(source string, destination string, overwrite bool) error {
	body := struct {
		Destination string `json:"destination"`
		Overwrite   bool   `json:"overwrite"`
		Source      string `json:"source"`
	}{
		Destination: destination,
		Overwrite:   overwrite,
		Source:      source,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "filemanager-actions/move", body, nil); err != nil {
		return err
	}

	return nil
}

// UploadFile uploads the provided byte data as a file for the session user
func (c *UserContext) UploadFile(uploadToPath string, fileData []byte) error {
	// Prepend / to uploadToPath if it doesn't exist.
	if uploadToPath[0] != '/' {
		uploadToPath = "/" + uploadToPath
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(uploadToPath))
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	if _, err = part.Write(fileData); err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("finalizing form data: %w", err)
	}

	// Now use this content type which includes the boundary.
	if _, err = c.uploadFile(http.MethodPost, "/api/filemanager-actions/upload?dir="+filepath.Dir(uploadToPath)+"&name="+filepath.Base(uploadToPath), body.Bytes(), nil, writer.FormDataContentType()); err != nil {
		return err
	}

	return nil
}

// UploadFileFromDisk (user) uploads the provided file for the session user.
//
// Example: UploadFileFromDisk("/domains/domain.tld/public_html/file.zip", "file.zip")
func (c *UserContext) UploadFileFromDisk(uploadToPath string, localFilePath string) error {
	var err error

	localFilePath, err = filepath.Abs(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve file: %w", err)
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return c.UploadFile(uploadToPath, fileData)
}
