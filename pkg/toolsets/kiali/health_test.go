package kiali

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kiali/kiali-mcp-server/pkg/api"
	"github.com/kiali/kiali-mcp-server/pkg/config"
	internalkiali "github.com/kiali/kiali-mcp-server/pkg/kiali"
)

// TestHealth_KialiClient tests the Kiali client Health method
func TestHealth_KialiClient(t *testing.T) {
	t.Run("successful health retrieval for all namespaces with default type", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{
						"productpage": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
				},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"",
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "appHealth")

		// Verify URL path
		expectedPath := "/api/clusters/health"
		assert.Equal(t, expectedPath, capturedURL.Path)
	})

	t.Run("successful health retrieval with specific namespaces", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{},
					"default":  map[string]interface{}{},
				},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo,default",
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)

		// Verify namespaces parameter
		assert.Equal(t, "bookinfo,default", capturedURL.Query().Get("namespaces"))
	})

	t.Run("health retrieval with type app", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{},
				},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type": "app",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "app", capturedURL.Query().Get("type"))
	})

	t.Run("health retrieval with type service", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth":      map[string]interface{}{},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type": "service",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "service", capturedURL.Query().Get("type"))
	})

	t.Run("health retrieval with type workload", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{},
				"workloadHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{},
				},
				"serviceHealth": map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type": "workload",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "workload", capturedURL.Query().Get("type"))
	})

	t.Run("health retrieval with custom rateInterval", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth":      map[string]interface{}{},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"rateInterval": "5m",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "5m", capturedURL.Query().Get("rateInterval"))
	})

	t.Run("health retrieval with queryTime", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth":      map[string]interface{}{},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"queryTime": "1609459200",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "1609459200", capturedURL.Query().Get("queryTime"))
	})

	t.Run("health retrieval with all parameters", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{},
				},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type":         "app",
			"rateInterval": "15m",
			"queryTime":    "1609459200",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo,default",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)

		// Verify all parameters
		assert.Equal(t, "bookinfo,default", capturedURL.Query().Get("namespaces"))
		assert.Equal(t, "app", capturedURL.Query().Get("type"))
		assert.Equal(t, "15m", capturedURL.Query().Get("rateInterval"))
		assert.Equal(t, "1609459200", capturedURL.Query().Get("queryTime"))
	})

	t.Run("Kiali server not configured", func(t *testing.T) {
		staticConfig := &config.StaticConfig{
			KialiServerURL: "",
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		_, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			nil,
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "kiali server URL not configured")
	})

	t.Run("Kiali server returns 404", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Namespace not found"))
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		_, err := kialiClient.Health(
			context.Background(),
			"",
			"non-existent-namespace",
			nil,
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Namespace not found")
	})

	t.Run("Kiali server returns 500", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		_, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			nil,
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
	})

	t.Run("empty namespaces parameter retrieves all namespaces", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"namespace1": map[string]interface{}{},
					"namespace2": map[string]interface{}{},
					"namespace3": map[string]interface{}{},
				},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"",
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		// Empty namespaces should not add the parameter to the query
		assert.Empty(t, capturedURL.Query().Get("namespaces"))
	})

	t.Run("special characters in namespace names", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth":      map[string]interface{}{},
				"workloadHealth": map[string]interface{}{},
				"serviceHealth":  map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"my-namespace-123,test-ns-456",
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "my-namespace-123,test-ns-456", capturedURL.Query().Get("namespaces"))
	})
}

