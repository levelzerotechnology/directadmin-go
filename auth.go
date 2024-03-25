package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type (
	credentials struct {
		username string
		passkey  string
	}

	LoginHistory struct {
		Attempts  int       `json:"attempts"`
		Host      string    `json:"host"`
		Timestamp time.Time `json:"timestamp"`
	}

	LoginKeyURL struct {
		AllowNetworks []string  `json:"allowNetworks"`
		Created       time.Time `json:"created"`
		CreatedBy     string    `json:"createdBy"`
		Expires       time.Time `json:"expires"`
		Id            string    `json:"id"`
		RedirectURL   string    `json:"redirectURL"`
		URL           string    `json:"url"`
	}
)

func (c *UserContext) CreateLoginURL(loginKeyURL *LoginKeyURL) (*LoginKeyURL, error) {
	var response *LoginKeyURL

	if _, err := c.api.makeRequestN(http.MethodPost, "login-keys/urls", c.credentials, loginKeyURL, response); err != nil {
		return nil, fmt.Errorf("failed to create login URL: %v", err)
	}

	return response, nil
}

func (c *UserContext) GetLoginURLs() ([]*LoginKeyURL, error) {
	var loginKeyURLs []*LoginKeyURL

	if _, err := c.api.makeRequestN(http.MethodGet, "login-keys/urls", c.credentials, nil, &loginKeyURLs); err != nil {
		return nil, fmt.Errorf("failed to get login URLs: %v", err)
	}

	return loginKeyURLs, nil
}

func (c *AdminContext) GetLoginHistory() ([]*LoginHistory, error) {
	var loginHistory []*LoginHistory

	if _, err := c.api.makeRequestN(http.MethodGet, "login-history", c.credentials, nil, &loginHistory); err != nil {
		return nil, fmt.Errorf("failed to get login history: %v", err)
	}

	if len(loginHistory) == 0 {
		return nil, fmt.Errorf("no login history found")
	}

	return loginHistory, nil
}

// GetMyUsername returns the current user's username. This is particularly useful when logging in as another user, as it
// trims the admin/reseller username automatically
func (c *UserContext) GetMyUsername() string {
	// if user is logged in via reseller, we need to remove the reseller username from the context's username
	if strings.Contains(c.credentials.username, "|") {
		return strings.Split(c.credentials.username, "|")[1]
	}

	return c.credentials.username
}

func (c *UserContext) Login() error {
	var response apiGenericResponse

	if _, err := c.api.makeRequest(http.MethodGet, "API_LOGIN_TEST", c.credentials, nil, &response); err != nil {
		return err
	}

	if response.Success != "Login OK" {
		return errors.New("login failed")
	}

	return nil
}

// LoginAsAdmin verifies the provided credentials against the DA API, then returns an admin-level context.
// The passkey can either be the user's password, or a login key
func (a *API) LoginAsAdmin(username string, passkey string) (*AdminContext, error) {
	userCtx, err := a.login(username, passkey)
	if err != nil {
		return nil, err
	}

	adminCtx := AdminContext{
		ResellerContext{
			UserContext: *userCtx,
		},
	}

	if adminCtx.User.Config.UserType != AccountRoleAdmin {
		return nil, fmt.Errorf("account is not an Admin, it is a %v", adminCtx.User.Config.UserType)
	}

	return &adminCtx, nil
}

// LoginAsReseller verifies the provided credentials against the DA API, then returns a reseller-level context.
// The passkey can either be the user's password, or a login key
func (a *API) LoginAsReseller(username string, passkey string) (*ResellerContext, error) {
	userCtx, err := a.login(username, passkey)
	if err != nil {
		return nil, err
	}

	resellerCtx := ResellerContext{
		UserContext: *userCtx,
	}

	if resellerCtx.User.Config.UserType != AccountRoleReseller {
		return nil, fmt.Errorf("account is not an Reseller, it is a %v", resellerCtx.User.Config.UserType)
	}

	return &resellerCtx, nil
}

// LoginAsUser verifies the provided credentials against the DA API, then returns a user-level context.
// The passkey can either be the user's password, or a login key
func (a *API) LoginAsUser(username string, passkey string) (*UserContext, error) {
	userCtx, err := a.login(username, passkey)
	if err != nil {
		return nil, err
	}

	if userCtx.User.Config.UserType != AccountRoleUser {
		return nil, fmt.Errorf("account is not a User, it is a %v", userCtx.User.Config.UserType)
	}

	return userCtx, nil
}

func (c *AdminContext) LoginAsMyReseller(username string) (*ResellerContext, error) {
	return c.api.LoginAsReseller(c.credentials.username+"|"+username, c.credentials.passkey)
}

func (c *ResellerContext) LoginAsMyUser(username string) (*UserContext, error) {
	return c.api.LoginAsUser(c.credentials.username+"|"+username, c.credentials.passkey)
}

func (a *API) login(username string, passkey string) (*UserContext, error) {
	userCtx := UserContext{
		api: a,
		credentials: credentials{
			username: username,
			passkey:  passkey,
		},
	}

	if err := userCtx.Login(); err != nil {
		return nil, err
	}

	userConfig, err := userCtx.GetMyUserConfig()
	if err != nil {
		return nil, err
	}

	userCtx.User.Config = *userConfig

	return &userCtx, nil
}
