package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cast"
)

const (
	DatabaseFormatGz  = DatabaseFormat("gz")
	DatabaseFormatSql = DatabaseFormat("sql")
)

type DatabaseFormat string

type Database struct {
	Name  string `json:"name" yaml:"name"`
	Size  int    `json:"size" yaml:"size"`
	Users int    `json:"users" yaml:"users"`
}

func (c *UserContext) CreateDatabase(name string, dbUser string, dbPassword string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("name", name)
	body.Set("user", dbUser)
	body.Set("passwd", dbPassword)
	body.Set("passwd2", dbPassword)

	if _, err := c.api.makeRequest(http.MethodPost, "API_DATABASES?action=create", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Database Created" {
		return fmt.Errorf("failed to create database: %v", response.Result)
	}

	return nil
}

func (c *UserContext) DeleteDatabases(names ...string) error {
	var response apiGenericResponse

	body := url.Values{}

	for index, name := range names {
		if !strings.Contains(name, c.GetMyUsername()+"_") {
			name = c.GetMyUsername() + "_" + name
		}

		body.Set("select"+cast.ToString(index), name)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_DATABASES?action=delete", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Databases Deleted" {
		return fmt.Errorf("failed to delete database: %v", response.Result)
	}

	return nil
}

func (c *UserContext) DownloadDatabase(name string, format DatabaseFormat, filePath string) error {
	var response apiGenericResponse

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

	if _, err := c.api.makeRequest(http.MethodPost, "DB/"+name, c.credentials, nil, &response, file); err != nil {
		return fmt.Errorf("failed to download database: %v", err)
	}

	return nil
}

// GetDatabases (user) returns an array of the session user's databases
func (c *UserContext) GetDatabases() ([]Database, error) {
	var databases []Database
	rawDatabases := struct {
		Databases map[string]struct {
			Name  string `json:"database"`
			Size  string `json:"size"`
			Users string `json:"nusers"`
		}
	}{}

	if _, err := c.api.makeRequest(http.MethodGet, "DB", c.credentials, nil, &rawDatabases); err != nil {
		return nil, err
	}

	for id, database := range rawDatabases.Databases {
		if id != "info" {
			databases = append(databases, Database{
				Name:  strings.Replace(database.Name, c.credentials.username+"_", "", 1),
				Size:  cast.ToInt(database.Size),
				Users: cast.ToInt(database.Users),
			})
		}
	}

	if len(databases) == 0 {
		return nil, errors.New("no databases were found")
	}

	return databases, nil
}

// ListDatabases (user) returns an array of all databases for the session user
func (c *UserContext) ListDatabases() (databaseList []string, err error) {
	if _, err = c.api.makeRequest(http.MethodGet, "API_DATABASES", c.credentials, nil, &databaseList); err != nil {
		return nil, err
	}

	return databaseList, nil
}
