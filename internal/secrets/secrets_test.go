package secrets

import (
	"testing"
)

func TestParseEnvMap_StringsPassThrough(t *testing.T) {
	m, err := ParseEnvMap(`{"LLM_API_KEY":"sk-","LLM_BASE_URL":"https://api.openai.com/v1"}`)
	if err != nil {
		t.Fatalf("ParseEnvMap: %v", err)
	}
	if m["LLM_API_KEY"] != "sk-" {
		t.Errorf("LLM_API_KEY = %q", m["LLM_API_KEY"])
	}
	if m["LLM_BASE_URL"] != "https://api.openai.com/v1" {
		t.Errorf("LLM_BASE_URL = %q", m["LLM_BASE_URL"])
	}
}

func TestParseEnvMap_NumericValues(t *testing.T) {
	m, err := ParseEnvMap(`{"LLM_MAX_TOKENS":4000,"LLM_TIMEOUT_SEC":120}`)
	if err != nil {
		t.Fatalf("ParseEnvMap: %v", err)
	}
	if m["LLM_MAX_TOKENS"] != "4000" {
		t.Errorf("LLM_MAX_TOKENS = %q, want 4000", m["LLM_MAX_TOKENS"])
	}
	if m["LLM_TIMEOUT_SEC"] != "120" {
		t.Errorf("LLM_TIMEOUT_SEC = %q, want 120", m["LLM_TIMEOUT_SEC"])
	}
}

func TestParseEnvMap_BoolValue(t *testing.T) {
	m, err := ParseEnvMap(`{"FEATURE_FLAG":true}`)
	if err != nil {
		t.Fatalf("ParseEnvMap: %v", err)
	}
	if m["FEATURE_FLAG"] != "true" {
		t.Errorf("FEATURE_FLAG = %q, want true", m["FEATURE_FLAG"])
	}
}

func TestParseEnvMap_NestedObjectRejected(t *testing.T) {
	if _, err := ParseEnvMap(`{"NESTED":{"a":1}}`); err == nil {
		t.Fatal("ParseEnvMap succeeded, want error for nested object")
	}
}

func TestParseEnvMap_ArrayRejected(t *testing.T) {
	if _, err := ParseEnvMap(`{"ITEMS":[1,2,3]}`); err == nil {
		t.Fatal("ParseEnvMap succeeded, want error for array")
	}
}

func TestParseEnvMap_NullValueSkipped(t *testing.T) {
	m, err := ParseEnvMap(`{"A":"x","B":null}`)
	if err != nil {
		t.Fatalf("ParseEnvMap: %v", err)
	}
	if _, ok := m["B"]; ok {
		t.Error("B should be absent for null value")
	}
	if m["A"] != "x" {
		t.Errorf("A = %q", m["A"])
	}
}

func TestParseEnvMap_InvalidJSONRejected(t *testing.T) {
	if _, err := ParseEnvMap("not-json"); err == nil {
		t.Fatal("ParseEnvMap succeeded, want error for invalid JSON")
	}
}
