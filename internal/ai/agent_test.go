package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func agentTestRequest() Request {
	return Request{SchemaVersion: SchemaVersion, Job: JobRequirements{Role: "Engineer"}, Facts: []Fact{{ID: "f1", Text: "Built Go deployment pipelines", Technologies: []string{"Go"}}}}
}

func TestRunProposalAgentRetriesAfterValidationErrors(t *testing.T) {
	var calls []int
	var feedback map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls = append(calls, len(body))
		text := "Built Go deployment pipelines for 99 teams"
		if len(calls) > 1 {
			var sent struct {
				Contents []struct {
					Parts []struct {
						FunctionResponse *struct {
							Response map[string]any `json:"response"`
						} `json:"functionResponse"`
					} `json:"parts"`
				} `json:"contents"`
			}
			_ = json.Unmarshal(body, &sent)
			feedback = sent.Contents[len(sent.Contents)-1].Parts[0].FunctionResponse.Response
			text = "Built Go deployment pipelines"
		}
		args, _ := json.Marshal(map[string]any{"proposals": []map[string]any{{"targetFactId": "f1", "supportingFactIds": []string{"f1"}, "text": text}}})
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":"submit_proposals","args":` + string(args) + `}}]}}]}`))
	}))
	defer server.Close()
	provider, err := newGemini("key", server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := RunProposalAgent(context.Background(), provider, "m", agentTestRequest())
	if err != nil {
		t.Fatalf("RunProposalAgent() error = %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("provider calls = %d, want 2", len(calls))
	}
	if errs, _ := feedback["errors"].([]any); len(errs) == 0 || !strings.Contains(errs[0].(string), "metrics") {
		t.Fatalf("agent feedback = %#v", feedback)
	}
	if result := DecodeAndValidate(agentTestRequest(), raw); len(result.Errors) > 0 {
		t.Fatalf("final proposals invalid: %v", result.Errors)
	}
}

func TestGeminiAgentStopsWithoutToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"done"}]}}]}`))
	}))
	defer server.Close()
	provider, _ := newGemini("key", server.URL, server.Client())
	if _, err := RunProposalAgent(context.Background(), provider, "m", agentTestRequest()); err == nil {
		t.Fatal("expected error when the model never calls a tool")
	}
}
