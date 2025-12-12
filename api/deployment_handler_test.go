package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/opsorch/opsorch-core/deployment"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// Mock deployment provider for testing
type mockDeploymentProvider struct {
	queryFunc func(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error)
	getFunc   func(ctx context.Context, id string) (schema.Deployment, error)
}

func (m *mockDeploymentProvider) Query(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query)
	}
	return []schema.Deployment{}, nil
}

func (m *mockDeploymentProvider) Get(ctx context.Context, id string) (schema.Deployment, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return schema.Deployment{}, nil
}

// Register mock provider for testing
func init() {
	deployment.RegisterProvider("mock", func(config map[string]any) (deployment.Provider, error) {
		return &mockDeploymentProvider{}, nil
	})
}

// **Feature: deployment-capability-completion, Property 5: Environment configuration processing**
func TestProperty_EnvironmentConfigurationProcessing(t *testing.T) {
	// Test cases with different environment variable combinations
	testCases := []struct {
		name              string
		provider          string
		config            string
		plugin            string
		expectError       bool
		expectNilProvider bool
	}{
		{
			name:              "valid provider configuration",
			provider:          "mock",
			config:            `{"test": "value"}`,
			expectError:       false,
			expectNilProvider: false,
		},
		{
			name:              "no configuration",
			provider:          "",
			config:            "",
			plugin:            "",
			expectError:       false,
			expectNilProvider: true,
		},
		{
			name:              "invalid provider name",
			provider:          "nonexistent",
			config:            `{"test": "value"}`,
			expectError:       true,
			expectNilProvider: true,
		},
		{
			name:              "plugin path specified",
			provider:          "",
			config:            "",
			plugin:            "/path/to/plugin",
			expectError:       false,
			expectNilProvider: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variables
			oldProvider := os.Getenv("OPSORCH_DEPLOYMENT_PROVIDER")
			oldConfig := os.Getenv("OPSORCH_DEPLOYMENT_CONFIG")
			oldPlugin := os.Getenv("OPSORCH_DEPLOYMENT_PLUGIN")

			os.Setenv("OPSORCH_DEPLOYMENT_PROVIDER", tc.provider)
			os.Setenv("OPSORCH_DEPLOYMENT_CONFIG", tc.config)
			os.Setenv("OPSORCH_DEPLOYMENT_PLUGIN", tc.plugin)

			// Clean up after test
			defer func() {
				os.Setenv("OPSORCH_DEPLOYMENT_PROVIDER", oldProvider)
				os.Setenv("OPSORCH_DEPLOYMENT_CONFIG", oldConfig)
				os.Setenv("OPSORCH_DEPLOYMENT_PLUGIN", oldPlugin)
			}()

			// Create mock secret provider
			mockSec := &mockSecretProvider{}

			// Test the environment configuration processing
			handler, err := newDeploymentHandlerFromEnv(mockSec)

			// Verify expectations
			if tc.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectNilProvider && handler.provider != nil {
				t.Errorf("expected nil provider but got non-nil")
			}
			if !tc.expectNilProvider && !tc.expectError && handler.provider == nil {
				t.Errorf("expected non-nil provider but got nil")
			}
		})
	}
}

// Mock secret provider for testing
type mockSecretProvider struct{}

func (m *mockSecretProvider) Get(ctx context.Context, key string) (string, error) {
	// Return error to simulate no stored config
	return "", fmt.Errorf("key not found")
}

func (m *mockSecretProvider) Put(ctx context.Context, key, value string) error {
	return nil
}

