package directadmin

import (
	"net/http"
)

type Plugin struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	MenuEntry struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Icon  string `json:"icon"`
	} `json:"menuEntry"`
}

// GetPlugins (user) returns the list of plugins in-use.
func (c *UserContext) GetPlugins() ([]*Plugin, error) {
	var plugins []*Plugin

	if _, err := c.makeRequestNew(http.MethodGet, "plugins/list", nil, &plugins); err != nil {
		return nil, err
	}

	return plugins, nil
}
