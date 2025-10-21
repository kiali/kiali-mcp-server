package kiali

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kiali/kiali-mcp-server/pkg/config"
	internalk8s "github.com/kiali/kiali-mcp-server/pkg/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Kiali struct {
	manager *Manager
}

type Manager struct {
	cfg          *rest.Config
	staticConfig *config.StaticConfig
}

// NewFromConfig creates a new Kiali client backed by the given static configuration.
func NewFromConfig(cfg *config.StaticConfig) *Kiali {
	return &Kiali{manager: &Manager{staticConfig: cfg}}
}

// validateAndGetBaseURL validates the Kiali client configuration and returns the base URL.
func (k *Kiali) validateAndGetBaseURL() (string, error) {
	if k == nil || k.manager == nil || k.manager.staticConfig == nil {
		return "", fmt.Errorf("kiali client not initialized")
	}
	baseURL := strings.TrimSpace(k.manager.staticConfig.KialiServerURL)
	if baseURL == "" {
		return "", fmt.Errorf("kiali server URL not configured")
	}
	return baseURL, nil
}

// createHTTPClient creates an HTTP client with appropriate TLS configuration.
func (k *Kiali) createHTTPClient() *http.Client {
	transport := &http.Transport{}
	if k.manager.staticConfig.KialiInsecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // allowed via configuration
	}
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}

// executeRequest executes an HTTP request and handles common error scenarios.
func (k *Kiali) executeRequest(ctx context.Context, authHeader, endpoint string) (string, error) {
	klog.V(0).Infof("kiali API call: %s", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	} else if k.manager.staticConfig.RequireOAuth {
		return "", fmt.Errorf("authorization token required for Kiali call")
	}

	client := k.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(body) > 0 {
			return "", fmt.Errorf("kiali API error: %s", strings.TrimSpace(string(body)))
		}
		return "", fmt.Errorf("kiali API error: status %d", resp.StatusCode)
	}
	return string(body), nil
}

// executeRequestWithBody executes an HTTP request with a body and handles common error scenarios.
func (k *Kiali) executeRequestWithBody(ctx context.Context, authHeader, method, endpoint, contentType string, body io.Reader) (string, error) {
	klog.V(0).Infof("kiali API call: %s %s", method, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return "", err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	} else if k.manager.staticConfig.RequireOAuth {
		return "", fmt.Errorf("authorization token required for Kiali call")
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := k.createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(respBody) > 0 {
			return "", fmt.Errorf("kiali API error: %s", strings.TrimSpace(string(respBody)))
		}
		return "", fmt.Errorf("kiali API error: status %d", resp.StatusCode)
	}
	return string(respBody), nil
}

// ValidationsList calls the Kiali validations API using the provided Authorization header value.
// The authHeader must be the full header value (for example: "Bearer <token>").
// `namespaces` may contain zero, one or many namespaces. If empty, returns validations from all namespaces.
func (k *Kiali) ValidationsList(ctx context.Context, authHeader string, namespaces []string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/istio/validations"

	// Add namespaces query parameter if any provided
	cleaned := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			cleaned = append(cleaned, ns)
		}
	}
	if len(cleaned) > 0 {
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", err
		}
		q := u.Query()
		q.Set("namespaces", strings.Join(cleaned, ","))
		u.RawQuery = q.Encode()
		endpoint = u.String()
	}

	return k.executeRequest(ctx, authHeader, endpoint)
}