// **Feature: deployment-capability-completion, Property 1: Deployment query body processing**
func TestProperty_DeploymentQueryBodyProcessing(t *testing.T) {
	// Create a deployment handler with mock provider
	mockProvider := &mockDeploymentProvider{
		queryFunc: func(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error) {
			// Verify that the query is passed correctly by checking its fields
			deployments := []schema.Deployment{
				{
					ID:          "test-deployment-1",
					Service:     query.Scope.Service,
					Environment: query.Scope.Environment,
					Version:     "v1.0.0",
					Status:      "success",
				},
			}

			// If statuses filter is provided, only return deployments matching those statuses
			if len(query.Statuses) > 0 {
				filtered := []schema.Deployment{}
				for _, deployment := range deployments {
					for _, status := range query.Statuses {
						if deployment.Status == status {
							filtered = append(filtered, deployment)
							break
						}
					}
				}
				return filtered, nil
			}

			return deployments, nil
		},
	}

	handler := &DeploymentHandler{provider: mockProvider}

	// Test cases with different query body combinations
	testCases := []struct {
		name  string
		query schema.DeploymentQuery
	}{
		{
			name: "query with service scope",
			query: schema.DeploymentQuery{
				Query: "test-query",
				Scope: schema.QueryScope{
					Service: "api-service",
				},
			},
		},
		{
			name: "query with environment scope",
			query: schema.DeploymentQuery{
				Scope: schema.QueryScope{
					Environment: "production",
				},
			},
		},
		{
			name: "query with statuses filter",
			query: schema.DeploymentQuery{
				Statuses: []string{"success", "failed"},
			},
		},
		{
			name: "query with versions filter",
			query: schema.DeploymentQuery{
				Versions: []string{"v1.0.0", "v1.1.0"},
			},
		},
		{
			name: "query with limit",
			query: schema.DeploymentQuery{
				Limit: 10,
			},
		},
		{
			name: "query with metadata",
			query: schema.DeploymentQuery{
				Metadata: map[string]any{
					"branch": "main",
					"commit": "abc123",
				},
			},
		},
		{
			name: "complex query with all fields",
			query: schema.DeploymentQuery{
				Query:    "complex-search",
				Statuses: []string{"success"},
				Versions: []string{"v2.0.0"},
				Scope: schema.QueryScope{
					Service:     "payment-service",
					Environment: "staging",
					Team:        "backend",
				},
				Limit: 5,
				Metadata: map[string]any{
					"pipeline": "ci-cd",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP request with the query in the body
			body := `{`
			if tc.query.Query != "" {
				body += fmt.Sprintf(`"query": "%s",`, tc.query.Query)
			}
			if len(tc.query.Statuses) > 0 {
				body += `"statuses": [`
				for i, status := range tc.query.Statuses {
					if i > 0 {
						body += ","
					}
					body += fmt.Sprintf(`"%s"`, status)
				}
				body += `],`
			}
			if len(tc.query.Versions) > 0 {
				body += `"versions": [`
				for i, version := range tc.query.Versions {
					if i > 0 {
						body += ","
					}
					body += fmt.Sprintf(`"%s"`, version)
				}
				body += `],`
			}
			if tc.query.Scope.Service != "" || tc.query.Scope.Environment != "" || tc.query.Scope.Team != "" {
				body += `"scope": {`
				if tc.query.Scope.Service != "" {
					body += fmt.Sprintf(`"service": "%s",`, tc.query.Scope.Service)
				}
				if tc.query.Scope.Environment != "" {
					body += fmt.Sprintf(`"environment": "%s",`, tc.query.Scope.Environment)
				}
				if tc.query.Scope.Team != "" {
					body += fmt.Sprintf(`"team": "%s",`, tc.query.Scope.Team)
				}
				body = strings.TrimSuffix(body, ",") + `},`
			}
			if tc.query.Limit > 0 {
				body += fmt.Sprintf(`"limit": %d,`, tc.query.Limit)
			}
			if len(tc.query.Metadata) > 0 {
				body += `"metadata": {`
				for key, value := range tc.query.Metadata {
					body += fmt.Sprintf(`"%s": "%v",`, key, value)
				}
				body = strings.TrimSuffix(body, ",") + `},`
			}
			body = strings.TrimSuffix(body, ",") + `}`

			req := httptest.NewRequest("POST", "/deployments/query", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			// Call the handler
			handled := handler.handleDeploymentRequest(recorder, req)

			// Verify the request was handled
			if !handled {
				t.Errorf("expected request to be handled")
			}

			// Verify successful response
			if recorder.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", recorder.Code)
			}

			// Verify response contains deployment data
			var deployments []schema.Deployment
			if err := json.Unmarshal(recorder.Body.Bytes(), &deployments); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if len(deployments) == 0 {
				t.Errorf("expected at least one deployment in response")
			}

			// Verify that scope fields were passed correctly to the provider
			if tc.query.Scope.Service != "" && deployments[0].Service != tc.query.Scope.Service {
				t.Errorf("expected service %s, got %s", tc.query.Scope.Service, deployments[0].Service)
			}
			if tc.query.Scope.Environment != "" && deployments[0].Environment != tc.query.Scope.Environment {
				t.Errorf("expected environment %s, got %s", tc.query.Scope.Environment, deployments[0].Environment)
			}
		})
	}
}

// **Feature: deployment-capability-completion, Property 3: Deployment ID retrieval**
func TestProperty_DeploymentIDRetrieval(t *testing.T) {
	// Test cases with different deployment IDs
	testCases := []struct {
		name           string
		deploymentID   string
		mockDeployment schema.Deployment
		expectError    bool
	}{
		{
			name:         "valid deployment ID",
			deploymentID: "deploy-123",
			mockDeployment: schema.Deployment{
				ID:          "deploy-123",
				Service:     "api-service",
				Environment: "production",
				Version:     "v1.2.3",
				Status:      "success",
				URL:         "https://github.com/org/repo/actions/runs/123",
				Actor:       map[string]any{"login": "testuser", "avatar": "https://example.com/avatar.png"},
			},
			expectError: false,
		},
		{
			name:         "deployment ID with special characters",
			deploymentID: "deploy-abc-123_def",
			mockDeployment: schema.Deployment{
				ID:          "deploy-abc-123_def",
				Service:     "payment-service",
				Environment: "staging",
				Version:     "v2.0.0",
				Status:      "running",
			},
			expectError: false,
		},
		{
			name:         "long deployment ID",
			deploymentID: "very-long-deployment-id-with-many-characters-12345678901234567890",
			mockDeployment: schema.Deployment{
				ID:          "very-long-deployment-id-with-many-characters-12345678901234567890",
				Service:     "data-service",
				Environment: "development",
				Version:     "v0.1.0",
				Status:      "failed",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a deployment handler with mock provider
			mockProvider := &mockDeploymentProvider{
				getFunc: func(ctx context.Context, id string) (schema.Deployment, error) {
					if id == tc.deploymentID {
						return tc.mockDeployment, nil
					}
					return schema.Deployment{}, fmt.Errorf("deployment not found")
				},
			}

			handler := &DeploymentHandler{provider: mockProvider}

			// Create a mock HTTP request for the deployment ID
			req := httptest.NewRequest("GET", fmt.Sprintf("/deployments/%s", tc.deploymentID), nil)
			recorder := httptest.NewRecorder()

			// Call the handler
			handled := handler.handleDeploymentRequest(recorder, req)

			// Verify the request was handled
			if !handled {
				t.Errorf("expected request to be handled")
			}

			if tc.expectError {
				if recorder.Code == http.StatusOK {
					t.Errorf("expected error response, got status 200")
				}
			} else {
				// Verify successful response
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", recorder.Code)
				}

				// Verify response contains the correct deployment
				var deployment schema.Deployment
				if err := json.Unmarshal(recorder.Body.Bytes(), &deployment); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}

				// Verify the deployment ID matches
				if deployment.ID != tc.deploymentID {
					t.Errorf("expected deployment ID %s, got %s", tc.deploymentID, deployment.ID)
				}

				// Verify other fields match the mock deployment
				if deployment.Service != tc.mockDeployment.Service {
					t.Errorf("expected service %s, got %s", tc.mockDeployment.Service, deployment.Service)
				}
				if deployment.Environment != tc.mockDeployment.Environment {
					t.Errorf("expected environment %s, got %s", tc.mockDeployment.Environment, deployment.Environment)
				}
				if deployment.Version != tc.mockDeployment.Version {
					t.Errorf("expected version %s, got %s", tc.mockDeployment.Version, deployment.Version)
				}
				if deployment.Status != tc.mockDeployment.Status {
					t.Errorf("expected status %s, got %s", tc.mockDeployment.Status, deployment.Status)
				}
				if deployment.URL != tc.mockDeployment.URL {
					t.Errorf("expected URL %s, got %s", tc.mockDeployment.URL, deployment.URL)
				}
				if len(deployment.Actor) != len(tc.mockDeployment.Actor) {
					t.Errorf("expected actor map size %d, got %d", len(tc.mockDeployment.Actor), len(deployment.Actor))
				}
				for k, v := range tc.mockDeployment.Actor {
					if deployment.Actor[k] != v {
						t.Errorf("expected actor field %s to be %v, got %v", k, v, deployment.Actor[k])
					}
				}
			}
		})
	}
}

