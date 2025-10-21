package kiali

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ServicesList returns the list of services across specified namespaces.
func (k *Kiali) ServicesList(ctx context.Context, authHeader string, namespaces string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/clusters/services?health=true&istioResources=true&rateInterval=60s&onlyDefinitions=false"
	if namespaces != "" {
		endpoint += "&namespaces=" + url.QueryEscape(namespaces)
	}

	return k.executeRequest(ctx, authHeader, endpoint)
}

// ServiceDetails returns the details for a specific service in a namespace.
func (k *Kiali) ServiceDetails(ctx context.Context, authHeader string, namespace string, service string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if service == "" {
		return "", fmt.Errorf("service name is required")
	}
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/services/%s?validate=true&rateInterval=60s",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(service))

	return k.executeRequest(ctx, authHeader, endpoint)
}

// ServiceMetrics returns the metrics for a specific service in a namespace.
// Parameters:
//   - namespace: the namespace containing the service
//   - service: the name of the service
//   - queryParams: optional query parameters map for filtering metrics (e.g., "duration", "step", "rateInterval", "direction", "reporter", "filters[]", "byLabels[]", etc.)
func (k *Kiali) ServiceMetrics(ctx context.Context, authHeader string, namespace string, service string, queryParams map[string]string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if service == "" {
		return "", fmt.Errorf("service name is required")
	}

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/services/%s/metrics",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(service))

	// Add query parameters if provided
	if len(queryParams) > 0 {
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", err
		}
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		endpoint = u.String()
	}

	return k.executeRequest(ctx, authHeader, endpoint)
}
