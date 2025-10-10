package kiali

import (
	"context"
	"crypto/tls"
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
