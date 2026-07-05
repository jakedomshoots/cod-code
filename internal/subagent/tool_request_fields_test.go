package subagent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_ToolRequest_UnmarshalJSON_accepts_bare_path_string(t *testing.T) {
	// Given
	var request ToolRequest

	// When
	err := json.Unmarshal([]byte(`"app/main.go"`), &request)

	// Then
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if request.Action != "" || request.Path != "app/main.go" {
		t.Fatalf("request = %+v, want bare path with empty action", request)
	}
}

func Test_ToolRequest_UnmarshalJSON_round_trips_url_app_tool_and_max_fields(t *testing.T) {
	// Given
	raw := `{"action":"browser_read","url":"http://127.0.0.1:8081/","max_bytes":2048}`

	// When
	var request ToolRequest
	err := json.Unmarshal([]byte(raw), &request)

	// Then
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if request.Action != "browser_read" {
		t.Fatalf("Action = %q, want browser_read", request.Action)
	}
	if request.URL != "http://127.0.0.1:8081/" {
		t.Fatalf("URL = %q, want loopback URL", request.URL)
	}
	if request.MaxBytes != 2048 {
		t.Fatalf("MaxBytes = %d, want 2048", request.MaxBytes)
	}
	if request.App != "" {
		t.Fatalf("App = %q, want empty when unset", request.App)
	}
}

func Test_ToolRequest_UnmarshalJSON_accepts_computer_snapshot_app_field(t *testing.T) {
	// Given
	raw := `{"action":"computer_snapshot","app":"Safari","max_bytes":4096}`

	// When
	var request ToolRequest
	err := json.Unmarshal([]byte(raw), &request)

	// Then
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if request.Action != "computer_snapshot" {
		t.Fatalf("Action = %q, want computer_snapshot", request.Action)
	}
	if request.App != "Safari" {
		t.Fatalf("App = %q, want Safari", request.App)
	}
	if request.MaxBytes != 4096 {
		t.Fatalf("MaxBytes = %d, want 4096", request.MaxBytes)
	}
}

func Test_ToolRequest_UnmarshalJSON_accepts_tool_manifest_tool_field(t *testing.T) {
	// Given
	raw := `{"action":"tool_manifest","tool":"tools.manifest"}`

	// When
	var request ToolRequest
	err := json.Unmarshal([]byte(raw), &request)

	// Then
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if request.Action != "tool_manifest" {
		t.Fatalf("Action = %q, want tool_manifest", request.Action)
	}
	if request.Tool != "tools.manifest" {
		t.Fatalf("Tool = %q, want tools.manifest", request.Tool)
	}
}

func Test_NormalizedToolRequestAction_infers_browser_read_when_only_url_is_supplied(t *testing.T) {
	// Given
	table := []struct {
		name    string
		request ToolRequest
		want    string
	}{
		{
			name:    "url only infers browser_read",
			request: ToolRequest{URL: "http://127.0.0.1:8080/"},
			want:    "browser_read",
		},
		{
			name:    "path only infers read_workspace",
			request: ToolRequest{Path: "app/main.go"},
			want:    "read_workspace",
		},
		{
			name:    "query only infers search_workspace",
			request: ToolRequest{Query: "needle"},
			want:    "search_workspace",
		},
		{
			name:    "explicit action wins over inferred",
			request: ToolRequest{Action: "computer_snapshot", App: "Finder", URL: "http://example.com"},
			want:    "computer_snapshot",
		},
		{
			name:    "app without action stays empty (no inference rule)",
			request: ToolRequest{App: "Finder"},
			want:    "",
		},
	}

	// When / Then
	for _, tc := range table {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := normalizedToolRequestAction(tc.request)
			if got != tc.want {
				t.Fatalf("normalizedToolRequestAction(%+v) = %q, want %q", tc.request, got, tc.want)
			}
		})
	}
}

func Test_ParseToolRequests_returns_browser_read_request_inferred_from_url(t *testing.T) {
	// Given
	payload := `{"status":"pass","summary":"inspect page","tool_requests":[{"url":"http://127.0.0.1:65535/page"}]}`

	// When
	requests, err := ParseToolRequests(payload)

	// Then
	if err != nil {
		t.Fatalf("ParseToolRequests returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests length = %d, want 1", len(requests))
	}
	got := requests[0]
	if got.Action != "browser_read" {
		t.Fatalf("Action = %q, want browser_read inferred from URL", got.Action)
	}
	if got.URL != "http://127.0.0.1:65535/page" {
		t.Fatalf("URL = %q, want loopback URL preserved", got.URL)
	}
}

func Test_ToolResult_RoundTrip_preserves_permission_and_receipt_sha256_fields(t *testing.T) {
	// Given
	result := ToolResult{
		Action:        "browser_read",
		Status:        "pass",
		URL:           "http://127.0.0.1:8080/",
		Permission:    "allow-localhost",
		ReceiptSHA256: "deadbeef",
	}

	// When
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var decoded ToolResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	// Then
	if !reflect.DeepEqual(decoded, result) {
		t.Fatalf("decoded = %+v, want %+v", decoded, result)
	}
}
