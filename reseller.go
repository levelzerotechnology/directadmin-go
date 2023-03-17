package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/spf13/cast"
)

// Reseller inherits User
type Reseller struct {
	User
}

type ResellerContext struct {
	UserContext
}

// CheckUserExists (reseller) checks if the given user exists
func (c *ResellerContext) CheckUserExists(username string) error {
	return c.checkObjectExists(url.Values{
		"type":  {"username"},
		"value": {username},
	})
}

// CreateUser (reseller) create a user.
//
// The following fields must be populated: Domain, Email, IpAddresses, Package, Username
func (c *ResellerContext) CreateUser(user UserConfig, password string, emailUser bool, customPackage *Package) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "create")
	body.Set("add", "Submit")
	body.Set("domain", user.Domain)
	body.Set("email", user.Email)
	body.Set("ip", user.IpAddresses[0])
	body.Set("package", user.Package)
	body.Set("passwd", password)
	body.Set("passwd2", password)
	body.Set("username", user.Username)

	if emailUser {
		body.Set("notify", "yes")
	} else {
		body.Set("notify", "no")
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_ACCOUNT_USER?action=create", c.credentials, body, &response); err != nil {
		return fmt.Errorf("failed to create user account: %v", err)
	}

	if response.Success != "User "+user.Username+" created" {
		return fmt.Errorf("failed to create user account: %v", response.Result)
	}

	return nil
}

// DeleteUsers (reseller) deletes all the users associated with the given usernames
func (c *ResellerContext) DeleteUsers(usernames ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("confirmed", "Confirm")
	body.Set("delete", "yes")

	for index, username := range usernames {
		body.Set("select"+cast.ToString(index), username)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_SELECT_USERS", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "User deleted" {
		return fmt.Errorf("failed to delete user(s): %v", response.Result)
	}

	return nil
}

// GetMyUsers (reseller) returns all users belonging to the session user
func (c *ResellerContext) GetMyUsers() ([]string, error) {
	var users []string

	if _, err := c.api.makeRequest(http.MethodGet, "API_SHOW_USERS", c.credentials, nil, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("no users could be found")
	}

	return users, nil
}

// GetMyUsersWithUsage (reseller) returns all users belonging to the session user, along with their UserConfig and UserUsage data
func (c *ResellerContext) GetMyUsersWithUsage() ([]User, error) {
	var err error
	var usernames []string
	var users []User

	usernames, err = c.GetMyUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	var errs []error
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(usernames))

	for _, username := range usernames {
		// convert to local variable to prevent variable overwrite
		userToProcess := username

		go func(username string) {
			defer wg.Done()

			config, err := c.GetUserConfig(username)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to get user config: %v", err))
				mu.Unlock()
				return
			}

			usage, err := c.GetUserUsage(username)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to get user usage: %v", err))
				mu.Unlock()
				return
			}

			mu.Lock()
			users = append(users, User{
				Config: *config,
				Usage:  *usage,
			})
			mu.Unlock()
		}(userToProcess)
	}

	wg.Wait()

	if len(errs) > 0 {
		counter := 0
		var errStrings []string

		for _, err := range errs {
			counter++
			errStrings = append(errStrings, "error "+cast.ToString(counter)+": "+err.Error())
		}

		return nil, errors.New(strings.Join(errStrings, "; "))
	}

	if len(users) == 0 {
		return nil, errors.New("no users were found")
	}

	return users, nil
}

// GetUserConfig (reseller) returns the given user's config
func (c *ResellerContext) GetUserConfig(username string) (*UserConfig, error) {
	var config UserConfig
	var err error
	var rawConfig rawUserConfig

	if _, err := c.api.makeRequest(http.MethodGet, "API_SHOW_USER_CONFIG?user="+username, c.credentials, nil, &rawConfig); err != nil {
		return nil, err
	}

	config, err = rawConfig.parse()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetUserUsage (reseller) returns the given user's usage.
func (c *ResellerContext) GetUserUsage(username string) (*UserUsage, error) {
	var rawUsage rawUserUsage
	var usage UserUsage

	if _, err := c.api.makeRequest(http.MethodGet, "API_SHOW_USER_USAGE?bytes=yes&user="+username, c.credentials, nil, &rawUsage); err != nil {
		return nil, err
	}

	usage = rawUsage.parse()

	return &usage, nil
}

func (c *ResellerContext) SuspendUser(username string) error {
	return c.toggleUserSuspension(true, username)
}

func (c *ResellerContext) SuspendUsers(usernames ...string) error {
	return c.toggleUserSuspension(true, usernames...)
}

func (c *ResellerContext) UnsuspendUser(username string) error {
	return c.toggleUserSuspension(false, username)
}

func (c *ResellerContext) UnsuspendUsers(usernames ...string) error {
	return c.toggleUserSuspension(false, usernames...)
}

func (c *ResellerContext) toggleUserSuspension(suspend bool, usernames ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	if suspend {
		body.Set("suspend", "Suspend")
	} else {
		body.Set("suspend", "Unsuspend")
	}

	counter := 0
	for _, username := range usernames {
		body.Set("select"+cast.ToString(counter), username)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_SELECT_USERS", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "User suspended" {
		return fmt.Errorf("failed to delete user(s): %v", response.Result)
	}

	return nil
}
