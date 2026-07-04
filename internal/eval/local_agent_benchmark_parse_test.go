package eval

import "testing"

func Test_ParseLocalAgentBenchmarkAgentTimeouts_validatesInput(t *testing.T) {
	// When
	timeouts, err := parseLocalAgentBenchmarkAgentTimeouts("opencode=600, pi = 360")
	// Then
	if err != nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentTimeouts returned error: %v", err)
	}
	if timeouts["opencode"] != 600 || timeouts["pi"] != 360 {
		t.Fatalf("timeouts = %+v, want parsed overrides", timeouts)
	}
	if _, err := parseLocalAgentBenchmarkAgentTimeouts("opencode=0"); err == nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentTimeouts accepted non-positive timeout")
	}
	if _, err := parseLocalAgentBenchmarkAgentTimeouts("opencode"); err == nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentTimeouts accepted malformed timeout")
	}
}

func Test_ParseLocalAgentBenchmarkAgentModels_validatesInput(t *testing.T) {
	// When
	models, err := parseLocalAgentBenchmarkAgentModels("opencode=openai/gpt-5.4-mini, pi = kimi/k2")
	// Then
	if err != nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentModels returned error: %v", err)
	}
	if models["opencode"] != "openai/gpt-5.4-mini" || models["pi"] != "kimi/k2" {
		t.Fatalf("models = %+v, want parsed overrides", models)
	}
	if _, err := parseLocalAgentBenchmarkAgentModels("opencode"); err == nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentModels accepted malformed model")
	}
	if _, err := parseLocalAgentBenchmarkAgentModels("opencode="); err == nil {
		t.Fatalf("parseLocalAgentBenchmarkAgentModels accepted empty model")
	}
}
