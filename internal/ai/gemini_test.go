package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGeminiModelsAndGenerateContract(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("x-goog-api-key") != "test-secret" {
			t.Fatalf("API key header was not set")
		}
		switch {
		case request.URL.Path == "/v1beta/models":
			return jsonResponse(http.StatusOK, `{"models":[{"name":"models/gemini-z","supportedGenerationMethods":["generateContent"]},{"name":"models/embed-only","supportedGenerationMethods":["embedContent"]},{"name":"models/gemini-a","supportedGenerationMethods":["generateContent"]}]}`), nil
		case strings.HasSuffix(request.URL.Path, ":generateContent"):
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			config := body["generationConfig"].(map[string]any)
			if config["responseMimeType"] != "application/json" || config["responseJsonSchema"] == nil || config["temperature"] != float64(0) {
				t.Fatalf("generation config = %#v", config)
			}
			return jsonResponse(http.StatusOK, `{"candidates":[{"content":{"parts":[{"text":"{\"schemaVersion\":\"tailorcv.ai.tailoring.v1\",\"proposals\":[]}"}]}}]}`), nil
		default:
			t.Fatalf("unexpected request URL %q", request.URL.String())
			return nil, nil
		}
	})}
	provider, err := newGemini("test-secret", "https://example.test/v1beta", client)
	if err != nil {
		t.Fatalf("newGemini() error = %v", err)
	}
	models, err := provider.Models(context.Background())
	if err != nil || len(models) != 2 || models[0] != "gemini-a" || models[1] != "gemini-z" {
		t.Fatalf("Models() = %#v, %v", models, err)
	}
	data, err := provider.Generate(context.Background(), "models/gemini-a", testRequest())
	if err != nil || !strings.Contains(string(data), SchemaVersion) {
		t.Fatalf("Generate() = %q, %v", data, err)
	}
}

func TestGeminiRedactsCredentialAndHonorsLimits(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"message":"API key super-secret rejected"}}`), nil
	})}
	provider, _ := newGemini("super-secret", "https://example.test/v1beta", client)
	_, err := provider.Models(context.Background())
	if err == nil || strings.Contains(err.Error(), "super-secret") || !strings.Contains(err.Error(), "API key [redacted] rejected") {
		t.Fatalf("Models() error = %v", err)
	}

	provider.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(strings.Repeat("x", maxOllamaResponseSize+1))), Header: make(http.Header)}, nil
	})}
	if _, err := provider.Models(context.Background()); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("oversized Models() error = %v", err)
	}
}

func TestGeminiGenerationHonorsCancellation(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	provider, _ := newGemini("test-secret", "https://example.test/v1beta", client)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Generate(ctx, "gemini-test", testRequest()); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("Generate() cancellation error = %v", err)
	}
}
