package directadmin

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateBackup (user) creates an account backup for the given domain, and the given items
func (c *UserContext) CreateBackup(domain string, backupItems ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "backup")
	body.Set("domain", domain)
	body.Set("form_version", "4")

	for index, backupItem := range backupItems {
		body.Set(fmt.Sprintf("select%d", index), backupItem)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "SITE_BACKUP", body, &response); err != nil {
		return err
	}

	if response.Success != "Backup creation added to queue" {
		return fmt.Errorf("failed to create backup: %v", response.Result)
	}

	return nil
}

// CreateBackupAllItems (user) wraps around CreateBackup and provides all available backup items
func (c *UserContext) CreateBackupAllItems(domain string) error {
	return c.CreateBackup(
		domain,
		"domain",
		"subdomain",
		"email",
		"email_data",
		"emailsettings",
		"forwarder",
		"autoresponder",
		"vacation",
		"list",
		"ftp",
		"ftpsettings",
		"database",
		"database_data",
		"trash",
	)
}

// GetBackups (user) returns an array of the session user's backups for the given domain
func (c *UserContext) GetBackups(domain string) ([]string, error) {
	var backups []string

	if _, err := c.makeRequestOld(http.MethodGet, "SITE_BACKUP?domain="+domain+"&ipp=50", nil, &backups); err != nil {
		return nil, err
	}

	return backups, nil
}

// RestoreBackup (user) restores an account backup for the given domain, and the given items
func (c *UserContext) RestoreBackup(domain string, backupFilename string, backupItems ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "restore")
	body.Set("domain", domain)
	body.Set("file", backupFilename)
	body.Set("form_version", "3")

	for index, backupItem := range backupItems {
		body.Set(fmt.Sprintf("select%d", index), backupItem)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "SITE_BACKUP", body, &response); err != nil {
		return err
	}

	if response.Success != "Restore will run in the background" {
		return fmt.Errorf("failed to restore backup: %v", response.Result)
	}

	return nil
}

// RestoreBackupAllItems (user) wraps around RestoreBackup and provides all available backup items
func (c *UserContext) RestoreBackupAllItems(domain string, backupFilename string) error {
	return c.RestoreBackup(
		domain,
		backupFilename,
		"domain",
		"subdomain",
		"email",
		"email_data",
		"emailsettings",
		"forwarder",
		"autoresponder",
		"vacation",
		"list",
		"ftp",
		"ftpsettings",
		"database",
		"database_data",
		"trash",
	)
}
