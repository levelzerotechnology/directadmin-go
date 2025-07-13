package directadmin

import (
	"errors"
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
		DBName     string `json:"dbName" yaml:"dbName"`
		DBPass     string `json:"dbPass" yaml:"dbPass"`
		DBPrefix   string `json:"dbPrefix" yaml:"dbPrefix"`
		DBUser     string `json:"dbUser" yaml:"dbUser"`
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
		ID        string `json:"id"`
		WebPath   string `json:"webPath"`
		WordPress struct {
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
		ID          int       `json:"id"`
		DisplayName string    `json:"displayName"`
		Email       string    `json:"email"`
		Login       string    `json:"login"`
		Registered  time.Time `json:"registered"`
		Roles       []string  `json:"roles"`
	}
)

// ChangeWordPressUserPassword (user) changes the password of the given WordPress user.
func (c *UserContext) ChangeWordPressUserPassword(locationID string, userID int, password string) error {
	var passwordObject struct {
		Password string `json:"password"`
	}

	if password == "" {
		return errors.New("password cannot be empty")
	}

	passwordObject.Password = password

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/locations/"+locationID+"/users/"+cast.ToString(userID)+"/change-password", passwordObject, nil); err != nil {
		return fmt.Errorf("failed to change wordpress user password: %w", err)
	}

	return nil
}

func (c *UserContext) CreateWordPressInstall(install WordPressInstall, createDatabase bool) error {
	if createDatabase {
		if err := c.CreateDatabaseWithUser(&DatabaseWithUser{
			Database: Database{
				Name: install.DBName,
			},
			Password: install.DBPass,
			User:     install.DBUser,
		}); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	contextUsername := c.GetMyUsername()
	dbPrefix := contextUsername + "_"

	if !strings.Contains(install.DBName, dbPrefix) {
		install.DBName = dbPrefix + install.DBName
	}

	if !strings.Contains(install.DBUser, dbPrefix) {
		install.DBUser = dbPrefix + install.DBUser
	}

	if !strings.Contains(install.DBPrefix, "_") {
		install.DBPrefix += "_"
	}

	// Remove / from the beginning of FilePath if it's there.
	if install.FilePath[0] == '/' {
		install.FilePath = install.FilePath[1:]
	}

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/install", install, nil); err != nil {
		if createDatabase {
			if dbErr := c.DeleteDatabase(install.DBName); dbErr != nil {
				err = fmt.Errorf("%w: %w", dbErr, err)
			}
		}
		return err
	}

	return nil
}

// CreateWordPressInstallQuick (user) creates a new wordpress install and automatically creates a database.
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
		return nil, fmt.Errorf("failed to get wordpress installs: %w", err)
	}

	return wordpressInstalls, nil
}

func (c *UserContext) GetWordPressSSOLink(locationID string, userID int) (string, error) {
	var ssoObject struct {
		URL string `json:"url"`
	}

	if _, err := c.makeRequestNew(http.MethodPost, "wordpress/locations/"+locationID+"/users/"+cast.ToString(userID)+"/sso-login", nil, &ssoObject); err != nil {
		return "", fmt.Errorf("failed to get wordpress installs: %w", err)
	}

	return ssoObject.URL, nil
}

func (c *UserContext) GetWordPressUsers(locationID string) ([]*WordPressUser, error) {
	var wordpressUsers []*WordPressUser

	if _, err := c.makeRequestNew(http.MethodGet, "wordpress/locations/"+locationID+"/users", nil, &wordpressUsers); err != nil {
		return nil, fmt.Errorf("failed to get wordpress users: %w", err)
	}

	return wordpressUsers, nil
}
