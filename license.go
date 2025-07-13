package directadmin

import (
	"fmt"
	"net/http"
	"time"
)

type License struct {
	Expires time.Time `json:"expires"`
	LID     int       `json:"lid"`
	Limits  struct {
		Legacy               bool `json:"legacy"`
		MaxAdminsOrResellers int  `json:"maxAdminsOrResellers"`
		MaxDomains           int  `json:"maxDomains"`
		MaxUsers             int  `json:"maxUsers"`
		OnlyVPS              bool `json:"onlyVPS"`
		ProPack              bool `json:"proPack"`
		Trial                bool `json:"trial"`
	} `json:"limits"`
	Name  string `json:"name"`
	PID   int    `json:"pid"`
	Type  string `json:"type"`
	UID   int    `json:"uid"`
	Usage struct {
		AdminsOrResellers int `json:"adminsOrResellers"`
		Domains           int `json:"domains"`
		Users             int `json:"users"`
	} `json:"usage"`
}

func (c *AdminContext) GetLicense() (*License, error) {
	var license License

	if _, err := c.makeRequestNew(http.MethodGet, "license", nil, &license); err != nil {
		return nil, fmt.Errorf("failed to get license: %w", err)
	}

	return &license, nil
}