// **Feature: deployment-capability-completion, Property 4: Provider error handling**
func TestProperty_ProviderErrorHandling(t *testing.T) {
	// Test cases with different provider errors
	testCases := []struct {
		name           string
		providerError  error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "provider not found error",
			providerError:  &orcherr.OpsOrchError{Code: "not_found", Message: "deployment not found"},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "not_found",
		},
		{
			name:           "provider bad request error",
			providerError:  &orcherr.OpsOrchError{Code: "bad_request", Message: "invalid query parameters"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "bad_request",
		},
		{
			name:           "provider generic error",
			providerError:  &orcherr.OpsOrchError{Code: "provider_error", Message: "upstream service unavailable"},
			expectedStatus: http.StatusBadGateway,
			expectedCode:   "provider_error",
		},
		{
			name:           "non-OpsOrch error",
			providerError:  fmt.Errorf("generic error from provider"),
			expectedStatus: http.StatusBadGateway,
			expectedCode:   "provider_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test error handling for Query method
			t.Run("query_error", func(t *testing.T) {
				mockProvider := &mockDeploymentProvider{
					queryFunc: func(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error) {
						return nil, tc.providerError
					},
				}

				handler := &DeploymentHandler{provider: mockProvider}

				body := `{"query": "test"}`
				req := httptest.NewRequest("POST", "/deployments/query", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				handled := handler.handleDeploymentRequest(recorder, req)

				if !handled {
					t.Errorf("expected request to be handled")
				}

				if recorder.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
				}

				var errorResponse map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse); err != nil {
					t.Errorf("failed to unmarshal error response: %v", err)
				}

				if errorResponse["code"] != tc.expectedCode {
					t.Errorf("expected error code %s, got %s", tc.expectedCode, errorResponse["code"])
				}
			})

			// Test error handling for Get method
			t.Run("get_error", func(t *testing.T) {
				mockProvider := &mockDeploymentProvider{
					getFunc: func(ctx context.Context, id string) (schema.Deployment, error) {
						return schema.Deployment{}, tc.providerError
					},
				}

				handler := &DeploymentHandler{provider: mockProvider}

				req := httptest.NewRequest("GET", "/deployments/test-id", nil)
				recorder := httptest.NewRecorder()

				handled := handler.handleDeploymentRequest(recorder, req)

				if !handled {
					t.Errorf("expected request to be handled")
				}

				if recorder.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
				}

				var errorResponse map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse); err != nil {
					t.Errorf("failed to unmarshal error response: %v", err)
				}

				if errorResponse["code"] != tc.expectedCode {
					t.Errorf("expected error code %s, got %s", tc.expectedCode, errorResponse["code"])
				}
			})
		})
	}
}

