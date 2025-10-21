package kiali

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kiali/kiali-mcp-server/pkg/config"
	internalkiali "github.com/kiali/kiali-mcp-server/pkg/kiali"
)

func TestWorkloadLogs_KialiClient(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		workload      string
		container     string
		queryParams   map[string]string
		mockResponse  string
		expectedURL   string
		expectedError bool
		errorMessage  string
	}{
		{
			name:         "basic workload logs request",
			namespace:    "bookinfo",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Application started", "2024-01-01T10:01:00Z INFO: Processing request"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1",
		},
		{
			name:         "workload logs with container filter",
			namespace:    "bookinfo",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Reviews service started"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1/logs?container=reviews",
		},
		{
			name:      "workload logs with time filter",
			namespace: "bookinfo",
			workload:  "reviews-v1",
			container: "reviews",
			queryParams: map[string]string{
				"since": "5m",
				"tail":  "50",
			},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Recent logs"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1",
		},
		{
			name:      "workload logs with previous container logs",
			namespace: "bookinfo",
			workload:  "reviews-v1",
			container: "reviews",
			queryParams: map[string]string{
				"previous": "true",
				"tail":     "100",
			},
			mockResponse: `{"logs": ["2024-01-01T09:00:00Z INFO: Previous container logs"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1",
		},
		{
			name:         "workload logs with all parameters",
			namespace:    "bookinfo",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{"since": "10m", "tail": "200", "previous": "true"},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Comprehensive logs"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1/logs?container=reviews&previous=true&since=10m&tail=200",
		},
		{
			name:          "missing namespace",
			namespace:     "",
			workload:      "reviews-v1",
			container:     "",
			queryParams:   map[string]string{},
			mockResponse:  "",
			expectedURL:   "",
			expectedError: true,
			errorMessage:  "namespace is required",
		},
		{
			name:          "missing workload",
			namespace:     "bookinfo",
			workload:      "",
			container:     "",
			queryParams:   map[string]string{},
			mockResponse:  "",
			expectedURL:   "",
			expectedError: true,
			errorMessage:  "workload name is required",
		},
		{
			name:         "empty logs response",
			namespace:    "bookinfo",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{},
			mockResponse: `{"logs": []}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1",
		},
		{
			name:         "large logs response",
			namespace:    "bookinfo",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{"tail": "1000"},
			mockResponse: `{"logs": ["` + strings.Repeat("2024-01-01T10:00:00Z INFO: Log entry\n", 1000) + `"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1",
		},
		{
			name:         "special characters in workload name",
			namespace:    "bookinfo",
			workload:     "reviews-v1-with-special_chars",
			container:    "reviews",
			queryParams:  map[string]string{},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Special workload logs"]}`,
			expectedURL:  "/api/namespaces/bookinfo/workloads/reviews-v1-with-special_chars",
		},
		{
			name:         "special characters in namespace",
			namespace:    "book-info-namespace",
			workload:     "reviews-v1",
			container:    "reviews",
			queryParams:  map[string]string{},
			mockResponse: `{"logs": ["2024-01-01T10:00:00Z INFO: Namespace with dashes"]}`,
			expectedURL:  "/api/namespaces/book-info-namespace/workloads/reviews-v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle WorkloadDetails request
				if strings.Contains(r.URL.Path, "/workloads/") && !strings.Contains(r.URL.Path, "/logs") {
					// Mock WorkloadDetails response
					workloadDetailsResponse := `{
						"pods": [
							{
								"name": "reviews-v1-pod-1",
								"containers": [
									{"name": "reviews"},
									{"name": "istio-proxy"}
								]
							}
						]
					}`
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(workloadDetailsResponse))
					return
				}

				// Handle PodLogs request
				if strings.Contains(r.URL.Path, "/logs") {
					// Verify the request URL for logs
					if tt.expectedURL != "" {
						// For logs, we expect the path to contain the pod name and logs endpoint
						if !strings.Contains(r.URL.Path, "/logs") {
							t.Errorf("Expected logs endpoint, got %s", r.URL.Path)
						}
					}

					// Verify the request method
					if r.Method != http.MethodGet {
						t.Errorf("Expected GET request, got %s", r.Method)
					}

					// Verify the Authorization header
					if r.Header.Get("Authorization") == "" {
						t.Error("Expected Authorization header to be set")
					}

					// Return the mock response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockResponse))
					return
				}

				// Default response for other requests
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create Kiali client with mock server URL
			cfg := &config.StaticConfig{
				KialiServerURL: server.URL,
			}
			kialiClient := internalkiali.NewFromConfig(cfg)

			// Test the WorkloadLogs method
			result, err := kialiClient.WorkloadLogs(context.Background(), "Bearer test-token", tt.namespace, tt.workload, tt.container, "", "", "", "", "")

			// Check for expected errors
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMessage, err.Error())
				}
				return
			}

			// Check for unexpected errors
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the response
			if tt.mockResponse != "" {
				expectedResponse := fmt.Sprintf("=== Pod: reviews-v1-pod-1 (Container: reviews) ===\n%s", tt.mockResponse)
				if result != expectedResponse {
					t.Errorf("Expected response %s, got %s", expectedResponse, result)
				}
			}
		})
	}
}