// TestHealthToolDefinition tests the tool definition
func TestHealthToolDefinition(t *testing.T) {
	tools := initHealth()

	var healthTool *api.Tool
	for i := range tools {
		if tools[i].Tool.Name == "health" {
			healthTool = &tools[i].Tool
			break
		}
	}

	require.NotNil(t, healthTool, "health tool should be registered")

	t.Run("tool has correct name", func(t *testing.T) {
		assert.Equal(t, "health", healthTool.Name)
	})

	t.Run("tool has description", func(t *testing.T) {
		assert.NotEmpty(t, healthTool.Description)
		assert.Contains(t, strings.ToLower(healthTool.Description), "health")
	})

	t.Run("tool has input schema", func(t *testing.T) {
		assert.NotNil(t, healthTool.InputSchema)
		assert.Equal(t, "object", healthTool.InputSchema.Type)
	})

	t.Run("tool has optional parameters", func(t *testing.T) {
		schema := healthTool.InputSchema
		require.NotNil(t, schema)
		assert.NotNil(t, schema.Properties)

		expectedParams := []string{
			"namespaces", "type", "rateInterval", "queryTime",
		}

		for _, param := range expectedParams {
			_, exists := schema.Properties[param]
			assert.True(t, exists, "Parameter %s should exist in schema", param)
		}
	})

	t.Run("tool parameters have descriptions", func(t *testing.T) {
		schema := healthTool.InputSchema
		require.NotNil(t, schema)

		for name, prop := range schema.Properties {
			assert.NotEmpty(t, prop.Description,
				"Parameter %s should have a description", name)
		}
	})

	t.Run("tool parameters have correct types", func(t *testing.T) {
		schema := healthTool.InputSchema
		require.NotNil(t, schema)

		for name, prop := range schema.Properties {
			assert.Equal(t, "string", prop.Type,
				"Parameter %s should be of type string", name)
		}
	})

	t.Run("tool has annotations", func(t *testing.T) {
		annotations := healthTool.Annotations
		assert.NotNil(t, annotations.ReadOnlyHint)
		assert.True(t, *annotations.ReadOnlyHint)
		assert.NotNil(t, annotations.DestructiveHint)
		assert.False(t, *annotations.DestructiveHint)
		assert.NotNil(t, annotations.IdempotentHint)
		assert.True(t, *annotations.IdempotentHint)
	})

	t.Run("type parameter description mentions valid values", func(t *testing.T) {
		schema := healthTool.InputSchema
		require.NotNil(t, schema)
		typeProp, exists := schema.Properties["type"]
		assert.True(t, exists)
		assert.Contains(t, strings.ToLower(typeProp.Description), "app")
		assert.Contains(t, strings.ToLower(typeProp.Description), "service")
		assert.Contains(t, strings.ToLower(typeProp.Description), "workload")
	})
}

// TestHealthRealWorldScenarios tests real-world user scenarios
func TestHealthRealWorldScenarios(t *testing.T) {
	t.Run("retrieve all app health across all namespaces", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"appHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{
						"details": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
						"productpage": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
					"default": map[string]interface{}{
						"kubernetes": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"",
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "appHealth")

		// Verify no namespace filter when querying all
		assert.Empty(t, capturedURL.Query().Get("namespaces"))
	})

	t.Run("retrieve service health for specific namespace", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"serviceHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{
						"details": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
						"productpage": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type": "service",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "serviceHealth")
		assert.Equal(t, "bookinfo", capturedURL.Query().Get("namespaces"))
		assert.Equal(t, "service", capturedURL.Query().Get("type"))
	})

	t.Run("retrieve workload health for multiple namespaces", func(t *testing.T) {
		var capturedURL *url.URL
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"workloadHealth": map[string]interface{}{
					"bookinfo": map[string]interface{}{
						"details-v1": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
					"istio-system": map[string]interface{}{
						"istiod": map[string]interface{}{
							"requests": map[string]interface{}{
								"errorRatio": 0.0,
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		staticConfig := &config.StaticConfig{
			KialiServerURL: mockServer.URL,
		}

		kialiClient := internalkiali.NewFromConfig(staticConfig)

		queryParams := map[string]string{
			"type": "workload",
		}

		result, err := kialiClient.Health(
			context.Background(),
			"",
			"bookinfo,istio-system",
			queryParams,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "workloadHealth")
		assert.Equal(t, "bookinfo,istio-system", capturedURL.Query().Get("namespaces"))
		assert.Equal(t, "workload", capturedURL.Query().Get("type"))
	})
}
