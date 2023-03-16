package directadmin

import (
	"fmt"
	"net/http"
	"strings"
)

type WordPressInstall struct {
	AdminEmail string `json:"adminEmail" yaml:"adminEmail"`
	AdminName  string `json:"adminName" yaml:"adminName"`
	AdminPass  string `json:"adminPass" yaml:"adminPass"`
	DbName     string `json:"dbName" yaml:"dbName"`
	DbPass     string `json:"dbPass" yaml:"dbPass"`
	DbPrefix   string `json:"dbPrefix" yaml:"dbPrefix"`
	DbUser     string `json:"dbUser" yaml:"dbUser"`
	FilePath   string `json:"filePath" yaml:"filePath"`
	Title      string `json:"title" yaml:"title"`
}

type WordPressLocation struct {
	FilePath  string `json:"filePath"`
	Host      string `json:"host"`
	Id        string `json:"id"`
	WebPath   string `json:"webPath"`
	Wordpress struct {
		AutoUpdateMajor bool   `json:"autoUpdateMajor"`
		AutoUpdateMinor bool   `json:"autoUpdateMinor"`
		Error           string `json:"error"`
		SiteURL         string `json:"siteURL"`
		Template        string `json:"template"`
		Title           string `json:"title"`
		Version         string `json:"version"`
	} `json:"wordpress"`
}

func (c *UserContext) CreateWordPressInstall(install WordPressInstall, createDatabase bool) error {
	if createDatabase {
		if err := c.CreateDatabase(install.DbName, install.DbUser, install.DbPass); err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
	}

	contextUsername := c.GetMyUsername()
	dbPrefix := contextUsername + "_"

	if !strings.Contains(install.DbName, dbPrefix) {
		install.DbName = dbPrefix + install.DbName
	}

	if !strings.Contains(install.DbUser, dbPrefix) {
		install.DbUser = dbPrefix + install.DbUser
	}

	if !strings.Contains(install.DbPrefix, "_") {
		install.DbPrefix = install.DbPrefix + "_"
	}

	// remove / from the beginning of FilePath if it's there'
	if install.FilePath[0] == '/' {
		install.FilePath = install.FilePath[1:]
	}

	if _, err := c.api.makeRequestN(http.MethodPost, "wordpress/install", c.credentials, install, nil); err != nil {
		if createDatabase {
			if dbErr := c.DeleteDatabases(install.DbName); dbErr != nil {
				err = fmt.Errorf("%v: %v", dbErr, err)
			}
		}
		return err
	}

	return nil
}

func (c *UserContext) DeleteWordPressInstall(id string) error {
	if _, err := c.api.makeRequestN(http.MethodDelete, "wordpress/locations/"+id, c.credentials, nil, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) GetWordPressInstalls() ([]*WordPressLocation, error) {
	var wordpressInstalls []*WordPressLocation

	if _, err := c.api.makeRequestN(http.MethodGet, "wordpress/locations", c.credentials, nil, &wordpressInstalls); err != nil {
		return nil, fmt.Errorf("failed to get wordpress installs: %v", err)
	}

	return wordpressInstalls, nil
}
