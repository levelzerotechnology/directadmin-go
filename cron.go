package directadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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
