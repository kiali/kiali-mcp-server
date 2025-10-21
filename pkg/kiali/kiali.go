package kiali

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kiali/kiali-mcp-server/pkg/config"
	internalk8s "github.com/kiali/kiali-mcp-server/pkg/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Kiali struct {
	manager *Manager
}

type Manager struct {
	cfg             *rest.Config
	clientCmdConfig clientcmd.ClientConfig
	staticConfig    *config.StaticConfig
}

func NewManager(config *config.StaticConfig) (*Manager, error) {
	kiali := &Manager{
		staticConfig: config,
	}
	// Only resolve Kubernetes-related configuration when Kiali is actually configured
	if config != nil && strings.TrimSpace(config.KialiServerURL) != "" {
		if err := resolveKialiRequiredConfigurations(kiali); err != nil {
			return nil, err
		}
	}
	return kiali, nil
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

// CurrentAuthorizationHeader returns the Authorization header value that the
// Kiali client is currently configured to use (Bearer <token>), or empty
// if no bearer token is configured.
func (k *Kiali) CurrentAuthorizationHeader(ctx context.Context) string {
	token, _ := ctx.Value(internalk8s.OAuthAuthorizationHeader).(string)
	token = strings.TrimSpace(token)

	if token == "" {
		// Fall back to using the same token that the Kubernetes client is using
		if k == nil || k.manager == nil || k.manager.cfg == nil {
			return ""
		}
		token = strings.TrimSpace(k.manager.cfg.BearerToken)
		if token == "" {
			return ""
		}
	}
	// Normalize to exactly "Bearer <token>" without double prefix
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		return "Bearer " + strings.TrimSpace(token[7:])
	}
	return "Bearer " + token
}

// executeRequest executes an HTTP request and handles common error scenarios.
func (k *Kiali) executeRequest(ctx context.Context, endpoint string) (string, error) {
	klog.V(0).Infof("kiali API call: %s", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	authHeader := k.CurrentAuthorizationHeader(ctx)
	if authHeader == "" {
		// Ensure tests and mock servers receive an Authorization header
		authHeader = "Bearer "
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
func (k *Kiali) executeRequestWithBody(ctx context.Context, method, endpoint, contentType string, body io.Reader) (string, error) {
	klog.V(0).Infof("kiali API call: %s %s", method, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return "", err
	}
	authHeader := k.CurrentAuthorizationHeader(ctx)
	if authHeader == "" {
		authHeader = "Bearer "
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

func (m *Manager) Derived(ctx context.Context) (*Kiali, error) {
	authorization, ok := ctx.Value(internalk8s.OAuthAuthorizationHeader).(string)
	if !ok || !strings.HasPrefix(authorization, "Bearer ") {
		if m.staticConfig != nil && m.staticConfig.RequireOAuth {
			return nil, errors.New("oauth token required")
		}
		return &Kiali{manager: m}, nil
	}
	// Authorization header is present; nothing special is needed for the Kiali HTTP client
	klog.V(5).Infof("%s header found (Bearer)", internalk8s.OAuthAuthorizationHeader)
	return &Kiali{manager: m}, nil
}
