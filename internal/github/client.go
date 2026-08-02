package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

const (
	defaultBaseURL  = "https://api.github.com"
	apiVersion      = "2026-03-10"
	maxResponseSize = 4 << 20
	maxPages        = 10
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type repositoryResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Homepage    string   `json:"homepage"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	Fork        bool     `json:"fork"`
	Archived    bool     `json:"archived"`
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: defaultBaseURL, httpClient: httpClient}
}

func (c *Client) ListPublicRepositories(ctx context.Context, username string) ([]domain.GitHubRepository, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("GitHub username is required")
	}
	repositories := make([]domain.GitHubRepository, 0)
	for page := 1; page <= maxPages; page++ {
		endpoint, err := url.Parse(c.baseURL + "/users/" + url.PathEscape(username) + "/repos")
		if err != nil {
			return nil, fmt.Errorf("build GitHub request: %w", err)
		}
		query := endpoint.Query()
		query.Set("type", "owner")
		query.Set("sort", "updated")
		query.Set("direction", "desc")
		query.Set("per_page", "100")
		query.Set("page", fmt.Sprintf("%d", page))
		endpoint.RawQuery = query.Encode()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create GitHub request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("X-GitHub-Api-Version", apiVersion)
		request.Header.Set("User-Agent", "TailorCV/0.1")
		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("fetch GitHub repositories: %w", err)
		}
		pageRepositories, err := decodeRepositoryResponse(response)
		if err != nil {
			return nil, err
		}
		for _, repository := range pageRepositories {
			repositories = append(repositories, domain.GitHubRepository{
				Name: repository.Name, Description: repository.Description, HTMLURL: repository.HTMLURL,
				Homepage: repository.Homepage, Language: repository.Language, Topics: repository.Topics,
				Fork: repository.Fork, Archived: repository.Archived,
			})
		}
		if len(pageRepositories) < 100 {
			return repositories, nil
		}
	}
	return repositories, nil
}

func decodeRepositoryResponse(response *http.Response) ([]repositoryResponse, error) {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("GitHub user was not found")
		}
		if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("GitHub rate limit reached; try again later")
		}
		return nil, fmt.Errorf("GitHub returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	limited := io.LimitReader(response.Body, maxResponseSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read GitHub response: %w", err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("GitHub response exceeded the size limit")
	}
	var repositories []repositoryResponse
	if err := json.Unmarshal(data, &repositories); err != nil {
		return nil, fmt.Errorf("decode GitHub repositories: %w", err)
	}
	return repositories, nil
}
