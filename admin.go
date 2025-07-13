package directadmin

import (
	"errors"
	"net/http"
)

type (
	// Admin inherits Reseller which inherits User.
	Admin struct {
		Reseller
	}

	AdminContext struct {
		ResellerContext
	}

	convertAccount struct {
		Account string `json:"account,omitempty"`
		Creator string `json:"creator,omitempty"`
	}
)

// ConvertResellerToUser (admin) converts the given reseller to a user account.
func (c *AdminContext) ConvertResellerToUser(username string, reseller string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "convert-reseller-to-user", convertAccount{Account: username, Creator: reseller}, nil); err != nil {
		return err
	}

	return nil
}

// ConvertUserToReseller (admin) converts the given user account to a reseller account.
func (c *AdminContext) ConvertUserToReseller(username string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "convert-user-to-reseller", convertAccount{Account: username}, nil); err != nil {
		return err
	}

	return nil
}

// DisableRedis (admin) disables Redis for the server.
func (c *AdminContext) DisableRedis() error {
	var response apiGenericResponseNew

	if _, err := c.makeRequestNew(http.MethodPost, "redis/disable", nil, &response); err != nil {
		return err
	}

	return nil
}

// EnableRedis (admin) enables Redis for the server.
func (c *AdminContext) EnableRedis() error {
	var response apiGenericResponseNew

	if _, err := c.makeRequestNew(http.MethodPost, "redis/enable", nil, &response); err != nil {
		return err
	}

	return nil
}

// GetAllUsers (admin) returns an array of all users.
func (c *AdminContext) GetAllUsers() ([]string, error) {
	var users []string

	if _, err := c.makeRequestOld(http.MethodGet, "API_SHOW_ALL_USERS", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetRedisStatus (admin) returns whether Redis is enabled and it's version.
func (c *AdminContext) GetRedisStatus() (bool, string, error) {
	var resp struct {
		Active  bool   `json:"active"`
		Version string `json:"version"`
	}

	if _, err := c.makeRequestNew(http.MethodGet, "redis/status", nil, &resp); err != nil {
		return false, "", err
	}

	return resp.Active, resp.Version, nil
}

// GetResellers (admin) returns an array of all resellers.
func (c *AdminContext) GetResellers() ([]string, error) {
	var users []string

	if _, err := c.makeRequestOld(http.MethodGet, "API_SHOW_RESELLERS", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// MoveUserToReseller (admin) moves the given user to the given reseller.
func (c *AdminContext) MoveUserToReseller(username string, reseller string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "change-user-creator", convertAccount{Account: username, Creator: reseller}, nil); err != nil {
		return err
	}

	return nil
}

// RestartDirectAdmin (admin) restarts the DirectAdmin process on the server.
func (c *AdminContext) RestartDirectAdmin() error {
	var response apiGenericResponseNew

	if _, err := c.makeRequestNew(http.MethodPost, "restart", nil, &response); err != nil {
		return err
	}

	return nil
}

// UpdateDirectAdmin (admin) initiates a DirectAdmin update on the server.
func (c *AdminContext) UpdateDirectAdmin() error {
	var response apiGenericResponseNew

	if _, err := c.makeRequestNew(http.MethodPost, "version/update", nil, &response); err != nil {
		return err
	}

	return nil
}

// UpdateHostname (admin) updates the server's hostname.
func (c *AdminContext) UpdateHostname(hostname string) error {
	if hostname == "" {
		return errors.New("missing hostname")
	}

	if _, err := c.makeRequestNew(http.MethodPost, "server-settings/change-hostname", map[string]string{"hostname": hostname}, nil); err != nil {
		return err
	}

	return nil
}
