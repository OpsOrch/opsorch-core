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

	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
	"github.com/opsorch/opsorch-core/team"
)

// Mock team provider for testing
type mockTeamProvider struct {
	queryFunc   func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error)
	getFunc     func(ctx context.Context, id string) (schema.Team, error)
	membersFunc func(ctx context.Context, teamID string) ([]schema.TeamMember, error)
}

func (m *mockTeamProvider) Query(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query)
	}
	return []schema.Team{}, nil
}

func (m *mockTeamProvider) Get(ctx context.Context, id string) (schema.Team, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return schema.Team{}, nil
}

func (m *mockTeamProvider) Members(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
	if m.membersFunc != nil {
		return m.membersFunc(ctx, teamID)
	}
	return []schema.TeamMember{}, nil
}

// Register mock provider for testing
func init() {
	team.RegisterProvider("mock", func(config map[string]any) (team.Provider, error) {
		return &mockTeamProvider{}, nil
	})
}

// handleTeamRequest is a helper to test team handler directly
func (h *TeamHandler) handleTeamRequest(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/teams") {
		return false
	}
	if h.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "team_provider_missing", Message: "team provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.TeamQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		teams, err := h.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.query")
		writeJSON(w, http.StatusOK, teams)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		team, err := h.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.get")
		writeJSON(w, http.StatusOK, team)
		return true
	case len(segments) == 3 && segments[2] == "members" && r.Method == http.MethodGet:
		teamID := segments[1]
		members, err := h.provider.Members(r.Context(), teamID)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.members")
		writeJSON(w, http.StatusOK, members)
		return true
	default:
		return false
	}
}

