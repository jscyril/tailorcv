package domain

import "testing"

func TestGenerateAITailoringInputValidation(t *testing.T) {
	valid, err := (GenerateAITailoringInput{
		JobID: "job", SelectedFactIDs: []string{"fact"}, Model: " qwen3:8b ",
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if valid.Provider != "ollama" || valid.Model != "qwen3:8b" || valid.Endpoint != DefaultOllamaEndpoint {
		t.Fatalf("validated input = %#v", valid)
	}
	if _, err := (GenerateAITailoringInput{JobID: "job", SelectedFactIDs: []string{"fact"}, Model: "model", Provider: "gemini"}).Validate(); err == nil {
		t.Fatal("Validate() accepted an unsupported provider")
	}
}

func TestValidateOllamaEndpoint(t *testing.T) {
	endpoint, err := ValidateOllamaEndpoint(" http://localhost:11434/ ")
	if err != nil || endpoint != "http://localhost:11434" {
		t.Fatalf("ValidateOllamaEndpoint() = %q, %v", endpoint, err)
	}
	for _, invalid := range []string{"file:///tmp/ollama", "http://user:secret@localhost:11434", "http://localhost:11434?token=secret"} {
		if _, err := ValidateOllamaEndpoint(invalid); err == nil {
			t.Fatalf("ValidateOllamaEndpoint(%q) expected error", invalid)
		}
	}
}