// MeshGraph calls the Kiali graph API using the provided Authorization header value.
// `namespaces` may contain zero, one or many namespaces. If empty, the API may return an empty graph
// or the server default, depending on Kiali configuration.
func (k *Kiali) MeshGraph(ctx context.Context, authHeader string, namespaces []string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/namespaces/graph"

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	// Static graph parameters per requirements
	q.Set("duration", "60s")
	q.Set("graphType", "versionedApp")
	q.Set("includeIdleEdges", "false")
	q.Set("injectServiceNodes", "true")
	q.Set("boxBy", "cluster,namespace,app")
	q.Set("ambientTraffic", "none")
	q.Set("appenders", "deadNode,istio,serviceEntry,meshCheck,workloadEntry,health")
	q.Set("rateGrpc", "requests")
	q.Set("rateHttp", "requests")
	q.Set("rateTcp", "sent")
	// Optional namespaces param
	cleaned := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			cleaned = append(cleaned, ns)
		}
	}
	if len(cleaned) > 0 {
		q.Set("namespaces", strings.Join(cleaned, ","))
	}
	u.RawQuery = q.Encode()
	endpoint = u.String()

	return k.executeRequest(ctx, authHeader, endpoint)
}

// MeshStatus calls the Kiali mesh graph API to get the status of mesh components.
// This returns information about mesh components like Istio, Kiali, Grafana, Prometheus
// and their interactions, versions, and health status.
func (k *Kiali) MeshStatus(ctx context.Context, authHeader string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/mesh/graph"

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("includeGateways", "false")
	q.Set("includeWaypoints", "false")
	u.RawQuery = q.Encode()
	endpoint = u.String()

	return k.executeRequest(ctx, authHeader, endpoint)
}

// MeshNamespaces calls the Kiali namespaces API using the provided Authorization header value.
// Returns all namespaces in the mesh that the user has access to.
func (k *Kiali) MeshNamespaces(ctx context.Context, authHeader string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/namespaces"

	return k.executeRequest(ctx, authHeader, endpoint)
}

// WorkloadsList returns the list of workloads across specified namespaces.
func (k *Kiali) WorkloadsList(ctx context.Context, authHeader string, namespaces string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/clusters/workloads?health=true&istioResources=true&rateInterval=60s"
	if namespaces != "" {
		endpoint += "&namespaces=" + url.QueryEscape(namespaces)
	}

	return k.executeRequest(ctx, authHeader, endpoint)
}

// WorkloadDetails returns the details for a specific workload in a namespace.
func (k *Kiali) WorkloadDetails(ctx context.Context, authHeader string, namespace string, workload string) (string, error) {
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
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/workloads/%s?validate=true&rateInterval=60s&health=true",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(workload))

	return k.executeRequest(ctx, authHeader, endpoint)
}

// WorkloadMetrics returns the metrics for a specific workload in a namespace.
// Parameters:
//   - namespace: the namespace containing the workload
//   - workload: the name of the workload
//   - queryParams: optional query parameters map for filtering metrics (e.g., "duration", "step", "rateInterval", "direction", "reporter", "filters[]", "byLabels[]", etc.)
func (k *Kiali) WorkloadMetrics(ctx context.Context, authHeader string, namespace string, workload string, queryParams map[string]string) (string, error) {
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

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/workloads/%s/metrics",
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

	return k.executeRequest(ctx, authHeader, endpoint)
}

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

// IstioConfig calls the Kiali Istio config API to get all Istio objects in the mesh.
// Returns the full YAML resources and additional details about each object.
func (k *Kiali) IstioConfig(ctx context.Context, authHeader string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/istio/config?validate=true"

	return k.executeRequest(ctx, authHeader, endpoint)
}

// IstioObjectDetails returns detailed information about a specific Istio object.
// Parameters:
//   - namespace: the namespace containing the Istio object
//   - group: the API group (e.g., "networking.istio.io", "gateway.networking.k8s.io")
//   - version: the API version (e.g., "v1", "v1beta1")
//   - kind: the resource kind (e.g., "DestinationRule", "VirtualService", "HTTPRoute")
//   - name: the name of the resource
func (k *Kiali) IstioObjectDetails(ctx context.Context, authHeader string, namespace, group, version, kind, name string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if group == "" {
		return "", fmt.Errorf("group is required")
	}
	if version == "" {
		return "", fmt.Errorf("version is required")
	}
	if kind == "" {
		return "", fmt.Errorf("kind is required")
	}
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/istio/%s/%s/%s/%s?validate=true&help=true",
		strings.TrimRight(baseURL, "/"),
		url.PathEscape(namespace),
		url.PathEscape(group),
		url.PathEscape(version),
		url.PathEscape(kind),
		url.PathEscape(name))

	return k.executeRequest(ctx, authHeader, endpoint)
}