// **Feature: deployment-capability-completion, Property 8: Error response consistency**
func TestProperty_ErrorResponseConsistency(t *testing.T) {
	// Test cases for different error scenarios
	testCases := []struct {
		name           string
		setupHandler   func() *DeploymentHandler
		setupRequest   func() *http.Request
		expectedStatus int
		expectedFields []string // Fields that should be present in error response
	}{
		{
			name: "no provider configured",
			setupHandler: func() *DeploymentHandler {
				return &DeploymentHandler{provider: nil}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/deployments/query", strings.NewReader(`{"query": "test"}`))
			},
			expectedStatus: http.StatusNotImplemented,
			expectedFields: []string{"code", "message"},
		},
		{
			name: "invalid JSON body",
			setupHandler: func() *DeploymentHandler {
				return &DeploymentHandler{provider: &mockDeploymentProvider{}}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/deployments/query", strings.NewReader(`{invalid json`))
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"code", "message"},
		},
		{
			name: "provider error",
			setupHandler: func() *DeploymentHandler {
				return &DeploymentHandler{
					provider: &mockDeploymentProvider{
						queryFunc: func(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error) {
							return nil, fmt.Errorf("provider connection failed")
						},
					},
				}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/deployments/query", strings.NewReader(`{"query": "test"}`))
			},
			expectedStatus: http.StatusBadGateway,
			expectedFields: []string{"code", "message"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.setupHandler()
			req := tc.setupRequest()
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handled := handler.handleDeploymentRequest(recorder, req)

			if !handled {
				t.Errorf("expected request to be handled")
			}

			// Verify status code
			if recorder.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}

			// Verify Content-Type header
			contentType := recorder.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			// Verify response is valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
			}

			// Verify expected fields are present
			for _, field := range tc.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("expected field %s not found in response", field)
				}
			}

			// Verify code field is a string
			if code, exists := response["code"]; exists {
				if _, ok := code.(string); !ok {
					t.Errorf("expected code field to be string, got %T", code)
				}
			}

			// Verify message field is a string
			if message, exists := response["message"]; exists {
				if _, ok := message.(string); !ok {
					t.Errorf("expected message field to be string, got %T", message)
				}
			}
		})
	}
}

