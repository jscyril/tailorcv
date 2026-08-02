package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestListPublicRepositoriesUsesRequiredHeadersAndPagination(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Accept") != "application/vnd.github+json" || request.Header.Get("X-GitHub-Api-Version") != apiVersion || request.Header.Get("User-Agent") == "" {
			t.Errorf("request headers = %#v", request.Header)
		}
		if request.URL.Query().Get("type") != "owner" || request.URL.Query().Get("per_page") != "100" {
			t.Errorf("query = %s", request.URL.RawQuery)
		}
		page, _ := strconv.Atoi(request.URL.Query().Get("page"))
		count := 100
		if page == 2 {
			count = 1
		}
		result := make([]repositoryResponse, count)
		for index := range result {
			result[index] = repositoryResponse{Name: "repo-" + strconv.Itoa(index), HTMLURL: "https://github.com/example/repo"}
		}
		data, _ := json.Marshal(result)
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})}
	client := NewClient(httpClient)
	client.baseURL = "https://api.github.test"

	repositories, err := client.ListPublicRepositories(context.Background(), "example")
	if err != nil {
		t.Fatalf("ListPublicRepositories() error = %v", err)
	}
	if len(repositories) != 101 || requests != 2 {
		t.Fatalf("repositories = %d, requests = %d", len(repositories), requests)
	}
}

func TestListPublicRepositoriesReportsNotFoundAndRateLimit(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusForbidden} {
		httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(bytes.NewReader(nil)), Header: make(http.Header)}, nil
		})}
		client := NewClient(httpClient)
		client.baseURL = "https://api.github.test"
		if _, err := client.ListPublicRepositories(context.Background(), "missing"); err == nil {
			t.Fatalf("status %d expected error", status)
		}
	}
}

func TestListPublicRepositoriesFetchesCompleteLanguages(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var data []byte
		if request.URL.Path == "/users/example/repos" {
			data, _ = json.Marshal([]repositoryResponse{{Name: "polyglot", HTMLURL: "https://github.com/example/polyglot", Language: "Go"}})
		} else if request.URL.Path == "/repos/example/polyglot/languages" {
			data = []byte(`{"TypeScript":120,"Go":900,"Shell":45}`)
		} else {
			t.Fatalf("unexpected request path %q", request.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})}
	client := NewClient(httpClient)
	client.baseURL = "https://api.github.test"

	repositories, err := client.ListPublicRepositories(context.Background(), "example")
	if err != nil {
		t.Fatalf("ListPublicRepositories() error = %v", err)
	}
	if len(repositories) != 1 || !repositories[0].LanguagesComplete || len(repositories[0].Languages) != 3 {
		t.Fatalf("repository languages = %#v", repositories)
	}
	if repositories[0].Languages[0].Name != "Go" || repositories[0].Languages[1].Name != "TypeScript" {
		t.Fatalf("languages are not sorted by byte count: %#v", repositories[0].Languages)
	}
}

func TestListPublicRepositoriesFallsBackAfterLanguageRateLimit(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.URL.Path == "/users/example/repos" {
			data, _ := json.Marshal([]repositoryResponse{{Name: "one", Language: "Go"}, {Name: "two", Language: "Rust"}})
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: http.StatusForbidden, Status: "403 Forbidden", Body: io.NopCloser(bytes.NewReader(nil)), Header: make(http.Header)}, nil
	})}
	client := NewClient(httpClient)
	client.baseURL = "https://api.github.test"

	repositories, err := client.ListPublicRepositories(context.Background(), "example")
	if err != nil {
		t.Fatalf("ListPublicRepositories() error = %v", err)
	}
	if len(repositories) != 2 || repositories[0].LanguagesComplete || repositories[1].LanguagesComplete || requests != 2 {
		t.Fatalf("fallback repositories = %#v, requests = %d", repositories, requests)
	}
}