// IstioObjectPatch patches an existing Istio object using PATCH method.
// Parameters:
//   - namespace: the namespace containing the Istio object
//   - group: the API group (e.g., "networking.istio.io", "gateway.networking.k8s.io")
//   - version: the API version (e.g., "v1", "v1beta1")
//   - kind: the resource kind (e.g., "DestinationRule", "VirtualService", "HTTPRoute")
//   - name: the name of the resource
//   - jsonPatch: the JSON patch data to apply
func (k *Kiali) IstioObjectPatch(ctx context.Context, authHeader string, namespace, group, version, kind, name, jsonPatch string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if group == "" {
		return "", fmt.Errorf("group is required")
	}
	if version == "" {
		return "", fmt.Errorf("version is required")
	}
	if kind == "" {
		return "", fmt.Errorf("kind is required")
	}
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	if jsonPatch == "" {
		return "", fmt.Errorf("json patch data is required")
	}
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/istio/%s/%s/%s/%s",
		strings.TrimRight(baseURL, "/"),
		url.PathEscape(namespace),
		url.PathEscape(group),
		url.PathEscape(version),
		url.PathEscape(kind),
		url.PathEscape(name))

	return k.executeRequestWithBody(ctx, authHeader, http.MethodPatch, endpoint, "application/json", strings.NewReader(jsonPatch))
}

// IstioObjectCreate creates a new Istio object using POST method.
// Parameters:
//   - namespace: the namespace where the Istio object will be created
//   - group: the API group (e.g., "networking.istio.io", "gateway.networking.k8s.io")
//   - version: the API version (e.g., "v1", "v1beta1")
//   - kind: the resource kind (e.g., "DestinationRule", "VirtualService", "HTTPRoute")
//   - jsonData: the JSON data for the new object
func (k *Kiali) IstioObjectCreate(ctx context.Context, authHeader string, namespace, group, version, kind, jsonData string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if group == "" {
		return "", fmt.Errorf("group is required")
	}
	if version == "" {
		return "", fmt.Errorf("version is required")
	}
	if kind == "" {
		return "", fmt.Errorf("kind is required")
	}
	if jsonData == "" {
		return "", fmt.Errorf("json data is required")
	}
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/istio/%s/%s/%s",
		strings.TrimRight(baseURL, "/"),
		url.PathEscape(namespace),
		url.PathEscape(group),
		url.PathEscape(version),
		url.PathEscape(kind))

	return k.executeRequestWithBody(ctx, authHeader, http.MethodPost, endpoint, "application/json", strings.NewReader(jsonData))
}

// IstioObjectDelete deletes an existing Istio object using DELETE method.
// Parameters:
//   - namespace: the namespace containing the Istio object
//   - group: the API group (e.g., "networking.istio.io", "gateway.networking.k8s.io")
//   - version: the API version (e.g., "v1", "v1beta1")
//   - kind: the resource kind (e.g., "DestinationRule", "VirtualService", "HTTPRoute", "Gateway")
//   - name: the name of the resource
func (k *Kiali) IstioObjectDelete(ctx context.Context, authHeader string, namespace, group, version, kind, name string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if group == "" {
		return "", fmt.Errorf("group is required")
	}
	if version == "" {
		return "", fmt.Errorf("version is required")
	}
	if kind == "" {
		return "", fmt.Errorf("kind is required")
	}
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	endpoint := fmt.Sprintf("%s/api/namespaces/%s/istio/%s/%s/%s/%s",
		strings.TrimRight(baseURL, "/"),
		url.PathEscape(namespace),
		url.PathEscape(group),
		url.PathEscape(version),
		url.PathEscape(kind),
		url.PathEscape(name))

	return k.executeRequestWithBody(ctx, authHeader, http.MethodDelete, endpoint, "", nil)
}

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

