package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLiveProviderContract is deliberately opt-in. It uses fictional evidence,
// never prints raw provider output, and validates with the production decoder.
func TestLiveProviderContract(t *testing.T) {
	providerName := strings.ToLower(strings.TrimSpace(os.Getenv("TAILORCV_AI_LIVE_PROVIDER")))
	if providerName == "" {
		t.Skip("set TAILORCV_AI_LIVE_PROVIDER to ollama or gemini")
	}

	var provider Provider
	var err error
	switch providerName {
	case "ollama":
		provider, err = NewOllama(os.Getenv("TAILORCV_OLLAMA_ENDPOINT"), nil)
	case "gemini":
		apiKey := strings.TrimSpace(os.Getenv("TAILORCV_GEMINI_API_KEY"))
		if apiKey == "" {
			t.Fatal("TAILORCV_GEMINI_API_KEY is required for the opt-in Gemini test")
		}
		provider, err = NewGemini(apiKey, nil)
	default:
		t.Fatalf("TAILORCV_AI_LIVE_PROVIDER %q is not supported", providerName)
	}
	if err != nil {
		t.Fatalf("create %s provider: %v", providerName, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	models, err := provider.Models(ctx)
	if err != nil {
		t.Fatalf("discover %s models: %v", providerName, err)
	}
	model := strings.TrimSpace(os.Getenv("TAILORCV_AI_LIVE_MODEL"))
	if model == "" {
		if len(models) != 1 {
			t.Fatalf("set TAILORCV_AI_LIVE_MODEL; provider returned %d models", len(models))
		}
		model = models[0]
	}
	if !includesModel(models, model) {
		t.Fatalf("model %q was not returned by %s discovery", model, providerName)
	}

	request := Request{
		SchemaVersion: SchemaVersion,
		Job: JobRequirements{
			Role:             "Platform Engineer",
			RequiredSkills:   []string{"Go"},
			PreferredSkills:  []string{"release automation"},
			Responsibilities: []string{"Improve reliable software delivery"},
			Keywords:         []string{"deployment", "release", "reliability"},
		},
		Facts: []Fact{{
			ID: "fictional-fact-1", SourceType: "experience", SourceLabel: "Fictional Systems · Engineer",
			Text: "Reduced deployment time by 40% with an audited Go release pipeline", Technologies: []string{"Go"},
		}},
	}
	raw, err := provider.Generate(ctx, model, request)
	if err != nil {
		t.Fatalf("generate with %s model %q: %v", providerName, model, err)
	}
	validation := DecodeAndValidate(request, raw)
	if len(validation.Errors) > 0 {
		t.Fatalf("%s model %q failed the production contract: %s", providerName, model, strings.Join(validation.Errors, "; "))
	}
	if len(validation.Response.Proposals) == 0 {
		t.Fatal("production validation passed without proposals")
	}
}

func includesModel(models []string, wanted string) bool {
	wanted = strings.TrimPrefix(wanted, "models/")
	for _, model := range models {
		if strings.TrimPrefix(model, "models/") == wanted {
			return true
		}
	}
	return false
}