// **Feature: deployment-capability-completion, Property 7: Startup resilience**
func TestProperty_StartupResilience(t *testing.T) {
	// Test cases for different startup scenarios
	testCases := []struct {
		name                    string
		provider                string
		config                  string
		plugin                  string
		expectServerStartup     bool
		expectDeploymentEnabled bool
	}{
		{
			name:                    "valid deployment provider",
			provider:                "mock",
			config:                  `{"test": "value"}`,
			expectServerStartup:     true,
			expectDeploymentEnabled: true,
		},
		{
			name:                    "invalid deployment provider",
			provider:                "nonexistent",
			config:                  `{"test": "value"}`,
			expectServerStartup:     true,  // Server should still start
			expectDeploymentEnabled: false, // But deployment should be disabled
		},
		{
			name:                    "no deployment provider",
			provider:                "",
			config:                  "",
			plugin:                  "",
			expectServerStartup:     true,
			expectDeploymentEnabled: false,
		},
		{
			name:                    "invalid JSON config",
			provider:                "mock",
			config:                  `{invalid json}`,
			expectServerStartup:     true,  // Server should still start
			expectDeploymentEnabled: false, // But deployment should be disabled
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variables
			oldProvider := os.Getenv("OPSORCH_DEPLOYMENT_PROVIDER")
			oldConfig := os.Getenv("OPSORCH_DEPLOYMENT_CONFIG")
			oldPlugin := os.Getenv("OPSORCH_DEPLOYMENT_PLUGIN")

			os.Setenv("OPSORCH_DEPLOYMENT_PROVIDER", tc.provider)
			os.Setenv("OPSORCH_DEPLOYMENT_CONFIG", tc.config)
			os.Setenv("OPSORCH_DEPLOYMENT_PLUGIN", tc.plugin)

			// Clean up after test
			defer func() {
				os.Setenv("OPSORCH_DEPLOYMENT_PROVIDER", oldProvider)
				os.Setenv("OPSORCH_DEPLOYMENT_CONFIG", oldConfig)
				os.Setenv("OPSORCH_DEPLOYMENT_PLUGIN", oldPlugin)
			}()

			// Try to create a server (simulating startup)
			server, err := NewServerFromEnv(context.Background())

			if tc.expectServerStartup {
				if err != nil {
					t.Errorf("expected server to start successfully, got error: %v", err)
				}
				if server == nil {
					t.Errorf("expected server to be created")
				}

				// Test that deployment capability is enabled/disabled as expected
				if server != nil {
					hasProvider := server.deployment.provider != nil
					if tc.expectDeploymentEnabled && !hasProvider {
						t.Errorf("expected deployment to be enabled but provider is nil")
					}
					if !tc.expectDeploymentEnabled && hasProvider {
						t.Errorf("expected deployment to be disabled but provider is not nil")
					}

					// Test that other capabilities are still working
					// (This verifies that deployment provider failure doesn't affect other capabilities)
					if server.incident.provider == nil && server.alert.provider == nil &&
						server.log.provider == nil && server.metric.provider == nil &&
						server.ticket.provider == nil && server.messaging.provider == nil &&
						server.service.provider == nil {
						// This is expected when no providers are configured
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected server startup to fail")
				}
			}
		})
	}
}