// WorkloadLogs returns logs for a specific workload's pods in a namespace.
// This method first gets workload details to find associated pods, then retrieves logs for each pod.
// Parameters:
//   - namespace: the namespace containing the workload
//   - workload: the name of the workload
//   - container: container name (optional, will be auto-detected if not provided)
//   - service: service name (optional)
//   - duration: time duration (e.g., "5m", "1h") - optional
//   - logType: type of logs (app, proxy, ztunnel, waypoint) - optional
//   - sinceTime: Unix timestamp for start time - optional
//   - maxLines: maximum number of lines to return - optional
func (k *Kiali) WorkloadLogs(ctx context.Context, authHeader string, namespace string, workload string, container string, service string, duration string, logType string, sinceTime string, maxLines string) (string, error) {
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if workload == "" {
		return "", fmt.Errorf("workload name is required")
	}
	// Container is optional - will be auto-detected if not provided

	// First, get workload details to find associated pods
	workloadDetails, err := k.WorkloadDetails(ctx, authHeader, namespace, workload)
	if err != nil {
		return "", fmt.Errorf("failed to get workload details: %v", err)
	}

	// Parse the workload details JSON to extract pod names and containers
	var workloadData struct {
		Pods []struct {
			Name       string `json:"name"`
			Containers []struct {
				Name string `json:"name"`
			} `json:"containers"`
		} `json:"pods"`
	}

	if err := json.Unmarshal([]byte(workloadDetails), &workloadData); err != nil {
		return "", fmt.Errorf("failed to parse workload details: %v", err)
	}

	if len(workloadData.Pods) == 0 {
		return "", fmt.Errorf("no pods found for workload %s in namespace %s", workload, namespace)
	}

	// Collect logs from all pods
	var allLogs []string
	for _, pod := range workloadData.Pods {
		// Auto-detect container if not provided
		podContainer := container
		if podContainer == "" {
			// Find the main application container (not istio-proxy or istio-init)
			for _, c := range pod.Containers {
				if c.Name != "istio-proxy" && c.Name != "istio-init" {
					podContainer = c.Name
					break
				}
			}
			// If no app container found, use the first container
			if podContainer == "" && len(pod.Containers) > 0 {
				podContainer = pod.Containers[0].Name
			}
		}

		if podContainer == "" {
			allLogs = append(allLogs, fmt.Sprintf("Error: No container found for pod %s", pod.Name))
			continue
		}

		podLogs, err := k.PodLogs(ctx, authHeader, namespace, pod.Name, podContainer, workload, service, duration, logType, sinceTime, maxLines)
		if err != nil {
			// Log the error but continue with other pods
			allLogs = append(allLogs, fmt.Sprintf("Error getting logs for pod %s: %v", pod.Name, err))
			continue
		}
		if podLogs != "" {
			allLogs = append(allLogs, fmt.Sprintf("=== Pod: %s (Container: %s) ===\n%s", pod.Name, podContainer, podLogs))
		}
	}

	if len(allLogs) == 0 {
		return "", fmt.Errorf("no logs found for workload %s in namespace %s", workload, namespace)
	}

	return strings.Join(allLogs, "\n\n"), nil
}

