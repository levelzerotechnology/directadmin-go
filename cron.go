package directadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// cronLinePattern splits a crontab line into its five schedule fields and the command, which
// keeps its own spacing.
var cronLinePattern = regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.+)$`)

// CronJob is one line of a user's crontab.
type CronJob struct {
	ID         string
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
	Command    string
}

// GetCronJobs (user) returns the user's cron jobs in crontab order.
func (c *UserContext) GetCronJobs() ([]CronJob, error) {
	// Keyed by job ID ("000", "001"), alongside the crontab's MAILTO and PATH variables.
	var raw map[string]string

	if _, err := c.makeRequestOld(http.MethodGet, "API_CRON_JOBS", nil, &raw); err != nil {
		return nil, fmt.Errorf("failed to get cron jobs: %w", err)
	}

	jobs := make([]CronJob, 0, len(raw))

	for id, line := range raw {
		if _, err := strconv.Atoi(id); err != nil {
			continue
		}

		fields := cronLinePattern.FindStringSubmatch(line)
		if fields == nil {
			return nil, fmt.Errorf("failed to parse cron job %s: %q", id, line)
		}

		jobs = append(jobs, CronJob{
			ID:         id,
			Minute:     fields[1],
			Hour:       fields[2],
			DayOfMonth: fields[3],
			Month:      fields[4],
			DayOfWeek:  fields[5],
			Command:    fields[6],
		})
	}

	sort.Slice(jobs, func(i, j int) bool { return jobs[i].ID < jobs[j].ID })

	return jobs, nil
}

// CreateCronJob (user) adds job to the user's crontab. job.ID is ignored.
//
// Sent to the documented CMD_API_CRON_JOBS create action. Callers should read the crontab back
// afterwards, as an unrecognised request is answered like a list request without an error.
func (c *UserContext) CreateCronJob(job CronJob) error {
	for _, field := range []string{job.Minute, job.Hour, job.DayOfMonth, job.Month, job.DayOfWeek, job.Command} {
		if field == "" || strings.ContainsAny(field, "\r\n") {
			return errors.New("failed to create cron job: every field must be set and single-line")
		}
	}

	body := url.Values{}
	body.Set("minute", job.Minute)
	body.Set("hour", job.Hour)
	body.Set("dayofmonth", job.DayOfMonth)
	body.Set("month", job.Month)
	body.Set("dayofweek", job.DayOfWeek)
	body.Set("command", job.Command)

	if _, err := c.makeRequestOld(http.MethodPost, "API_CRON_JOBS?action=create", body, nil); err != nil {
		return fmt.Errorf("failed to create cron job: %w", err)
	}

	return nil
}

// UpdateCronJob (user) replaces the schedule and command of the job with job.ID.
//
// Sent as JSON to CMD_CRON_JOBS, the way the Evolution skin saves one. The form-encoded
// CMD_API_CRON_JOBS save is silently treated as a list request.
func (c *UserContext) UpdateCronJob(job CronJob) error {
	body, err := json.Marshal(map[string]string{
		"id":         job.ID,
		"minute":     job.Minute,
		"hour":       job.Hour,
		"dayofmonth": job.DayOfMonth,
		"month":      job.Month,
		"dayofweek":  job.DayOfWeek,
		"command":    job.Command,
		"json":       "yes",
		"save":       "yes",
	})
	if err != nil {
		return err
	}

	req, err := c.newRequest(context.Background(), http.MethodPost, c.getRequestURLOld("CRON_JOBS?json=yes"), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	_, resp, err := c.makeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to update cron job %s: %w", job.ID, err)
	}

	var response apiGenericResponse
	if err = json.Unmarshal(resp, &response); err == nil && response.Error != "" {
		return fmt.Errorf("failed to update cron job %s: %s", job.ID, errorOrResult(response))
	}

	return nil
}
