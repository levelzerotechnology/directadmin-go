package directadmin

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CloudLinuxGetUsageCharts retrieves the CloudLinux usage charts for the given period and ID.
//
// period: e.g., "1h", "1d", "1w", "1m"
// id: the user ID or LVE ID (e.g., "1000001025")
func (c *UserContext) CloudLinuxGetUsageCharts(period string, id string) ([]*CloudLinuxChartData, error) {
	rawChart, err := c.cloudLinuxGetUsageCharts(period, id, "svg")
	if err != nil {
		return nil, err
	}

	// There is a non-SVG endpoint; however, it's significantly slower.
	charts, err := parseCloudLinuxSVG(rawChart)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG chart: %w", err)
	}

	return charts, nil
}

// CloudLinuxGetUsageChartsAsImage retrieves the CloudLinux usage charts for the given period and ID.
//
// period: e.g., "1h", "1d", "1w", "1m"
// id: the user ID or LVE ID (e.g., "1000001025")
// format: SVG or PNG
func (c *UserContext) CloudLinuxGetUsageChartsAsImage(period string, id string, format string) (string, error) {
	format = strings.ToLower(format)
	if format != "png" && format != "svg" {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	chart, err := c.cloudLinuxGetUsageCharts(period, id, format)
	if err != nil {
		return "", err
	}

	return chart, nil
}

// CloudLinuxGetUsageCharts retrieves the CloudLinux usage charts for the given period and ID.
//
// period: e.g., "1h", "1d", "1w", "1m"
// id: the user ID or LVE ID (e.g., "1000001025")
func (c *UserContext) cloudLinuxGetUsageCharts(period string, id string, format string) (string, error) {
	if err := c.CreateSession(); err != nil {
		return "", fmt.Errorf("failed to create user session: %w", err)
	}

	csrfToken, err := c.cloudlinuxCreateCSRFToken()
	if err != nil {
		return "", fmt.Errorf("failed to create CSRF token: %w", err)
	}

	body := url.Values{}
	body.Set("command", "cloudlinux-charts")
	body.Set("params[period]", period)
	body.Set("params[format]", format)
	body.Set("params[id]", id)
	body.Set("csrftoken", csrfToken)

	var resp struct {
		Chart  string `json:"chart"`
		Result string `json:"result"`
	}

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS_RESELLER/lvemanager_spa/index.raw?c=send-request", body, &resp); err != nil {
		return "", err
	}

	if resp.Result != "success" {
		return "", fmt.Errorf("failed to retrieve charts: %s", resp.Result)
	}

	return resp.Chart, nil
}

type CloudLinuxUser struct {
	CageFS string `json:"cageFS"`
	Domain string `json:"domain"`
	ID     int    `json:"id"`
	Limits struct {
		CPU struct {
			All string `json:"all"`
		} `json:"cpu"`
		EP string `json:"ep"`
		IO struct {
			All string `json:"all"`
		} `json:"io"`
		IOPS  string `json:"iops"`
		NProc string `json:"nproc"`
		PMem  string `json:"pmem"`
		VMem  string `json:"vmem"`
	} `json:"limits"`
	Package  string `json:"package"`
	Username string `json:"username"`
}

// CloudLinuxGetUsers retrieves all accessible CLoudLinux users with their resource limits.
func (c *UserContext) CloudLinuxGetUsers() ([]*CloudLinuxUser, error) {
	if err := c.CreateSession(); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	csrfToken, err := c.cloudlinuxCreateCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create CSRF token: %w", err)
	}

	body := url.Values{}
	body.Set("command", "cloudlinux-limits")
	body.Set("method", "get")
	body.Set("csrftoken", csrfToken)

	var resp struct {
		Result string            `json:"result"`
		Users  []*CloudLinuxUser `json:"users"`
	}

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS_RESELLER/lvemanager_spa/index.raw?c=send-request", body, &resp); err != nil {
		return nil, err
	}

	if resp.Result != "success" {
		return nil, fmt.Errorf("failed to retrieve users: %s", resp.Result)
	}

	return resp.Users, nil
}

// cloudlinuxCreateCSRFToken creates a CSRF token for the CloudLinux plugin if one doesn't already exist.
func (c *UserContext) cloudlinuxCreateCSRFToken() (string, error) {
	const endpoint = "PLUGINS_RESELLER/lvemanager_spa/index.raw"

	csrfToken := c.getCSRFToken(c.getRequestURLOld(endpoint))
	if csrfToken != "" {
		return csrfToken, nil
	}

	// Retrieve CSRF token.
	if _, err := c.makeRequestOld(http.MethodGet, "PLUGINS_RESELLER/lvemanager_spa/index.raw?a=cookie", nil, nil); err != nil {
		return "", err
	}

	return c.getCSRFToken(c.getRequestURLOld(endpoint)), nil
}

// dataPoint represents a single time-series data point with multiple series values.
type dataPoint struct {
	Fault     bool               `json:"fault,omitempty"`
	Timestamp string             `json:"timestamp"`
	Values    map[string]float64 `json:"values"`
}

