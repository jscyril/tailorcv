package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/jscyril/tailorcv/internal/domain"
)

var errRateLimit = errors.New("GitHub rate limit reached")

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

// GetPublicRepository fetches one public github.com owner/repository URL.
func (c *Client) GetPublicRepository(ctx context.Context, repositoryURL string) (domain.GitHubRepository, error) {
	owner, repository, err := parsePublicRepositoryURL(repositoryURL)
	if err != nil {
		return domain.GitHubRepository{}, err
	}
	endpoint := c.baseURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.GitHubRepository{}, fmt.Errorf("create GitHub repository request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	request.Header.Set("User-Agent", "TailorCV/0.1")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return domain.GitHubRepository{}, fmt.Errorf("fetch GitHub repository: %w", err)
	}
	item, err := decodeSingleRepositoryResponse(response)
	if err != nil {
		return domain.GitHubRepository{}, err
	}
	return c.hydrateRepository(ctx, owner, item)
}

type repositoryResponse struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Homepage    string   `json:"homepage"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	Fork        bool     `json:"fork"`
	Archived    bool     `json:"archived"`
	Visibility  string   `json:"visibility"`
	UpdatedAt   string   `json:"updated_at"`
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
	languageRequestsAvailable := true
	readmeRequestsAvailable := true
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
			item := repositoryToDomain(repository)
			if !repository.Fork && !repository.Archived && repository.Language != "" && languageRequestsAvailable {
				languages, err := c.listRepositoryLanguages(ctx, username, repository.Name)
				if err == nil {
					item.Languages, item.LanguagesComplete = languages, true
				} else if errors.Is(err, errRateLimit) {
					languageRequestsAvailable = false
				} else if ctx.Err() != nil {
					return nil, ctx.Err()
				}
			}
			if !repository.Fork && !repository.Archived && readmeRequestsAvailable {
				readme, err := c.getRepositoryReadme(ctx, username, repository.Name)
				if err == nil {
					item.Readme, item.ReadmeComplete = readme, true
				} else if errors.Is(err, errRateLimit) {
					readmeRequestsAvailable = false
				} else if ctx.Err() != nil {
					return nil, ctx.Err()
				}
			}
			repositories = append(repositories, item)
		}
		if len(pageRepositories) < 100 {
			return repositories, nil
		}
	}
	return repositories, nil
}

func (c *Client) hydrateRepository(ctx context.Context, owner string, repository repositoryResponse) (domain.GitHubRepository, error) {
	item := repositoryToDomain(repository)
	if item.Fork || item.Archived {
		return item, nil
	}
	if repository.Language != "" {
		languages, err := c.listRepositoryLanguages(ctx, owner, repository.Name)
		if err == nil {
			item.Languages, item.LanguagesComplete = languages, true
		} else if ctx.Err() != nil {
			return domain.GitHubRepository{}, ctx.Err()
		}
	}
	readme, err := c.getRepositoryReadme(ctx, owner, repository.Name)
	if err == nil {
		item.Readme, item.ReadmeComplete = readme, true
	} else if ctx.Err() != nil {
		return domain.GitHubRepository{}, ctx.Err()
	}
	return item, nil
}

func repositoryToDomain(repository repositoryResponse) domain.GitHubRepository {
	return domain.GitHubRepository{ID: repository.ID, Name: repository.Name, Description: repository.Description, HTMLURL: repository.HTMLURL, Homepage: repository.Homepage, Language: repository.Language, Topics: repository.Topics, Fork: repository.Fork, Archived: repository.Archived, Visibility: repository.Visibility, UpdatedAt: repository.UpdatedAt}
}

func parsePublicRepositoryURL(value string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || (parsed.Host != "github.com" && parsed.Host != "www.github.com") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", fmt.Errorf("enter a public GitHub repository URL such as https://github.com/owner/repository")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("enter a public GitHub repository URL such as https://github.com/owner/repository")
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

func (c *Client) getRepositoryReadme(ctx context.Context, owner, repository string) (string, error) {
	endpoint := c.baseURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository) + "/readme"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create GitHub README request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github.raw+json")
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	request.Header.Set("User-Agent", "TailorCV/0.1")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("fetch GitHub repository README: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
		return "", errRateLimit
	}
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return "", fmt.Errorf("GitHub README returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, domain.MaxGitHubReadmeBytes+1))
	if err != nil {
		return "", fmt.Errorf("read GitHub repository README: %w", err)
	}
	if len(data) > domain.MaxGitHubReadmeBytes {
		return "", fmt.Errorf("GitHub repository README exceeded the size limit")
	}
	return strings.TrimSpace(string(data)), nil
}

func (c *Client) listRepositoryLanguages(ctx context.Context, owner, repository string) ([]domain.RepositoryLanguage, error) {
	endpoint := c.baseURL + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repository) + "/languages"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create GitHub languages request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", apiVersion)
	request.Header.Set("User-Agent", "TailorCV/0.1")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch GitHub repository languages: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
		return nil, errRateLimit
	}
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return nil, fmt.Errorf("GitHub languages returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read GitHub languages response: %w", err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("GitHub languages response exceeded the size limit")
	}
	var responseLanguages map[string]int64
	if err := json.Unmarshal(data, &responseLanguages); err != nil {
		return nil, fmt.Errorf("decode GitHub repository languages: %w", err)
	}
	languages := make([]domain.RepositoryLanguage, 0, len(responseLanguages))
	for name, bytes := range responseLanguages {
		languages = append(languages, domain.RepositoryLanguage{Name: name, Bytes: bytes})
	}
	sort.Slice(languages, func(left, right int) bool {
		if languages[left].Bytes == languages[right].Bytes {
			return languages[left].Name < languages[right].Name
		}
		return languages[left].Bytes > languages[right].Bytes
	})
	return languages, nil
}

func decodeRepositoryResponse(response *http.Response) ([]repositoryResponse, error) {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("GitHub user was not found")
		}
		if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("%w; try again later", errRateLimit)
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

func decodeSingleRepositoryResponse(response *http.Response) (repositoryResponse, error) {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		if response.StatusCode == http.StatusNotFound {
			return repositoryResponse{}, fmt.Errorf("GitHub repository was not found or is private")
		}
		if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
			return repositoryResponse{}, fmt.Errorf("%w; try again later", errRateLimit)
		}
		return repositoryResponse{}, fmt.Errorf("GitHub returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return repositoryResponse{}, fmt.Errorf("read GitHub repository response: %w", err)
	}
	if len(data) > maxResponseSize {
		return repositoryResponse{}, fmt.Errorf("GitHub repository response exceeded the size limit")
	}
	var repository repositoryResponse
	if err := json.Unmarshal(data, &repository); err != nil {
		return repositoryResponse{}, fmt.Errorf("decode GitHub repository: %w", err)
	}
	return repository, nil
}