// **Feature: team-capability, Property: Environment configuration processing**
func TestTeamProperty_EnvironmentConfigurationProcessing(t *testing.T) {
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
			oldProvider := os.Getenv("OPSORCH_TEAM_PROVIDER")
			oldConfig := os.Getenv("OPSORCH_TEAM_CONFIG")
			oldPlugin := os.Getenv("OPSORCH_TEAM_PLUGIN")

			os.Setenv("OPSORCH_TEAM_PROVIDER", tc.provider)
			os.Setenv("OPSORCH_TEAM_CONFIG", tc.config)
			os.Setenv("OPSORCH_TEAM_PLUGIN", tc.plugin)

			defer func() {
				os.Setenv("OPSORCH_TEAM_PROVIDER", oldProvider)
				os.Setenv("OPSORCH_TEAM_CONFIG", oldConfig)
				os.Setenv("OPSORCH_TEAM_PLUGIN", oldPlugin)
			}()

			mockSec := &mockSecretProvider{}
			handler, err := newTeamHandlerFromEnv(mockSec)

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

// **Feature: team-capability, Property: Team query body processing**
func TestTeamProperty_TeamQueryBodyProcessing(t *testing.T) {
	mockProvider := &mockTeamProvider{
		queryFunc: func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
			teams := []schema.Team{
				{
					ID:   "team-1",
					Name: query.Name,
					Tags: query.Tags,
				},
			}
			return teams, nil
		},
	}

	handler := &TeamHandler{provider: mockProvider}

	testCases := []struct {
		name  string
		query schema.TeamQuery
	}{
		{
			name: "query with name filter",
			query: schema.TeamQuery{
				Name: "backend",
			},
		},
		{
			name: "query with tags filter",
			query: schema.TeamQuery{
				Tags: map[string]string{
					"department": "engineering",
					"region":     "us-west",
				},
			},
		},
		{
			name: "query with scope",
			query: schema.TeamQuery{
				Scope: schema.QueryScope{
					Service:     "api-service",
					Environment: "production",
				},
			},
		},
		{
			name: "query with metadata",
			query: schema.TeamQuery{
				Metadata: map[string]any{
					"custom_field": "value",
				},
			},
		},
		{
			name: "complex query with all fields",
			query: schema.TeamQuery{
				Name: "platform",
				Tags: map[string]string{
					"team_type": "core",
				},
				Scope: schema.QueryScope{
					Service:     "payment-service",
					Environment: "staging",
					Team:        "backend",
				},
				Metadata: map[string]any{
					"oncall": true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.query)
			req := httptest.NewRequest("POST", "/teams/query", strings.NewReader(string(bodyBytes)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handled := handler.handleTeamRequest(recorder, req)

			if !handled {
				t.Errorf("expected request to be handled")
			}

			if recorder.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", recorder.Code)
			}

			var teams []schema.Team
			if err := json.Unmarshal(recorder.Body.Bytes(), &teams); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if len(teams) == 0 {
				t.Errorf("expected at least one team in response")
			}

			// Verify that name filter was passed correctly
			if tc.query.Name != "" && teams[0].Name != tc.query.Name {
				t.Errorf("expected name %s, got %s", tc.query.Name, teams[0].Name)
			}
		})
	}
}

// **Feature: team-capability, Property: Team ID retrieval**
func TestTeamProperty_TeamIDRetrieval(t *testing.T) {
	testCases := []struct {
		name        string
		teamID      string
		mockTeam    schema.Team
		expectError bool
	}{
		{
			name:   "valid team ID",
			teamID: "team-123",
			mockTeam: schema.Team{
				ID:     "team-123",
				Name:   "Backend Team",
				Parent: "engineering",
				Tags:   map[string]string{"department": "engineering"},
			},
			expectError: false,
		},
		{
			name:   "team ID with special characters",
			teamID: "team-abc-123_def",
			mockTeam: schema.Team{
				ID:   "team-abc-123_def",
				Name: "Platform Team",
			},
			expectError: false,
		},
		{
			name:   "long team ID",
			teamID: "very-long-team-id-with-many-characters-12345678901234567890",
			mockTeam: schema.Team{
				ID:   "very-long-team-id-with-many-characters-12345678901234567890",
				Name: "Data Science Team",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProvider := &mockTeamProvider{
				getFunc: func(ctx context.Context, id string) (schema.Team, error) {
					if id == tc.teamID {
						return tc.mockTeam, nil
					}
					return schema.Team{}, fmt.Errorf("team not found")
				},
			}

			handler := &TeamHandler{provider: mockProvider}

			req := httptest.NewRequest("GET", fmt.Sprintf("/teams/%s", tc.teamID), nil)
			recorder := httptest.NewRecorder()

			handled := handler.handleTeamRequest(recorder, req)

			if !handled {
				t.Errorf("expected request to be handled")
			}

			if tc.expectError {
				if recorder.Code == http.StatusOK {
					t.Errorf("expected error response, got status 200")
				}
			} else {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", recorder.Code)
				}

				var team schema.Team
				if err := json.Unmarshal(recorder.Body.Bytes(), &team); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}

				if team.ID != tc.teamID {
					t.Errorf("expected team ID %s, got %s", tc.teamID, team.ID)
				}
				if team.Name != tc.mockTeam.Name {
					t.Errorf("expected name %s, got %s", tc.mockTeam.Name, team.Name)
				}
				if team.Parent != tc.mockTeam.Parent {
					t.Errorf("expected parent %s, got %s", tc.mockTeam.Parent, team.Parent)
				}
			}
		})
	}
}

