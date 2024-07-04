package directadmin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cast"
)

type (
	WordPressInstall struct {
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

	WordPressInstallQuick struct {
		AdminEmail string `json:"adminEmail" yaml:"adminEmail"`
		AdminName  string `json:"adminName" yaml:"adminName"`
		AdminPass  string `json:"adminPass" yaml:"adminPass"`
		FilePath   string `json:"filePath" yaml:"filePath"`
		Title      string `json:"title" yaml:"title"`
	}

	WordPressLocation struct {
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

	WordPressUser struct {
		Id          int       `json:"id"`
		DisplayName string    `json:"displayName"`
		Email       string    `json:"email"`
		Login       string    `json:"login"`
		Registered  time.Time `json:"registered"`
		Roles       []string  `json:"roles"`
	}
)

// ChangeWordPressUserPassword (user) changes the password of the given wordpress user.
func (c *UserContext) ChangeWordPressUserPassword(locationId string, userId int, password string) error {
	var passwordObject struct {
		Password string `json:"password"`
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	passwordObject.Password = password

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/locations/"+locationId+"/users/"+cast.ToString(userId)+"/change-password", passwordObject, nil); err != nil {
		return fmt.Errorf("failed to change wordpress user password: %v", err)
	}

	return nil
}

func (c *UserContext) CreateWordPressInstall(install WordPressInstall, createDatabase bool) error {
	if createDatabase {
		if err := c.CreateDatabaseWithUser(&DatabaseWithUser{
			Database: Database{
				Name: install.DbName,
			},
			Password: install.DbPass,
			User:     install.DbUser,
		}); err != nil {
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

	// remove / from the beginning of FilePath if it's there
	if install.FilePath[0] == '/' {
		install.FilePath = install.FilePath[1:]
	}

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/install", install, nil); err != nil {
		if createDatabase {
			if dbErr := c.DeleteDatabase(install.DbName); dbErr != nil {
				err = fmt.Errorf("%v: %v", dbErr, err)
			}
		}
		return err
	}

	return nil
}

// CreateWordPressInstallQuick (user) creates a new wordpress install and automatically creates a database
func (c *UserContext) CreateWordPressInstallQuick(install WordPressInstallQuick) error {
	// remove / from the beginning of FilePath if it's there
	if install.FilePath[0] == '/' {
		install.FilePath = install.FilePath[1:]
	}

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/install-quick", install, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) DeleteWordPressInstall(id string) error {
	if _, err := c.makeRequestNew(http.MethodDelete, "wordpress/locations/"+id, nil, nil); err != nil {
		return err
	}

	return nil
}

func (c *UserContext) GetWordPressInstalls() ([]*WordPressLocation, error) {
	var wordpressInstalls []*WordPressLocation

	if _, err := c.makeRequestNew(http.MethodGet, "wordpress/locations", nil, &wordpressInstalls); err != nil {
		return nil, fmt.Errorf("failed to get wordpress installs: %v", err)
	}

	return wordpressInstalls, nil
}

func (c *UserContext) GetWordPressSSOLink(locationId string, userId int) (string, error) {
	var ssoObject struct {
		URL string `json:"url"`
	}

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/locations/"+locationId+"/users/"+cast.ToString(userId)+"/sso-login", nil, &ssoObject); err != nil {
		return "", fmt.Errorf("failed to get wordpress installs: %v", err)
	}

	return ssoObject.URL, nil
}

func (c *UserContext) GetWordPressUsers(locationId string) ([]*WordPressUser, error) {
	var wordpressUsers []*WordPressUser

	if _, err := c.makeRequestNew(http.MethodGet, "wordpress/locations/"+locationId+"/users", nil, &wordpressUsers); err != nil {
		return nil, fmt.Errorf("failed to get wordpress users: %v", err)
	}

	return wordpressUsers, nil
}
