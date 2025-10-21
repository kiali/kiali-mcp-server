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