// **Feature: team-capability, Property: Team members retrieval**
func TestTeamProperty_TeamMembersRetrieval(t *testing.T) {
	testCases := []struct {
		name        string
		teamID      string
		mockMembers []schema.TeamMember
		expectError bool
	}{
		{
			name:   "team with multiple members",
			teamID: "team-123",
			mockMembers: []schema.TeamMember{
				{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: "owner"},
				{ID: "user-2", Name: "Bob", Email: "bob@example.com", Role: "member"},
				{ID: "user-3", Name: "Charlie", Email: "charlie@example.com", Role: "oncall"},
			},
			expectError: false,
		},
		{
			name:        "team with no members",
			teamID:      "empty-team",
			mockMembers: []schema.TeamMember{},
			expectError: false,
		},
		{
			name:   "team with member handles",
			teamID: "team-456",
			mockMembers: []schema.TeamMember{
				{
					ID:     "user-4",
					Name:   "Diana",
					Email:  "diana@example.com",
					Handle: "@diana-slack",
					Role:   "manager",
					Metadata: map[string]any{
						"slack_id": "U1234567890",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockProvider := &mockTeamProvider{
				membersFunc: func(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
					if teamID == tc.teamID {
						return tc.mockMembers, nil
					}
					return nil, fmt.Errorf("team not found")
				},
			}

			handler := &TeamHandler{provider: mockProvider}

			req := httptest.NewRequest("GET", fmt.Sprintf("/teams/%s/members", tc.teamID), nil)
			recorder := httptest.NewRecorder()

			handled := handler.handleTeamRequest(recorder, req)

			if !handled {
				t.Errorf("expected request to be handled")
			}

			if tc.expectError {
				if recorder.Code == http.StatusOK {
					t.Errorf("expected error response, got status 200")
				}
			} else {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", recorder.Code)
				}

				var members []schema.TeamMember
				if err := json.Unmarshal(recorder.Body.Bytes(), &members); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}

				if len(members) != len(tc.mockMembers) {
					t.Errorf("expected %d members, got %d", len(tc.mockMembers), len(members))
				}

				for i, member := range members {
					if member.ID != tc.mockMembers[i].ID {
						t.Errorf("expected member ID %s, got %s", tc.mockMembers[i].ID, member.ID)
					}
					if member.Name != tc.mockMembers[i].Name {
						t.Errorf("expected name %s, got %s", tc.mockMembers[i].Name, member.Name)
					}
					if member.Email != tc.mockMembers[i].Email {
						t.Errorf("expected email %s, got %s", tc.mockMembers[i].Email, member.Email)
					}
					if member.Role != tc.mockMembers[i].Role {
						t.Errorf("expected role %s, got %s", tc.mockMembers[i].Role, member.Role)
					}
				}
			}
		})
	}
}

// **Feature: team-capability, Property: Provider error handling**
func TestTeamProperty_ProviderErrorHandling(t *testing.T) {
	testCases := []struct {
		name           string
		providerError  error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "provider not found error",
			providerError:  &orcherr.OpsOrchError{Code: "not_found", Message: "team not found"},
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
				mockProvider := &mockTeamProvider{
					queryFunc: func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
						return nil, tc.providerError
					},
				}

				handler := &TeamHandler{provider: mockProvider}

				body := `{"name": "test"}`
				req := httptest.NewRequest("POST", "/teams/query", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				handled := handler.handleTeamRequest(recorder, req)

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
				mockProvider := &mockTeamProvider{
					getFunc: func(ctx context.Context, id string) (schema.Team, error) {
						return schema.Team{}, tc.providerError
					},
				}

				handler := &TeamHandler{provider: mockProvider}

				req := httptest.NewRequest("GET", "/teams/test-id", nil)
				recorder := httptest.NewRecorder()

				handled := handler.handleTeamRequest(recorder, req)

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

			// Test error handling for Members method
			t.Run("members_error", func(t *testing.T) {
				mockProvider := &mockTeamProvider{
					membersFunc: func(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
						return nil, tc.providerError
					},
				}

				handler := &TeamHandler{provider: mockProvider}

				req := httptest.NewRequest("GET", "/teams/test-id/members", nil)
				recorder := httptest.NewRecorder()

				handled := handler.handleTeamRequest(recorder, req)

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

// **Feature: team-capability, Property: Error response consistency**
func TestTeamProperty_ErrorResponseConsistency(t *testing.T) {
	testCases := []struct {
		name           string
		setupHandler   func() *TeamHandler
		setupRequest   func() *http.Request
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "no provider configured",
			setupHandler: func() *TeamHandler {
				return &TeamHandler{provider: nil}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/teams/query", strings.NewReader(`{"name": "test"}`))
			},
			expectedStatus: http.StatusNotImplemented,
			expectedFields: []string{"code", "message"},
		},
		{
			name: "invalid JSON body",
			setupHandler: func() *TeamHandler {
				return &TeamHandler{provider: &mockTeamProvider{}}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/teams/query", strings.NewReader(`{invalid json`))
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"code", "message"},
		},
		{
			name: "provider error",
			setupHandler: func() *TeamHandler {
				return &TeamHandler{
					provider: &mockTeamProvider{
						queryFunc: func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
							return nil, fmt.Errorf("provider connection failed")
						},
					},
				}
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/teams/query", strings.NewReader(`{"name": "test"}`))
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

			handled := handler.handleTeamRequest(recorder, req)

			if !handled {
				t.Errorf("expected request to be handled")
			}

			if recorder.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}

			contentType := recorder.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
			}

			for _, field := range tc.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("expected field %s not found in response", field)
				}
			}

			if code, exists := response["code"]; exists {
				if _, ok := code.(string); !ok {
					t.Errorf("expected code field to be string, got %T", code)
				}
			}

			if message, exists := response["message"]; exists {
				if _, ok := message.(string); !ok {
					t.Errorf("expected message field to be string, got %T", message)
				}
			}
		})
	}
}

