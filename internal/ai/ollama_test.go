package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestOllamaModelsAndGenerateContract(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/api/tags":
			return jsonResponse(http.StatusOK, `{"models":[{"name":"qwen3:8b"},{"model":"gemma3:4b"}]}`), nil
		case "/api/generate":
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if body["stream"] != false || body["format"] == nil || !strings.Contains(body["prompt"].(string), SchemaVersion) {
				t.Fatalf("generation request = %#v", body)
			}
			return jsonResponse(http.StatusOK, `{"response":"{\"schemaVersion\":\"tailorcv.ai.tailoring.v1\",\"proposals\":[]}"}`), nil
		default:
			t.Fatalf("unexpected path %q", request.URL.Path)
			return nil, nil
		}
	})}
	provider, err := NewOllama("http://localhost:11434", client)
	if err != nil {
		t.Fatalf("NewOllama() error = %v", err)
	}
	models, err := provider.Models(context.Background())
	if err != nil || len(models) != 2 || models[0] != "gemma3:4b" {
		t.Fatalf("Models() = %#v, %v", models, err)
	}
	data, err := provider.Generate(context.Background(), "qwen3:8b", testRequest())
	if err != nil || !strings.Contains(string(data), SchemaVersion) {
		t.Fatalf("Generate() = %q, %v", data, err)
	}
}

func TestOllamaReportsErrorsAndResponseLimits(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusNotFound, `{"error":"model not found"}`), nil
	})}
	provider, _ := NewOllama("http://localhost:11434", client)
	if _, err := provider.Generate(context.Background(), "missing", testRequest()); err == nil || !strings.Contains(err.Error(), "model not found") {
		t.Fatalf("Generate() error = %v", err)
	}

	oversized := &Ollama{endpoint: "http://localhost", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(strings.Repeat("x", maxOllamaResponseSize+1))), Header: make(http.Header)}, nil
	})}}
	if _, err := oversized.Models(context.Background()); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("Models() error = %v", err)
	}
}

func TestOllamaGenerationHonorsCancellation(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	provider, _ := NewOllama("http://localhost:11434", client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Generate(ctx, "qwen3:8b", testRequest()); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("Generate() cancellation error = %v", err)
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
