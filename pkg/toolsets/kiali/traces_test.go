package kiali

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kiali/kiali-mcp-server/pkg/config"
	internalkiali "github.com/kiali/kiali-mcp-server/pkg/kiali"
)

func TestAppTraces_KialiClient(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		app           string
		queryParams   map[string]string
		mockResponse  string
		expectedURL   string
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "basic app traces request",
			namespace:    "bookinfo",
			app:          "productpage",
			queryParams:  map[string]string{},
			mockResponse: `{"traces": [{"traceID": "trace1", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces",
		},
		{
			name:      "app traces with time filter",
			namespace: "bookinfo",
			app:       "productpage",
			queryParams: map[string]string{
				"startMicros": "1640995200000000",
				"endMicros":   "1640995260000000",
			},
			mockResponse: `{"traces": [{"traceID": "trace2", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces?endMicros=1640995260000000&startMicros=1640995200000000",
		},
		{
			name:      "app traces with limit",
			namespace: "bookinfo",
			app:       "productpage",
			queryParams: map[string]string{
				"limit": "50",
			},
			mockResponse: `{"traces": [{"traceID": "trace3", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces?limit=50",
		},
		{
			name:      "app traces with min duration",
			namespace: "bookinfo",
			app:       "productpage",
			queryParams: map[string]string{
				"minDuration": "1000",
			},
			mockResponse: `{"traces": [{"traceID": "trace4", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces?minDuration=1000",
		},
		{
			name:      "app traces with tags",
			namespace: "bookinfo",
			app:       "productpage",
			queryParams: map[string]string{
				"tags": `{"http.method":"GET"}`,
			},
			mockResponse: `{"traces": [{"traceID": "trace5", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces?tags=%7B%22http.method%22%3A%22GET%22%7D",
		},
		{
			name:      "app traces with cluster name",
			namespace: "bookinfo",
			app:       "productpage",
			queryParams: map[string]string{
				"clusterName": "cluster1",
			},
			mockResponse: `{"traces": [{"traceID": "trace6", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/apps/productpage/traces?clusterName=cluster1",
		},
		{
			name:          "missing namespace",
			namespace:     "",
			app:           "productpage",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "namespace is required",
		},
		{
			name:          "missing app name",
			namespace:     "bookinfo",
			app:           "",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "app name is required",
		},
		{
			name:          "server error",
			namespace:     "bookinfo",
			app:           "productpage",
			queryParams:   map[string]string{},
			mockResponse:  "Internal Server Error",
			expectedError: true,
			errorMessage:  "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if the URL matches expected
				if !strings.Contains(r.URL.Path, strings.Split(tt.expectedURL, "?")[0]) {
					t.Errorf("Expected URL path to contain %s, got %s", strings.Split(tt.expectedURL, "?")[0], r.URL.Path)
				}

				// Check query parameters if expected
				if strings.Contains(tt.expectedURL, "?") {
					expectedQuery := strings.Split(tt.expectedURL, "?")[1]
					actualQuery := r.URL.RawQuery
					if actualQuery != expectedQuery {
						t.Errorf("Expected query parameters %s, got %s", expectedQuery, actualQuery)
					}
				}

				if tt.expectedError {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(tt.errorMessage))
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			// Create Kiali client with mock server URL
			cfg := &config.StaticConfig{
				KialiServerURL: server.URL,
			}
			kialiClient := internalkiali.NewFromConfig(cfg)

			// Test the AppTraces method
			result, err := kialiClient.AppTraces(context.Background(), "Bearer test-token", tt.namespace, tt.app, tt.queryParams)

			// Check for expected errors
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("Expected response '%s', got '%s'", tt.mockResponse, result)
				}
			}
		})
	}
}

func TestServiceTraces_KialiClient(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		service       string
		queryParams   map[string]string
		mockResponse  string
		expectedURL   string
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "basic service traces request",
			namespace:    "bookinfo",
			service:      "productpage",
			queryParams:  map[string]string{},
			mockResponse: `{"traces": [{"traceID": "trace1", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces",
		},
		{
			name:      "service traces with time filter",
			namespace: "bookinfo",
			service:   "productpage",
			queryParams: map[string]string{
				"startMicros": "1640995200000000",
				"endMicros":   "1640995260000000",
			},
			mockResponse: `{"traces": [{"traceID": "trace2", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces?endMicros=1640995260000000&startMicros=1640995200000000",
		},
		{
			name:      "service traces with limit",
			namespace: "bookinfo",
			service:   "productpage",
			queryParams: map[string]string{
				"limit": "50",
			},
			mockResponse: `{"traces": [{"traceID": "trace3", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces?limit=50",
		},
		{
			name:      "service traces with min duration",
			namespace: "bookinfo",
			service:   "productpage",
			queryParams: map[string]string{
				"minDuration": "1000",
			},
			mockResponse: `{"traces": [{"traceID": "trace4", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces?minDuration=1000",
		},
		{
			name:      "service traces with tags",
			namespace: "bookinfo",
			service:   "productpage",
			queryParams: map[string]string{
				"tags": `{"http.method":"GET"}`,
			},
			mockResponse: `{"traces": [{"traceID": "trace5", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces?tags=%7B%22http.method%22%3A%22GET%22%7D",
		},
		{
			name:      "service traces with cluster name",
			namespace: "bookinfo",
			service:   "productpage",
			queryParams: map[string]string{
				"clusterName": "cluster1",
			},
			mockResponse: `{"traces": [{"traceID": "trace6", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/services/productpage/traces?clusterName=cluster1",
		},
		{
			name:          "missing namespace",
			namespace:     "",
			service:       "productpage",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "namespace is required",
		},
		{
			name:          "missing service name",
			namespace:     "bookinfo",
			service:       "",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "service name is required",
		},
		{
			name:          "server error",
			namespace:     "bookinfo",
			service:       "productpage",
			queryParams:   map[string]string{},
			mockResponse:  "Internal Server Error",
			expectedError: true,
			errorMessage:  "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if the URL matches expected
				if !strings.Contains(r.URL.Path, strings.Split(tt.expectedURL, "?")[0]) {
					t.Errorf("Expected URL path to contain %s, got %s", strings.Split(tt.expectedURL, "?")[0], r.URL.Path)
				}

				// Check query parameters if expected
				if strings.Contains(tt.expectedURL, "?") {
					expectedQuery := strings.Split(tt.expectedURL, "?")[1]
					actualQuery := r.URL.RawQuery
					if actualQuery != expectedQuery {
						t.Errorf("Expected query parameters %s, got %s", expectedQuery, actualQuery)
					}
				}

				if tt.expectedError {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(tt.errorMessage))
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			// Create Kiali client with mock server URL
			cfg := &config.StaticConfig{
				KialiServerURL: server.URL,
			}
			kialiClient := internalkiali.NewFromConfig(cfg)

			// Test the ServiceTraces method
			result, err := kialiClient.ServiceTraces(context.Background(), "Bearer test-token", tt.namespace, tt.service, tt.queryParams)

			// Check for expected errors
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("Expected response '%s', got '%s'", tt.mockResponse, result)
				}
			}
		})
	}
}

func TestWorkloadTraces_KialiClient(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		workload      string
		queryParams   map[string]string
		mockResponse  string
		expectedURL   string
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "basic workload traces request",
			namespace:    "bookinfo",
			workload:     "productpage-v1",
			queryParams:  map[string]string{},
			mockResponse: `{"traces": [{"traceID": "trace1", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces",
		},
		{
			name:      "workload traces with time filter",
			namespace: "bookinfo",
			workload:  "productpage-v1",
			queryParams: map[string]string{
				"startMicros": "1640995200000000",
				"endMicros":   "1640995260000000",
			},
			mockResponse: `{"traces": [{"traceID": "trace2", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces?endMicros=1640995260000000&startMicros=1640995200000000",
		},
		{
			name:      "workload traces with limit",
			namespace: "bookinfo",
			workload:  "productpage-v1",
			queryParams: map[string]string{
				"limit": "50",
			},
			mockResponse: `{"traces": [{"traceID": "trace3", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces?limit=50",
		},
		{
			name:      "workload traces with min duration",
			namespace: "bookinfo",
			workload:  "productpage-v1",
			queryParams: map[string]string{
				"minDuration": "1000",
			},
			mockResponse: `{"traces": [{"traceID": "trace4", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces?minDuration=1000",
		},
		{
			name:      "workload traces with tags",
			namespace: "bookinfo",
			workload:  "productpage-v1",
			queryParams: map[string]string{
				"tags": `{"http.method":"GET"}`,
			},
			mockResponse: `{"traces": [{"traceID": "trace5", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces?tags=%7B%22http.method%22%3A%22GET%22%7D",
		},
		{
			name:      "workload traces with cluster name",
			namespace: "bookinfo",
			workload:  "productpage-v1",
			queryParams: map[string]string{
				"clusterName": "cluster1",
			},
			mockResponse: `{"traces": [{"traceID": "trace6", "spans": []}]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/productpage-v1/traces?clusterName=cluster1",
		},
		{
			name:          "missing namespace",
			namespace:     "",
			workload:      "productpage-v1",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "namespace is required",
		},
		{
			name:          "missing workload name",
			namespace:     "bookinfo",
			workload:      "",
			queryParams:   map[string]string{},
			expectedError: true,
			errorMessage:  "workload name is required",
		},
		{
			name:          "server error",
			namespace:     "bookinfo",
			workload:      "productpage-v1",
			queryParams:   map[string]string{},
			mockResponse:  "Internal Server Error",
			expectedError: true,
			errorMessage:  "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if the URL matches expected
				if !strings.Contains(r.URL.Path, strings.Split(tt.expectedURL, "?")[0]) {
					t.Errorf("Expected URL path to contain %s, got %s", strings.Split(tt.expectedURL, "?")[0], r.URL.Path)
				}

				// Check query parameters if expected
				if strings.Contains(tt.expectedURL, "?") {
					expectedQuery := strings.Split(tt.expectedURL, "?")[1]
					actualQuery := r.URL.RawQuery
					if actualQuery != expectedQuery {
						t.Errorf("Expected query parameters %s, got %s", expectedQuery, actualQuery)
					}
				}

				if tt.expectedError {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(tt.errorMessage))
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			// Create Kiali client with mock server URL
			cfg := &config.StaticConfig{
				KialiServerURL: server.URL,
			}
			kialiClient := internalkiali.NewFromConfig(cfg)

			// Test the WorkloadTraces method
			result, err := kialiClient.WorkloadTraces(context.Background(), "Bearer test-token", tt.namespace, tt.workload, tt.queryParams)

			// Check for expected errors
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("Expected response '%s', got '%s'", tt.mockResponse, result)
				}
			}
		})
	}
}