// **Feature: team-capability, Property: Startup resilience**
func TestTeamProperty_StartupResilience(t *testing.T) {
	testCases := []struct {
		name                string
		provider            string
		config              string
		plugin              string
		expectServerStartup bool
		expectTeamEnabled   bool
	}{
		{
			name:                "valid team provider",
			provider:            "mock",
			config:              `{"test": "value"}`,
			expectServerStartup: true,
			expectTeamEnabled:   true,
		},
		{
			name:                "invalid team provider",
			provider:            "nonexistent",
			config:              `{"test": "value"}`,
			expectServerStartup: true,  // Server should still start
			expectTeamEnabled:   false, // But team should be disabled
		},
		{
			name:                "no team provider",
			provider:            "",
			config:              "",
			plugin:              "",
			expectServerStartup: true,
			expectTeamEnabled:   false,
		},
		{
			name:                "invalid JSON config",
			provider:            "mock",
			config:              `{invalid json}`,
			expectServerStartup: true,  // Server should still start
			expectTeamEnabled:   false, // But team should be disabled
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldProvider := os.Getenv("OPSORCH_TEAM_PROVIDER")
			oldConfig := os.Getenv("OPSORCH_TEAM_CONFIG")
			oldPlugin := os.Getenv("OPSORCH_TEAM_PLUGIN")

			os.Setenv("OPSORCH_TEAM_PROVIDER", tc.provider)
			os.Setenv("OPSORCH_TEAM_CONFIG", tc.config)
			os.Setenv("OPSORCH_TEAM_PLUGIN", tc.plugin)

			defer func() {
				os.Setenv("OPSORCH_TEAM_PROVIDER", oldProvider)
				os.Setenv("OPSORCH_TEAM_CONFIG", oldConfig)
				os.Setenv("OPSORCH_TEAM_PLUGIN", oldPlugin)
			}()

			server, err := NewServerFromEnv(context.Background())

			if tc.expectServerStartup {
				if err != nil {
					t.Errorf("expected server to start successfully, got error: %v", err)
				}
				if server == nil {
					t.Errorf("expected server to be created")
				}

				if server != nil {
					hasProvider := server.team.provider != nil
					if tc.expectTeamEnabled && !hasProvider {
						t.Errorf("expected team to be enabled but provider is nil")
					}
					if !tc.expectTeamEnabled && hasProvider {
						t.Errorf("expected team to be disabled but provider is not nil")
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

// **Feature: team-capability, Property: Endpoint routing**
func TestTeamProperty_EndpointRouting(t *testing.T) {
	mockProvider := &mockTeamProvider{
		queryFunc: func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
			return []schema.Team{{ID: "team-1", Name: "Test Team"}}, nil
		},
		getFunc: func(ctx context.Context, id string) (schema.Team, error) {
			return schema.Team{ID: id, Name: "Test Team"}, nil
		},
		membersFunc: func(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
			return []schema.TeamMember{{ID: "user-1", Name: "Test User"}}, nil
		},
	}

	server := &Server{
		corsOrigin: "*",
		team:       TeamHandler{provider: mockProvider},
	}

	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectHandled  bool
	}{
		{
			name:           "POST /teams/query",
			method:         "POST",
			path:           "/teams/query",
			body:           `{"name": "test"}`,
			expectedStatus: http.StatusOK,
			expectHandled:  true,
		},
		{
			name:           "GET /teams/{id}",
			method:         "GET",
			path:           "/teams/team-123",
			expectedStatus: http.StatusOK,
			expectHandled:  true,
		},
		{
			name:           "GET /teams/{id}/members",
			method:         "GET",
			path:           "/teams/team-123/members",
			expectedStatus: http.StatusOK,
			expectHandled:  true,
		},
		{
			name:           "invalid method for query",
			method:         "GET",
			path:           "/teams/query",
			expectedStatus: http.StatusOK, // Will be handled as GET /teams/{id} where id="query"
			expectHandled:  true,
		},
		{
			name:           "non-team path returns false",
			method:         "GET",
			path:           "/incidents/123",
			expectedStatus: 0,
			expectHandled:  false,
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

			handled := server.handleTeam(recorder, req)

			if handled != tc.expectHandled {
				t.Errorf("expected handled=%v, got %v", tc.expectHandled, handled)
			}

			if tc.expectHandled && recorder.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}
		})
	}
}

// **Feature: team-capability, Property: Full server integration**
func TestTeamProperty_FullServerIntegration(t *testing.T) {
	mockProvider := &mockTeamProvider{
		queryFunc: func(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
			return []schema.Team{
				{
					ID:     "team-backend",
					Name:   "Backend Team",
					Parent: "engineering",
					Tags:   map[string]string{"department": "engineering"},
				},
			}, nil
		},
		getFunc: func(ctx context.Context, id string) (schema.Team, error) {
			return schema.Team{
				ID:     id,
				Name:   "Backend Team",
				Parent: "engineering",
			}, nil
		},
		membersFunc: func(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
			return []schema.TeamMember{
				{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: "owner"},
				{ID: "user-2", Name: "Bob", Email: "bob@example.com", Role: "member"},
			}, nil
		},
	}

	server := &Server{
		corsOrigin: "*",
		team:       TeamHandler{provider: mockProvider},
	}

	// Test query endpoint through full server
	t.Run("query through server", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/teams/query", strings.NewReader(`{"name": "backend"}`))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var teams []schema.Team
		if err := json.Unmarshal(recorder.Body.Bytes(), &teams); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		if len(teams) != 1 {
			t.Errorf("expected 1 team, got %d", len(teams))
		}
	})

	// Test get endpoint through full server
	t.Run("get through server", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/teams/team-backend", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var team schema.Team
		if err := json.Unmarshal(recorder.Body.Bytes(), &team); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		if team.ID != "team-backend" {
			t.Errorf("expected team ID team-backend, got %s", team.ID)
		}
	})

	// Test members endpoint through full server
	t.Run("members through server", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/teams/team-backend/members", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var members []schema.TeamMember
		if err := json.Unmarshal(recorder.Body.Bytes(), &members); err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		if len(members) != 2 {
			t.Errorf("expected 2 members, got %d", len(members))
		}
	})

	// Test CORS headers
	t.Run("CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/teams/query", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200 for OPTIONS, got %d", recorder.Code)
		}

		corsOrigin := recorder.Header().Get("Access-Control-Allow-Origin")
		if corsOrigin != "*" {
			t.Errorf("expected CORS origin *, got %s", corsOrigin)
		}
	})
}
