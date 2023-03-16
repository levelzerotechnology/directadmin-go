package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/google/go-querystring/query"
	"github.com/spf13/cast"
)

type ResellerPackage struct {
	Package
	OversellEnabled bool `json:"oversellEnabled" url:"oversellEnabled"`
	UserQuota       int  `json:"userQuota" url:"userQuota"`
}

// CreateResellerPackage (admin) creates the provided package
func (c *AdminContext) CreateResellerPackage(pack ResellerPackage) error {
	var response apiGenericResponse

	body, err := query.Values(pack.translate())
	if err != nil {
		return err
	}

	if _, err = c.api.makeRequest(http.MethodPost, "MANAGE_RESELLER_PACKAGES?add=yes", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Saved" {
		return fmt.Errorf("failed to create package: %v", response.Result)
	}

	return nil
}

// DeleteResellerPackages (admin) deletes all the specified packs for the session user
func (c *AdminContext) DeleteResellerPackages(packs ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("delete", "yes")

	for index, pack := range packs {
		body.Set("delete"+cast.ToString(index), pack)
		body.Set("select"+cast.ToString(index), pack)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "MANAGE_USER_PACKAGES", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Deleted" {
		return fmt.Errorf("failed to delete pack: %v", response.Result)
	}

	return nil
}

// GetResellerPackage (admin) returns the single specified package
func (c *AdminContext) GetResellerPackage(packageName string) (ResellerPackage, error) {
	var rawPack rawResellerPackage

	if _, err := c.api.makeRequest(http.MethodGet, "API_PACKAGES_USER?package="+packageName, c.credentials, nil, &rawPack); err != nil {
		return ResellerPackage{}, fmt.Errorf("failed to get package info for %v: %v", packageName, err)
	}

	rawPack.Name = packageName

	return rawPack.translate(), nil
}

// GetResellerPackages (admin) returns all packages belonging to the session user
func (c *AdminContext) GetResellerPackages() ([]ResellerPackage, error) {
	var packageList []string
	var packages []ResellerPackage

	if _, err := c.api.makeRequest(http.MethodGet, "API_PACKAGES_USER", c.credentials, nil, &packageList); err != nil {
		return nil, err
	}

	var errs []error
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(packageList))

	for _, packageName := range packageList {
		// convert to local variable to prevent variable overwrite
		packageToProcess := packageName

		go func(packageName string) {
			defer wg.Done()

			pack, err := c.GetResellerPackage(packageName)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			packages = append(packages, pack)
			mu.Unlock()
		}(packageToProcess)
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

	if len(packages) == 0 {
		return nil, errors.New("no packages could be found")
	}

	return packages, nil
}

// RenameResellerPackage (admin) renames the provided package
func (c *AdminContext) RenameResellerPackage(oldPackName string, newPackName string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("old_package", oldPackName)
	body.Set("new_package", newPackName)

	if _, err := c.api.makeRequest(http.MethodPost, "MANAGE_USER_PACKAGES?action=rename", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Saved" {
		return fmt.Errorf("failed to rename package: %v", response.Result)
	}

	return nil
}

// UpdateResellerPackage (admin) accepts a Package object and updates the version on DA with it
func (c *AdminContext) UpdateResellerPackage(pack ResellerPackage) error {
	// DA's update functionality is virtually identical to create, so we'll just use that
	return c.CreateResellerPackage(pack)
}
