package kiali

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// AppTraces returns distributed tracing data for a specific app in a namespace.
// Parameters:
//   - namespace: the namespace containing the app
//   - app: the name of the app
//   - queryParams: optional query parameters map for filtering traces (e.g., "startMicros", "endMicros", "limit", "minDuration", "tags", "clusterName")
func (k *Kiali) AppTraces(ctx context.Context, namespace string, app string, queryParams map[string]string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/apps/%s/traces",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(app))

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

	return k.executeRequest(ctx, endpoint)
}

// ServiceTraces returns distributed tracing data for a specific service in a namespace.
// Parameters:
//   - namespace: the namespace containing the service
//   - service: the name of the service
//   - queryParams: optional query parameters map for filtering traces (e.g., "startMicros", "endMicros", "limit", "minDuration", "tags", "clusterName")
func (k *Kiali) ServiceTraces(ctx context.Context, namespace string, service string, queryParams map[string]string) (string, error) {
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

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/services/%s/traces",
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

	return k.executeRequest(ctx, endpoint)
}

// WorkloadTraces returns distributed tracing data for a specific workload in a namespace.
// Parameters:
//   - namespace: the namespace containing the workload
//   - workload: the name of the workload
//   - queryParams: optional query parameters map for filtering traces (e.g., "startMicros", "endMicros", "limit", "minDuration", "tags", "clusterName")
func (k *Kiali) WorkloadTraces(ctx context.Context, namespace string, workload string, queryParams map[string]string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if workload == "" {
		return "", fmt.Errorf("workload name is required")
	}

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/workloads/%s/traces",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(workload))

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

	return k.executeRequest(ctx, endpoint)
}
