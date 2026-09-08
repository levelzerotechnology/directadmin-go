package directadmin

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	DatabaseFormatGZ  = DatabaseFormat("gz")
	DatabaseFormatSQL = DatabaseFormat("sql")
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
		ID       int    `json:"id"`
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

// CreateDatabase (user) creates a new database.
func (c *UserContext) CreateDatabase(database *Database) error {
	database.Name = c.addUsernamePrefix(database.Name)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-db", database, nil); err != nil {
		return err
	}

	return nil
}

// CreateDatabaseWithUser (user) creates a new database and database user.
func (c *UserContext) CreateDatabaseWithUser(database *DatabaseWithUser) error {
	database.Name = c.addUsernamePrefix(database.Name)
	database.User = c.addUsernamePrefix(database.User)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-db-with-user", database, nil); err != nil {
		return err
	}

	return nil
}

// CreateDatabaseUser (user) creates a new database user with the specified username, password, and host patterns.
// It prepends the username prefix if the caller didn't do it.
func (c *UserContext) CreateDatabaseUser(databaseUser *DatabaseUser) error {
	databaseUser.User = c.addUsernamePrefix(databaseUser.User)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/create-user", databaseUser, nil); err != nil {
		return err
	}

	return nil
}

// DeleteDatabase (user) removes a database identified by databaseName after applying the username prefix.
func (c *UserContext) DeleteDatabase(databaseName string) error {
	databaseName = c.addUsernamePrefix(databaseName)

	if _, err := c.makeRequestNew(http.MethodDelete, "db-manage/databases/"+databaseName, nil, nil); err != nil {
		return err
	}

	return nil
}

// DownloadDatabase (user) retrieves a database by name and format from the server as a
// stream. The caller closes it. The method appends the username and ensures the file uses a
// valid DatabaseFormat (gz or sql).
//
// The old API reports failures as JSON with a 200, so the head of the stream is checked
// before it is handed back.
func (c *UserContext) DownloadDatabase(name string, format DatabaseFormat) (io.ReadCloser, error) {
	switch format {
	case DatabaseFormatGZ, DatabaseFormatSQL:
		break
	default:
		return nil, fmt.Errorf("invalid database format: %v", format)
	}

	name = name + "." + string(format)

	if !strings.Contains(name, c.GetMyUsername()+"_") {
		name = c.GetMyUsername() + "_" + name
	}

	body, err := c.downloadStream(http.MethodPost, c.getRequestURLOld("DB/"+name)+"?json=yes")
	if err != nil {
		return nil, err
	}

	return checkStreamError(body)
}

// ExportDatabase (user) returns an export of the given database as a stream, since a dump
// routinely runs to hundreds of megabytes. The caller closes it.
//
// Wrap it in ReadAll to get the whole thing as bytes.
func (c *UserContext) ExportDatabase(databaseName string, gzip bool) (io.ReadCloser, error) {
	databaseName = c.addUsernamePrefix(databaseName)

	return c.downloadStream(http.MethodGet, c.getRequestURLNew("db-manage/databases/"+databaseName+"/export?gzip="+strconv.FormatBool(gzip)))
}

// GetDatabase (user) returns the given database.
func (c *UserContext) GetDatabase(databaseName string) (*Database, error) {
	databaseName = c.addUsernamePrefix(databaseName)

	var database Database

	if _, err := c.makeRequestNew(http.MethodGet, "db-show/databases/"+databaseName, nil, &database); err != nil {
		return nil, err
	}

	return &database, nil
}

// GetDatabases (user) returns an array of the session user's databases.
func (c *UserContext) GetDatabases() ([]*Database, error) {
	var databases []*Database

	if _, err := c.makeRequestNew(http.MethodGet, "db-show/databases", nil, &databases); err != nil {
		return nil, err
	}

	return databases, nil
}

// GetDatabaseProcesses (admin) returns an array of current database processes.
func (c *UserContext) GetDatabaseProcesses() ([]*DatabaseProcess, error) {
	var databaseProcesses []*DatabaseProcess

	if _, err := c.makeRequestNew(http.MethodGet, "db-monitor/processes", nil, &databaseProcesses); err != nil {
		return nil, err
	}

	return databaseProcesses, nil
}

// CreatePHPMyAdminLoginURL (user) returns a one-time URL that logs directly into PHPMyAdmin for the active account.
func (c *UserContext) CreatePHPMyAdminLoginURL() (string, error) {
	var response struct {
		URL string `json:"url"`
	}

	if _, err := c.makeRequestNew(http.MethodPost, "phpmyadmin-sso/account-access", nil, &response); err != nil {
		return "", err
	}

	return response.URL, nil
}

// ImportDatabase (user) imports the given dump into the given database, reading it as it is
// sent, so a dump of any size can be imported.
//
// sql must be able to report its length: a *bytes.Reader (bytes.NewReader(data)),
// *bytes.Buffer, *strings.Reader or *os.File.
func (c *UserContext) ImportDatabase(databaseName string, emptyExistingDatabase bool, sql io.Reader) error {
	databaseName = c.addUsernamePrefix(databaseName)

	body, length, contentType, err := multipartFileBody("sqlfile", "filename", sql)
	if err != nil {
		return err
	}

	if _, err = c.uploadStream(http.MethodPost, "/api/db-manage/databases/"+databaseName+"/import?clean="+strconv.FormatBool(emptyExistingDatabase), body, length, nil, contentType); err != nil {
		return err
	}

	return nil
}

// UpdateDatabaseUserHosts (user) updates the given database user's hosts.
func (c *UserContext) UpdateDatabaseUserHosts(username string, hosts []string) error {
	username = c.addUsernamePrefix(username)

	if _, err := c.makeRequestNew(http.MethodPost, "db-manage/users/"+username+"/change-hosts", hosts, nil); err != nil {
		return err
	}

	return nil
}

// UpdateDatabaseUserPassword (user) updates the given database user's password.
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
