package model

import "testing"

func Test_JSONPayload_extracts_embedded_json_object(t *testing.T) {
	payload, ok := JSONPayload("Sure, here is the result:\n{\"summary\":\"done\",\"patches\":[]}\nThanks.")
	if !ok {
		t.Fatal("JSONPayload ok = false, want true")
	}
	if payload != `{"summary":"done","patches":[]}` {
		t.Fatalf("payload = %q, want embedded JSON object", payload)
	}
}

func Test_JSONPayload_keeps_braces_inside_strings(t *testing.T) {
	payload, ok := JSONPayload("```json\n{\"summary\":\"keeps { braces } in strings\",\"patches\":[]}\n```")
	if !ok {
		t.Fatal("JSONPayload ok = false, want true")
	}
	if payload != `{"summary":"keeps { braces } in strings","patches":[]}` {
		t.Fatalf("payload = %q, want complete fenced JSON object", payload)
	}
}

func Test_JSONPayload_ignores_text_without_json_object(t *testing.T) {
	if payload, ok := JSONPayload("plain provider prose"); ok || payload != "" {
		t.Fatalf("JSONPayload = %q, %v, want empty false", payload, ok)
	}
}