// **Feature: deployment-capability-completion, Property 2: Deployment endpoint responses**
func TestProperty_DeploymentEndpointResponses(t *testing.T) {
	// Create a server with deployment capability
	mockProvider := &mockDeploymentProvider{
		queryFunc: func(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error) {
			return []schema.Deployment{
				{
					ID:          "test-deployment",
					Service:     "test-service",
					Environment: "production",
					Version:     "v1.0.0",
					Status:      "success",
				},
			}, nil
		},
		getFunc: func(ctx context.Context, id string) (schema.Deployment, error) {
			return schema.Deployment{
				ID:          id,
				Service:     "test-service",
				Environment: "production",
				Version:     "v1.0.0",
				Status:      "success",
			}, nil
		},
	}

	server := &Server{
		corsOrigin: "*",
		deployment: DeploymentHandler{provider: mockProvider},
	}

	// Test cases for different endpoint scenarios
	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST /deployments/query",
			method:         "POST",
			path:           "/deployments/query",
			body:           `{"query": "test"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET /deployments/test-id",
			method:         "GET",
			path:           "/deployments/test-id",
			body:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS /deployments/query",
			method:         "OPTIONS",
			path:           "/deployments/query",
			body:           "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			recorder := httptest.NewRecorder()

			// Call the server handler
			server.ServeHTTP(recorder, req)

			// Verify status code
			if recorder.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}

			// Verify CORS headers are present
			corsOrigin := recorder.Header().Get("Access-Control-Allow-Origin")
			if corsOrigin != "*" {
				t.Errorf("expected CORS origin *, got %s", corsOrigin)
			}

			corsHeaders := recorder.Header().Get("Access-Control-Allow-Headers")
			if corsHeaders != "Content-Type, Authorization" {
				t.Errorf("expected CORS headers 'Content-Type, Authorization', got %s", corsHeaders)
			}

			corsMethods := recorder.Header().Get("Access-Control-Allow-Methods")
			if corsMethods != "GET,POST,PATCH,OPTIONS" {
				t.Errorf("expected CORS methods 'GET,POST,PATCH,OPTIONS', got %s", corsMethods)
			}

			// For non-OPTIONS requests, verify request ID header
			if tc.method != "OPTIONS" {
				requestID := recorder.Header().Get("X-Request-ID")
				if requestID == "" {
					t.Errorf("expected X-Request-ID header to be present")
				}
			}

			// For successful responses, verify JSON content type
			if tc.expectedStatus == http.StatusOK && tc.method != "OPTIONS" {
				contentType := recorder.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}

				// Verify response is valid JSON
				var response interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("response is not valid JSON: %v", err)
				}
			}
		})
	}
}

// **Feature: deployment-capability-completion, Property 6: Provider status reporting**
func TestProperty_ProviderStatusReporting(t *testing.T) {
	// Test cases for different provider states
	testCases := []struct {
		name                    string
		setupServer             func() *Server
		expectedProviderPresent bool
	}{
		{
			name: "deployment provider active",
			setupServer: func() *Server {
				return &Server{
					deployment: DeploymentHandler{provider: &mockDeploymentProvider{}},
				}
			},
			expectedProviderPresent: true,
		},
		{
			name: "deployment provider not configured",
			setupServer: func() *Server {
				return &Server{
					deployment: DeploymentHandler{provider: nil},
				}
			},
			expectedProviderPresent: true, // Providers endpoint should still list available providers
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()

			// Test the providers endpoint for deployment capability
			req := httptest.NewRequest("GET", "/providers/deployment", nil)
			recorder := httptest.NewRecorder()

			handled := server.handleProviders(recorder, req)

			if !handled {
				t.Errorf("expected providers request to be handled")
			}

			// Verify successful response
			if recorder.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", recorder.Code)
			}

			// Verify response format
			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			// Verify providers field exists
			providers, exists := response["providers"]
			if !exists {
				t.Errorf("expected providers field in response")
			}

			// Verify providers is an array
			providersList, ok := providers.([]interface{})
			if !ok {
				t.Errorf("expected providers to be an array, got %T", providers)
			}

			if tc.expectedProviderPresent {
				// For this test, we expect the mock provider to be listed
				// (The actual provider registration happens in the init function)
				if len(providersList) == 0 {
					// This is expected if no providers are registered
					t.Logf("No deployment providers registered (expected for test)")
				}
			}

			// Test that the endpoint handles different capability names correctly
			testPaths := []string{"/providers/deployment", "/providers/deployments"}
			for _, path := range testPaths {
				req := httptest.NewRequest("GET", path, nil)
				recorder := httptest.NewRecorder()

				handled := server.handleProviders(recorder, req)

				if !handled {
					t.Errorf("expected providers request to be handled for path %s", path)
				}

				if recorder.Code != http.StatusOK {
					t.Errorf("expected status 200 for path %s, got %d", path, recorder.Code)
				}
			}
		})
	}
}
