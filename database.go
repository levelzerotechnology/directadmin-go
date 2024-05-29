package directadmin

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	DatabaseFormatGz  = DatabaseFormat("gz")
	DatabaseFormatSql = DatabaseFormat("sql")
)

type (
	DatabaseFormat string

	Database struct {
		Name             string `json:"database"`
		DefaultCharset   string `json:"defaultCharset"`
		DefaultCollation string `json:"defaultCollation"`
		DefinerIssues    int    `json:"definerIssues"`
		EventCount       int    `json:"eventCount"`
		RoutineCount     int    `json:"routineCount"`
		SizeBytes        int    `json:"sizeBytes"`
		TableCount       int    `json:"tableCount"`
		TriggerCount     int    `json:"triggerCount"`
		UserCount        int    `json:"userCount"`
		ViewCount        int    `json:"viewCount"`
	}

	DatabaseProcess struct {
		Command  string `json:"command"`
		Database string `json:"database"`
		Host     string `json:"host"`
		Id       int    `json:"id"`
		Info     string `json:"info"`
		State    string `json:"state"`
		Time     int    `json:"time"`
		User     string `json:"user"`
	}

	DatabaseUser struct {
		HostPatterns []string `json:"hostPatterns"`
		Password     string   `json:"password"`
		User         string   `json:"dbuser"`
	}

	DatabaseWithUser struct {
		Database
		Password string `json:"password"`
		User     string `json:"dbuser"`
	}
)

func (c *UserContext) CreateDatabase(database *Database) error {
	database.Name = c.addUsernamePrefix(database.Name)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-db", database, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) CreateDatabaseWithUser(database *DatabaseWithUser) error {
	database.Name = c.addUsernamePrefix(database.Name)
	database.User = c.addUsernamePrefix(database.User)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-db-with-user", database, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) CreateDatabaseUser(databaseUser *DatabaseUser) error {
	databaseUser.User = c.addUsernamePrefix(databaseUser.User)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-user", databaseUser, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) DeleteDatabase(databaseName string) error {
	databaseName = c.addUsernamePrefix(databaseName)

	if _, err := c.makeRequestNew(http.MethodDelete, "db-manage/databases/"+databaseName, nil, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) DownloadDatabase(name string, format DatabaseFormat, filePath string) error {
	name = name + "." + string(format)

	if !strings.Contains(name, c.GetMyUsername()+"_") {
		name = c.GetMyUsername() + "_" + name
	}

	switch format {
	case DatabaseFormatGz, DatabaseFormatSql:
		break
	default:
		return fmt.Errorf("invalid database format: %v", format)
	}

	var file *os.File
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
	}

	resp, err := c.makeRequestOld(http.MethodPost, "DB/"+name, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download database: %v", err)
	}

	if _, err = file.Write(resp); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// ExportDatabase (user) returns an export of the given database
func (c *UserContext) ExportDatabase(databaseName string, gzip bool) ([]byte, error) {
	databaseName = c.addUsernamePrefix(databaseName)

	export, err := c.makeRequestNew(http.MethodGet, "db-manage/databases/"+databaseName+"/export?gzip="+strconv.FormatBool(gzip), nil, nil)
	if err != nil {
		return nil, err
	}

	return export, nil
}

// GetDatabase (user) returns the given database
func (c *UserContext) GetDatabase(databaseName string) (*Database, error) {
	databaseName = c.addUsernamePrefix(databaseName)

	var database Database

	if _, err := c.makeRequestNew(http.MethodGet, "db-show/databases/"+databaseName, nil, &database); err != nil {
		return nil, err
	}

	return &database, nil
}

// GetDatabases (user) returns an array of the session user's databases
func (c *UserContext) GetDatabases() ([]*Database, error) {
	var databases []*Database

	if _, err := c.makeRequestNew(http.MethodGet, "db-show/databases", nil, &databases); err != nil {
		return nil, err
	}

	return databases, nil
}

// GetDatabaseProcesses (admin) returns an array of current database processes
func (c *UserContext) GetDatabaseProcesses() ([]*DatabaseProcess, error) {
	var databaseProcesses []*DatabaseProcess

	if _, err := c.makeRequestNew(http.MethodGet, "db-monitor/processes", nil, &databaseProcesses); err != nil {
		return nil, err
	}

	return databaseProcesses, nil
}

// ImportDatabase (user) imports the given database export into the given database
func (c *UserContext) ImportDatabase(databaseName string, emptyExistingDatabase bool, sql []byte) error {
	databaseName = c.addUsernamePrefix(databaseName)

	var byteBuffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&byteBuffer)

	formFile, err := multipartWriter.CreateFormFile("sqlfile", "filename")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = formFile.Write(sql); err != nil {
		return fmt.Errorf("failed to write to form file: %w", err)
	}

	if err = multipartWriter.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	if _, err = c.uploadFile(http.MethodPost, "/api/db-manage/databases/"+databaseName+"/import?clean="+strconv.FormatBool(emptyExistingDatabase), byteBuffer.Bytes(), nil, multipartWriter.FormDataContentType()); err != nil {
		return err
	}

	return nil
}

// UpdateDatabaseUserHosts (user) updates the given database user's hosts
func (c *UserContext) UpdateDatabaseUserHosts(username string, hosts []string) error {
	username = c.addUsernamePrefix(username)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/users/"+username+"/change-hosts", hosts, nil); err != nil {
		return err
	}

	return nil
}

// UpdateDatabaseUserPassword (user) updates the given database user's password
func (c *UserContext) UpdateDatabaseUserPassword(username string, password string) error {
	username = c.addUsernamePrefix(username)

	newPassword := struct {
		NewPassword string `json:"newPassword"`
	}{
		password,
	}

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/users/"+username+"/change-password", newPassword, nil); err != nil {
		return err
	}

	return nil
}
