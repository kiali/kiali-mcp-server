package kiali

import (
	"context"
	"net/url"
	"strings"
)

// Health returns health status for apps, workloads, and services across namespaces.
// Parameters:
//   - namespaces: comma-separated list of namespaces (optional, if empty returns health for all accessible namespaces)
//   - queryParams: optional query parameters map for filtering health data (e.g., "type", "rateInterval", "queryTime")
//   - type: health type - "app", "service", or "workload" (default: "app")
//   - rateInterval: rate interval for fetching error rate (default: "10m")
//   - queryTime: Unix timestamp for the prometheus query (optional)
func (k *Kiali) Health(ctx context.Context, authHeader string, namespaces string, queryParams map[string]string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/api/clusters/health"

	// Build query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()

	// Add namespaces if provided
	if namespaces != "" {
		q.Set("namespaces", namespaces)
	}

	// Add optional query parameters
	if len(queryParams) > 0 {
		for key, value := range queryParams {
			q.Set(key, value)
		}
	}

	u.RawQuery = q.Encode()
	endpoint = u.String()

	return k.executeRequest(ctx, authHeader, endpoint)
}