func TestWorkloadLogsToolDefinition(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "tool_has_correct_name",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				if tools[0].Tool.Name != "workload_logs" {
					t.Errorf("Expected tool name 'workload_logs', got '%s'", tools[0].Tool.Name)
				}
			},
		},
		{
			name: "tool_has_description",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				if tools[0].Tool.Description == "" {
					t.Error("Expected tool to have a description")
				}
				if !strings.Contains(tools[0].Tool.Description, "workload") {
					t.Error("Expected description to mention 'workload'")
				}
				if !strings.Contains(tools[0].Tool.Description, "logs") {
					t.Error("Expected description to mention 'logs'")
				}
			},
		},
		{
			name: "tool_has_input_schema",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				if tools[0].Tool.InputSchema == nil {
					t.Error("Expected tool to have an input schema")
				}
				if tools[0].Tool.InputSchema.Type != "object" {
					t.Errorf("Expected input schema type 'object', got '%s'", tools[0].Tool.InputSchema.Type)
				}
			},
		},
		{
			name: "tool_has_required_parameters",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				required := tools[0].Tool.InputSchema.Required
				if len(required) != 2 {
					t.Errorf("Expected 2 required parameters, got %d", len(required))
				}
				expectedRequired := []string{"namespace", "workload"}
				for _, req := range expectedRequired {
					found := false
					for _, r := range required {
						if r == req {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected required parameter '%s' not found", req)
					}
				}
			},
		},
		{
			name: "tool_has_optional_parameters",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				properties := tools[0].Tool.InputSchema.Properties
				expectedOptional := []string{"container", "since", "tail", "previous"}
				for _, opt := range expectedOptional {
					if _, exists := properties[opt]; !exists {
						t.Errorf("Expected optional parameter '%s' not found", opt)
					}
				}
			},
		},
		{
			name: "tool_parameters_have_descriptions",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				properties := tools[0].Tool.InputSchema.Properties
				for paramName, paramSchema := range properties {
					if paramSchema.Description == "" {
						t.Errorf("Parameter '%s' should have a description", paramName)
					}
				}
			},
		},
		{
			name: "tool_parameters_have_correct_types",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				properties := tools[0].Tool.InputSchema.Properties

				// Check string parameters
				stringParams := []string{"namespace", "workload", "container", "since"}
				for _, param := range stringParams {
					if schema, exists := properties[param]; exists {
						if schema.Type != "string" {
							t.Errorf("Expected parameter '%s' to be string type, got '%s'", param, schema.Type)
						}
					}
				}

				// Check integer parameters
				if schema, exists := properties["tail"]; exists {
					if schema.Type != "integer" {
						t.Errorf("Expected parameter 'tail' to be integer type, got '%s'", schema.Type)
					}
				}

				// Check boolean parameters
				if schema, exists := properties["previous"]; exists {
					if schema.Type != "boolean" {
						t.Errorf("Expected parameter 'previous' to be boolean type, got '%s'", schema.Type)
					}
				}
			},
		},
		{
			name: "tool_has_annotations",
			testFunc: func(t *testing.T) {
				tools := initLogs()
				if len(tools) == 0 {
					t.Fatal("Expected at least one tool")
				}
				annotations := tools[0].Tool.Annotations
				if annotations.Title == "" {
					t.Error("Expected tool to have a title annotation")
				}
				if annotations.ReadOnlyHint == nil || !*annotations.ReadOnlyHint {
					t.Error("Expected tool to be marked as read-only")
				}
				if annotations.DestructiveHint == nil || *annotations.DestructiveHint {
					t.Error("Expected tool to be marked as non-destructive")
				}
				if annotations.IdempotentHint == nil || *annotations.IdempotentHint {
					t.Error("Expected tool to be marked as non-idempotent (logs can change)")
				}
				if annotations.OpenWorldHint == nil || !*annotations.OpenWorldHint {
					t.Error("Expected tool to be marked as open-world")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestWorkloadLogsRealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		namespace   string
		workload    string
		container   string
		queryParams map[string]string
		description string
	}{
		{
			name:        "debug_recent_errors",
			namespace:   "bookinfo",
			workload:    "reviews-v1",
			container:   "reviews",
			queryParams: map[string]string{"since": "10m", "tail": "100"},
			description: "Get recent logs from reviews workload to debug errors",
		},
		{
			name:        "monitor_specific_container",
			namespace:   "bookinfo",
			workload:    "productpage-v1",
			container:   "productpage",
			queryParams: map[string]string{"tail": "50"},
			description: "Monitor logs from specific container in productpage workload",
		},
		{
			name:        "investigate_crash",
			namespace:   "bookinfo",
			workload:    "ratings-v1",
			container:   "ratings",
			queryParams: map[string]string{"previous": "true", "tail": "200"},
			description: "Investigate crash by getting logs from previous terminated containers",
		},
		{
			name:        "comprehensive_logging",
			namespace:   "bookinfo",
			workload:    "details-v1",
			container:   "details",
			queryParams: map[string]string{"since": "1h", "tail": "500", "previous": "true"},
			description: "Get comprehensive logs for detailed analysis",
		},
		{
			name:        "production_monitoring",
			namespace:   "production",
			workload:    "api-gateway",
			container:   "api-gateway",
			queryParams: map[string]string{"since": "5m", "tail": "1000"},
			description: "Monitor production API gateway logs in real-time",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create a mock HTTP server for this scenario
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle WorkloadDetails request
				if strings.Contains(r.URL.Path, "/workloads/") && !strings.Contains(r.URL.Path, "/logs") {
					// Mock WorkloadDetails response
					workloadDetailsResponse := `{
						"pods": [
							{
								"name": "` + scenario.workload + `-pod-1",
								"containers": [
									{"name": "` + scenario.container + `"},
									{"name": "istio-proxy"}
								]
							}
						]
					}`
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(workloadDetailsResponse))
					return
				}

				// Handle PodLogs request
				if strings.Contains(r.URL.Path, "/logs") {
					// Verify the request is properly formed
					expectedPath := fmt.Sprintf("/api/namespaces/%s/pods/%s-pod-1/logs", scenario.namespace, scenario.workload)
					if !strings.HasPrefix(r.URL.Path, expectedPath) {
						t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
					}

					// Verify container parameter if specified
					if scenario.container != "" {
						if r.URL.Query().Get("container") != scenario.container {
							t.Errorf("Expected container parameter %s, got %s", scenario.container, r.URL.Query().Get("container"))
						}
					}

					// Verify other query parameters
					for key, expectedValue := range scenario.queryParams {
						var actualValue string
						switch key {
						case "since":
							actualValue = r.URL.Query().Get("duration")
						case "tail":
							actualValue = r.URL.Query().Get("maxLines")
						case "previous":
							actualValue = r.URL.Query().Get("sinceTime")
						default:
							actualValue = r.URL.Query().Get(key)
						}
						if actualValue != expectedValue {
							t.Errorf("Expected query parameter %s=%s, got %s", key, expectedValue, actualValue)
						}
					}

					// Return mock response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					mockLogs := fmt.Sprintf(`{"logs": ["2024-01-01T10:00:00Z INFO: %s - %s"]}`, scenario.name, scenario.description)
					w.Write([]byte(mockLogs))
					return
				}

				// Default response for other requests
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create Kiali client and test the scenario
			cfg := &config.StaticConfig{
				KialiServerURL: server.URL,
			}
			kialiClient := internalkiali.NewFromConfig(cfg)

			result, err := kialiClient.WorkloadLogs(
				context.Background(),
				"Bearer test-token",
				scenario.namespace,
				scenario.workload,
				scenario.container,
				"",                               // service
				scenario.queryParams["since"],    // duration
				"",                               // logType
				scenario.queryParams["previous"], // sinceTime (for previous logs)
				scenario.queryParams["tail"],     // maxLines
			)

			if err != nil {
				t.Errorf("Scenario '%s' failed with error: %v", scenario.name, err)
				return
			}

			// Verify the response format includes pod headers with container info
			expectedResponse := fmt.Sprintf("=== Pod: %s-pod-1 (Container: %s) ===\n{\"logs\": [\"2024-01-01T10:00:00Z INFO: %s - %s\"]}", scenario.workload, scenario.container, scenario.name, scenario.description)
			if result != expectedResponse {
				t.Errorf("Scenario '%s' expected response %s, got %s", scenario.name, expectedResponse, result)
				return
			}

			// Verify the response contains expected content
			if !strings.Contains(result, scenario.name) {
				t.Errorf("Scenario '%s' response should contain scenario name", scenario.name)
			}
		})
	}
}
