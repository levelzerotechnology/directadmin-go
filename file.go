package directadmin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// CreateArchive (user) creates a zip of the given files on the server.
//
// The destination path is relative by default.
func (c *UserContext) CreateArchive(destinationPath string, sources ...string) error {
	if destinationPath == "" || len(sources) == 0 {
		return errors.New("no destination path or sources provided")
	}

	if !strings.HasSuffix(destinationPath, ".zip") {
		destinationPath += ".zip"
	}

	body := struct {
		Destination string   `json:"destination"`
		Sources     []string `json:"sources"`
	}{
		Destination: destinationPath,
		Sources:     sources,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "filemanager-actions/create-archive", body, nil); err != nil {
		return err
	}

	return nil
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
		return errors.New("no files provided")
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
	return writeToDisk(outputPath, func() ([]byte, error) {
		return c.DownloadFile(filePath)
	})
}

// ExtractArchive (user) unzips the given file path on the server.
func (c *UserContext) ExtractArchive(destinationDir string, source string, mergeAndOverwrite bool) error {
	if destinationDir == "" || source == "" {
		return errors.New("no destination directory or source provided")
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

// UploadFile uploads the provided byte data as a file for the session user.
func (c *UserContext) UploadFile(uploadToPath string, fileData []byte, overwrite bool) error {
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

	overwriteQuery := "false"
	if overwrite {
		overwriteQuery = "true"
	}

	// Now use this content type which includes the boundary.
	if _, err = c.uploadFile(http.MethodPost, "/api/filemanager-actions/upload?dir="+filepath.Dir(uploadToPath)+"&overwrite="+overwriteQuery+"&name="+filepath.Base(uploadToPath), body.Bytes(), nil, writer.FormDataContentType()); err != nil {
		return err
	}

	return nil
}

// UploadFileFromDisk (user) uploads the provided file for the session user.
//
// Example: UploadFileFromDisk("/domains/domain.tld/public_html/file.zip", "file.zip").
func (c *UserContext) UploadFileFromDisk(uploadToPath string, localFilePath string, overwrite bool) error {
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

	return c.UploadFile(uploadToPath, fileData, overwrite)
}

// writeToDisk wraps a function that returns data and writes the output to the given path.
// It handles file creation, cleanup on failure, and proper error handling.
func writeToDisk(outputPath string, dataFunc func() ([]byte, error)) error {
	if outputPath == "" {
		return errors.New("no file path provided")
	}

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		return fmt.Errorf("file already exists: %s", outputPath)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	defer func() {
		if err != nil {
			fileRemoveErr := os.Remove(outputPath)
			if fileRemoveErr != nil {
				err = fmt.Errorf("%w: %w", err, fileRemoveErr)
			}
		}
	}()

	data, err := dataFunc()
	if err != nil {
		return err
	}

	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
