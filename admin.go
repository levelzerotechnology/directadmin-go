package directadmin

import "net/http"

// Admin inherits Reseller which inherits User
type Admin struct {
	Reseller
}

type AdminContext struct {
	ResellerContext
}

type convertAccount struct {
	Account string `json:"account,omitempty"`
	Creator string `json:"creator,omitempty"`
}

func (c *AdminContext) ConvertResellerToUser(username string, reseller string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "convert-reseller-to-user", convertAccount{Account: username, Creator: reseller}, nil); err != nil {
		return err
	}

	return nil
}

func (c *AdminContext) ConvertUserToReseller(username string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "convert-user-to-reseller", convertAccount{Account: username}, nil); err != nil {
		return err
	}

	return nil
}

func (c *AdminContext) DisableRedis() error {
	var response apiGenericResponseN

	if _, err := c.makeRequestNew(http.MethodPost, "redis/disable", nil, &response); err != nil {
		return err
	}

	return nil
}

func (c *AdminContext) EnableRedis() error {
	var response apiGenericResponseN

	if _, err := c.makeRequestNew(http.MethodPost, "redis/enable", nil, &response); err != nil {
		return err
	}

	return nil
}

// GetAllUsers (admin) returns an array of all users
func (c *AdminContext) GetAllUsers() ([]string, error) {
	var users []string

	if _, err := c.makeRequestOld(http.MethodGet, "API_SHOW_ALL_USERS", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetResellers (admin) returns an array of all resellers
func (c *AdminContext) GetResellers() ([]string, error) {
	var users []string

	if _, err := c.makeRequestOld(http.MethodGet, "API_SHOW_RESELLERS", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// TODO: finish implementation
func (c *AdminContext) GetResellersWithUsage() ([]string, error) {
	var users []string

	if _, err := c.makeRequestOld(http.MethodGet, "RESELLER_SHOW", nil, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *AdminContext) MoveUserToReseller(username string, reseller string) error {
	if _, err := c.makeRequestNew(http.MethodPost, "change-user-creator", convertAccount{Account: username, Creator: reseller}, nil); err != nil {
		return err
	}

	return nil
}

func (c *AdminContext) RestartDirectAdmin() error {
	var response apiGenericResponseN

	if _, err := c.makeRequestNew(http.MethodPost, "restart", nil, &response); err != nil {
		return err
	}

	return nil
}

func (c *AdminContext) UpdateDirectAdmin() error {
	var response apiGenericResponseN

	if _, err := c.makeRequestNew(http.MethodPost, "version/update", nil, &response); err != nil {
		return err
	}

	return nil
}