// CloudLinuxChartData represents the parsed data for a single chart.
type CloudLinuxChartData struct {
	ChartTitle  string      `json:"chartTitle"`
	DataPoints  []dataPoint `json:"dataPoints"`
	FaultEvents []string    `json:"faultEvents,omitempty"`
	Unit        string      `json:"unit"`
}

// parseCloudLinuxSVG extracts time-series data from a CloudLinux SVG chart.
func parseCloudLinuxSVG(svgContent string) ([]*CloudLinuxChartData, error) {
	var charts []*CloudLinuxChartData

	// Regex to find individual chart SVG blocks.
	chartBlockRegex := regexp.MustCompile(`<svg[^>]*id="(svg-\d+)"[^>]*>([\s\S]*?)</svg>`)
	chartBlocks := chartBlockRegex.FindAllStringSubmatch(svgContent, -1)

	// Regex to find the chart title within a block.
	titleRegex := regexp.MustCompile(`<text[^>]*x="55"[^>]*y="10"[^>]*>([^<]+)</text>`)

	// Regex to find legend entries within a block.
	legendRegex := regexp.MustCompile(`<rect[^>]*fill="([^"]+)"[^>]*/>\s*<text[^>]*>([^<]+)</text>`)

	// Regex to extract data from line elements with onmousemove handlers.
	lineRegex := regexp.MustCompile(`<line[^>]*onmousemove="show_tip\(evt,\s*'(svg-\d+)',\s*[\d.]+,\s*[\d.]+,\s*[\d.]+,\s*[\d.]+,\s*'([^']+)',\s*'([^']+)',\s*'([^']+)',\s*'([^']+)'\)"[^>]*stroke="([^"]+)"[^>]*/>`)

	// Regex to extract fault lines (vertical lines without onmousemove, typically with stroke="aquamarine").
	// These are lines where x1 == x2 (vertical) within a chart's clip-path group.
	faultLineRegex := regexp.MustCompile(`<line[^>]*stroke="aquamarine"[^>]*x1="([\d.]+)"[^>]*x2="([\d.]+)"[^>]*y1="([\d.]+)"[^>]*y2="([\d.]+)"[^>]*/>`)

	// Build per-chart legends and titles.
	chartTitles := make(map[string]string)
	chartLegends := make(map[string]map[string]string) // svgID -> color -> name.
	chartContents := make(map[string]string)           // svgID -> block content.

	for _, block := range chartBlocks {
		if len(block) < 3 {
			continue
		}
		svgID := block[1]
		content := block[2]
		chartContents[svgID] = content

		// Extract title.
		if titleMatch := titleRegex.FindStringSubmatch(content); len(titleMatch) >= 2 {
			chartTitles[svgID] = titleMatch[1]
		}

		// Extract legend for this chart.
		chartLegends[svgID] = make(map[string]string)
		legendMatches := legendRegex.FindAllStringSubmatch(content, -1)
		for _, match := range legendMatches {
			if len(match) >= 3 {
				chartLegends[svgID][match[1]] = match[2]
			}
		}
	}

	// Extract all data points from lines with onmousemove handlers.
	matches := lineRegex.FindAllStringSubmatch(svgContent, -1)

	// Group data points by chart ID, then by timestamp, with series values.
	chartDataMap := make(map[string]map[string]map[string]float64)
	chartUnits := make(map[string]string)

	for _, match := range matches {
		if len(match) < 7 {
			continue
		}

		svgID := match[1]
		t1, v1 := match[2], match[3]
		t2, v2 := match[4], match[5]
		color := match[6]

		// Map color to series name using this chart's legend.
		seriesName := color
		if legend, ok := chartLegends[svgID]; ok {
			if name, ok := legend[color]; ok {
				seriesName = name
			}
		}

		if chartDataMap[svgID] == nil {
			chartDataMap[svgID] = make(map[string]map[string]float64)
		}

		// Store unit for this chart.
		if chartUnits[svgID] == "" {
			chartUnits[svgID] = extractUnitFromCloudLinuxChart(v1)
		}

		ts1, _ := parseCloudLinuxTimestamp(t1)
		if chartDataMap[svgID][ts1] == nil {
			chartDataMap[svgID][ts1] = make(map[string]float64)
		}
		chartDataMap[svgID][ts1][seriesName] = parseCloudLinuxChartValue(v1)

		ts2, _ := parseCloudLinuxTimestamp(t2)
		if chartDataMap[svgID][ts2] == nil {
			chartDataMap[svgID][ts2] = make(map[string]float64)
		}
		chartDataMap[svgID][ts2][seriesName] = parseCloudLinuxChartValue(v2)
	}

	// Extract fault events per chart.
	chartFaultTimestamps := make(map[string]map[string]bool)
	for svgID, content := range chartContents {
		faultMatches := faultLineRegex.FindAllStringSubmatch(content, -1)
		for _, match := range faultMatches {
			if len(match) < 5 {
				continue
			}
			x1, _ := strconv.ParseFloat(match[1], 64)
			x2, _ := strconv.ParseFloat(match[2], 64)

			// Only process vertical lines (fault markers have x1 == x2).
			if x1 != x2 {
				continue
			}

			// Find the nearest timestamp for this x position.
			timestamp := findTimestampForXPosition(svgID, x1, svgContent)

			// Mark this timestamp as having a fault.
			if chartFaultTimestamps[svgID] == nil {
				chartFaultTimestamps[svgID] = make(map[string]bool)
			}

			chartFaultTimestamps[svgID][timestamp] = true
		}
	}

	// Convert maps to sorted slices.
	for svgID, timestampMap := range chartDataMap {
		chart := &CloudLinuxChartData{
			ChartTitle: chartTitles[svgID],
			Unit:       chartUnits[svgID],
			DataPoints: make([]dataPoint, 0, len(timestampMap)),
		}

		for timestamp, values := range timestampMap {
			chart.DataPoints = append(chart.DataPoints, dataPoint{
				Timestamp: timestamp,
				Values:    values,
				Fault:     chartFaultTimestamps[svgID][timestamp],
			})
		}

		// Sort data points by timestamp.
		sort.Slice(chart.DataPoints, func(i, j int) bool {
			return chart.DataPoints[i].Timestamp < chart.DataPoints[j].Timestamp
		})

		charts = append(charts, chart)
	}

	return charts, nil
}

