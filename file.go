package directadmin

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cast"
)

// CheckFileExists (user) checks if the given file exists on the server
func (c *UserContext) CheckFileExists(filePath string) error {
	var response apiGenericResponse

	if _, err := c.makeRequestOld(http.MethodGet, "FILE_MANAGER?action=exists&path="+filePath, nil, &response); err != nil {
		return err
	}

	if response.Success != "File exists check" {
		return fmt.Errorf("file doesn't exist: %v", response.Error)
	}

	return nil
}

// DeleteFiles (user) deletes all the specified files for the session user
func (c *UserContext) DeleteFiles(skipTrash bool, files ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("button", "delete")

	if skipTrash {
		body.Set("trash", "no")
	} else {
		body.Set("trash", "yes")
	}

	for index, file := range files {
		// add / to the beginning of filePath if it doesn't exist
		if file[0] != '/' {
			file = "/" + file
		}

		body.Set("select"+cast.ToString(index), file)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "FILE_MANAGER?action=multiple", body, &response); err != nil {
		return err
	}

	if response.Success != "Files deleted" {
		return fmt.Errorf("failed to delete files: %v", response.Result)
	}

	return nil
}

// DownloadFile (user) downloads the given file path from the server
func (c *UserContext) DownloadFile(filePath string) ([]byte, error) {
	return c.makeRequestNew(http.MethodGet, "filemanager/download?path="+filePath, nil, nil)
}

// DownloadFileToDisk (user) wraps DownloadFile and writes the output to the given path
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

// ExtractFile unzips the given file path on the server
func (c *UserContext) ExtractFile(filePath string, file string) error {
	var response apiGenericResponse

	// add / to the beginning of filePath if it doesn't exist
	if filePath[0] != '/' {
		filePath = "/" + filePath
	}

	// add / to the beginning of file if it doesn't exist
	if file[0] != '/' {
		file = "/" + file
	}

	body := url.Values{}
	body.Set("directory", filePath)
	body.Set("path", file)
	body.Set("page", "2")

	if _, err := c.makeRequestOld(http.MethodPost, "FILE_MANAGER?action=extract", body, &response); err != nil {
		return err
	}

	if response.Success != "File Extracted" {
		return fmt.Errorf("failed to extract file: %v", response.Result)
	}

	return nil
}

// UploadFile uploads the provided byte data as a file for the session user
func (c *UserContext) UploadFile(uploadToPath string, fileData []byte) error {
	// Add / to the beginning of uploadToPath if it doesn't exist.
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
	if _, err = c.uploadFile(http.MethodPost, "/api/filemanager/upload?dir="+filepath.Dir(uploadToPath)+"&name="+filepath.Base(uploadToPath), body.Bytes(), nil, writer.FormDataContentType()); err != nil {
		return err
	}

	return nil
}

// UploadFileFromDisk (user) uploads the provided file for the session user
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
