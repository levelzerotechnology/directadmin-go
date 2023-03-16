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

	if _, err := c.api.makeRequest(http.MethodGet, "FILE_MANAGER?action=exists&path="+filePath, c.credentials, nil, &response); err != nil {
		return err
	}

	if response.Success != "File exists check" {
		return fmt.Errorf("file doesn't exist: %v", response.Error)
	}

	return nil
}

// CreateFile (user) creates the provided file for the session user
func (c *UserContext) CreateFile(uploadToPath string, filePath string) error {
	return c.CreateFiles(uploadToPath, filePath)
}

// CreateFiles (user) creates the provided files for the session user
func (c *UserContext) CreateFiles(uploadToPath string, filePaths ...string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files provided")
	}

	counter := 0
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.FormDataContentType()

	for _, filePath := range filePaths {
		var err error
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			continue
		}

		file, _ := os.Open(filePath)
		defer file.Close()

		part, _ := writer.CreateFormFile("file"+cast.ToString(counter), filepath.Base(file.Name()))

		if _, err = io.Copy(part, file); err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
	}

	writer.Close()

	var response apiGenericResponse

	// add / to the beginning of uploadToPath if it doesn't exist
	if uploadToPath[0] != '/' {
		uploadToPath = "/" + uploadToPath
	}

	if _, err := c.api.uploadFile("FILE_MANAGER?action=upload&path="+uploadToPath, c.credentials, body, &response, writer.FormDataContentType()); err != nil {
		return err
	}

	if response.Success != "Upload successful" {
		return fmt.Errorf("failed to create file: %v", response.Result)
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

	if _, err := c.api.makeRequest(http.MethodPost, "FILE_MANAGER?action=multiple", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Files deleted" {
		return fmt.Errorf("failed to delete files: %v", response.Result)
	}

	return nil
}

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

	if _, err := c.api.makeRequest(http.MethodPost, "FILE_MANAGER?action=extract", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "File Extracted" {
		return fmt.Errorf("failed to extract file: %v", response.Result)
	}

	return nil
}