// findTimestampForXPosition finds the timestamp corresponding to a given x position in the SVG.
// It does this by finding the nearest data line and interpolating from its timestamps.
func findTimestampForXPosition(svgID string, xPos float64, svgContent string) string {
	// Regex to extract x positions and timestamps from lines in this chart.
	lineRegex := regexp.MustCompile(`<line[^>]*onmousemove="show_tip\(evt,\s*'` + regexp.QuoteMeta(svgID) + `',\s*([\d.]+),\s*[\d.]+,\s*([\d.]+),\s*[\d.]+,\s*'([^']+)',\s*'[^']+',\s*'([^']+)',\s*'[^']+'\)"[^>]*/>`)

	matches := lineRegex.FindAllStringSubmatch(svgContent, -1)

	var closestTimestamp string
	closestDistance := -1.0

	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		x1, _ := strconv.ParseFloat(match[1], 64)
		x2, _ := strconv.ParseFloat(match[2], 64)
		t1 := match[3]
		t2 := match[4]

		// Check distance to x1.
		dist1 := abs(x1 - xPos)
		if closestDistance < 0 || dist1 < closestDistance {
			closestDistance = dist1
			ts, _ := parseCloudLinuxTimestamp(t1)
			closestTimestamp = ts
		}

		// Check distance to x2.
		dist2 := abs(x2 - xPos)
		if dist2 < closestDistance {
			closestDistance = dist2
			ts, _ := parseCloudLinuxTimestamp(t2)
			closestTimestamp = ts
		}
	}

	return closestTimestamp
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// extractUnitFromCloudLinuxChart extracts the unit suffix from a value string.
func extractUnitFromCloudLinuxChart(rawValue string) string {
	rawValue = strings.TrimSpace(rawValue)
	for _, suffix := range []string{"%", "MB", "GB", "KB", "KB/s", "MB/s", "ms", "s", "B"} {
		if strings.HasSuffix(rawValue, suffix) {
			return suffix
		}
	}
	return ""
}

// parseCloudLinuxChartValue extracts numeric value from strings like "113%", "57MB", "1.5GB".
func parseCloudLinuxChartValue(rawValue string) float64 {
	// Remove common suffixes.
	cleaned := strings.TrimSpace(rawValue)
	cleaned = strings.TrimSuffix(cleaned, "%")
	cleaned = strings.TrimSuffix(cleaned, "MB")
	cleaned = strings.TrimSuffix(cleaned, "GB")
	cleaned = strings.TrimSuffix(cleaned, "KB")
	cleaned = strings.TrimSuffix(cleaned, "B")
	cleaned = strings.TrimSuffix(cleaned, "ms")
	cleaned = strings.TrimSuffix(cleaned, "s")

	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}

	return val
}

// parseCloudLinuxTimestamp converts CloudLinux timestamp format to RFC3339.
// Input format: "Jan-29 04:28PM"
// Output format: "2025-01-29T16:28:00Z"
func parseCloudLinuxTimestamp(ts string) (string, error) {
	// CloudLinux format: "Jan-29 04:28PM"
	// Go parse layout: "Jan-02 03:04PM"
	parsed, err := time.Parse("Jan-02 03:04PM", ts)
	if err != nil {
		return ts, err
	}

	// The SVG doesn't include the year, so we assume the current year.
	now := time.Now().UTC()
	parsed = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), 0, 0, time.UTC)

	// Handle year boundary (e.g., if it's January and the timestamp is from December).
	if parsed.After(now.AddDate(0, 0, 1)) {
		parsed = parsed.AddDate(-1, 0, 0)
	}

	return parsed.Format(time.RFC3339), nil
}