// PodLogs returns logs for a specific pod using the Kiali API endpoint.
// Parameters:
//   - namespace: the namespace containing the pod
//   - podName: the name of the pod
//   - container: container name (optional, will be auto-detected if not provided)
//   - workload: workload name (optional)
//   - service: service name (optional)
//   - duration: time duration (e.g., "5m", "1h") - optional
//   - logType: type of logs (app, proxy, ztunnel, waypoint) - optional
//   - sinceTime: Unix timestamp for start time - optional
//   - maxLines: maximum number of lines to return - optional
func (k *Kiali) PodLogs(ctx context.Context, authHeader string, namespace string, podName string, container string, workload string, service string, duration string, logType string, sinceTime string, maxLines string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if podName == "" {
		return "", fmt.Errorf("pod name is required")
	}
	// Container is optional - will be auto-detected if not provided
	podContainer := container
	if podContainer == "" {
		// Get pod details to find containers
		podDetails, err := k.executeRequest(ctx, authHeader, fmt.Sprintf("%s/api/namespaces/%s/pods/%s", strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(podName)))
		if err != nil {
			return "", fmt.Errorf("failed to get pod details: %v", err)
		}

		// Parse pod details to extract container names
		var podData struct {
			Containers []struct {
				Name string `json:"name"`
			} `json:"containers"`
		}

		if err := json.Unmarshal([]byte(podDetails), &podData); err != nil {
			return "", fmt.Errorf("failed to parse pod details: %v", err)
		}

		// Find the main application container (not istio-proxy or istio-init)
		for _, c := range podData.Containers {
			if c.Name != "istio-proxy" && c.Name != "istio-init" {
				podContainer = c.Name
				break
			}
		}
		// If no app container found, use the first container
		if podContainer == "" && len(podData.Containers) > 0 {
			podContainer = podData.Containers[0].Name
		}

		if podContainer == "" {
			return "", fmt.Errorf("no container found for pod %s in namespace %s", podName, namespace)
		}
	}

	endpoint := fmt.Sprintf("%s/api/namespaces/%s/pods/%s/logs",
		strings.TrimRight(baseURL, "/"), url.PathEscape(namespace), url.PathEscape(podName))

	// Add query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()

	// Required parameters
	q.Set("container", podContainer)

	// Optional parameters
	if workload != "" {
		q.Set("workload", workload)
	}
	if service != "" {
		q.Set("service", service)
	}
	if duration != "" {
		q.Set("duration", duration)
	}
	if logType != "" {
		q.Set("logType", logType)
	}
	if sinceTime != "" {
		q.Set("sinceTime", sinceTime)
	}
	if maxLines != "" {
		q.Set("maxLines", maxLines)
	}

	u.RawQuery = q.Encode()
	endpoint = u.String()

	return k.executeRequest(ctx, authHeader, endpoint)
}

func (m *Manager) Derived(ctx context.Context) (*Kiali, error) {
	authorization, ok := ctx.Value(internalk8s.OAuthAuthorizationHeader).(string)
	if !ok || !strings.HasPrefix(authorization, "Bearer ") {
		if m.staticConfig.RequireOAuth {
			return nil, errors.New("oauth token required")
		}
		return &Kiali{manager: m}, nil
	}
	klog.V(5).Infof("%s header found (Bearer), using provided bearer token", internalk8s.OAuthAuthorizationHeader)
	derivedCfg := &rest.Config{
		Host:    m.cfg.Host,
		APIPath: m.cfg.APIPath,
		// Copy only server verification TLS settings (CA bundle and server name)
		TLSClientConfig: rest.TLSClientConfig{
			Insecure:   m.cfg.Insecure,
			ServerName: m.cfg.ServerName,
			CAFile:     m.cfg.CAFile,
			CAData:     m.cfg.CAData,
		},
		BearerToken: strings.TrimPrefix(authorization, "Bearer "),
		// pass custom UserAgent to identify the client
		UserAgent:   internalk8s.CustomUserAgent,
		QPS:         m.cfg.QPS,
		Burst:       m.cfg.Burst,
		Timeout:     m.cfg.Timeout,
		Impersonate: rest.ImpersonationConfig{},
	}
	derived := &Kiali{manager: &Manager{
		cfg:          derivedCfg,
		staticConfig: m.staticConfig,
	}}
	return derived, nil
}
