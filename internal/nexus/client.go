package nexus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

type Repository struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Type   string `json:"type"`
}

type Component struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Format     string    `json:"format"`
	Group      string    `json:"group"`
	Name       string    `json:"name"`
	Version    string    `json:"version"`
	Assets     []Asset   `json:"assets"`
}

type Asset struct {
	DownloadURL  string    `json:"downloadUrl"`
	Path         string    `json:"path"`
	ID           string    `json:"id"`
	Repository   string    `json:"repository"`
	Format       string    `json:"format"`
	LastModified time.Time `json:"lastModified"`
}

type ComponentPage struct {
	Items             []Component `json:"items"`
	ContinuationToken string      `json:"continuationToken"`
}

type RepositoryPage struct {
	Items             []Repository `json:"items"`
	ContinuationToken string       `json:"continuationToken"`
}

func NewClient(baseURL, username, password string, timeout int) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) GetDockerRepositories() ([]Repository, error) {
	var allRepos []Repository
	continuationToken := ""

	for {
		path := "/service/rest/v1/repositories"
		if continuationToken != "" {
			path += "?continuationToken=" + continuationToken
		}

		body, err := c.doRequest("GET", path)
		if err != nil {
			return nil, err
		}

		var repos []Repository
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("failed to parse repositories: %w", err)
		}

		// Filter for docker hosted repositories
		for _, repo := range repos {
			if repo.Format == "docker" && repo.Type == "hosted" {
				allRepos = append(allRepos, repo)
			}
		}

		// Check if there are more pages (this endpoint doesn't use continuation token)
		break
	}

	return allRepos, nil
}

func (c *Client) GetComponents(repository string) ([]Component, error) {
	var allComponents []Component
	continuationToken := ""

	for {
		path := fmt.Sprintf("/service/rest/v1/components?repository=%s", repository)
		if continuationToken != "" {
			path += "&continuationToken=" + continuationToken
		}

		body, err := c.doRequest("GET", path)
		if err != nil {
			return nil, err
		}

		var page ComponentPage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("failed to parse components: %w", err)
		}

		allComponents = append(allComponents, page.Items...)

		if page.ContinuationToken == "" {
			break
		}
		continuationToken = page.ContinuationToken
	}

	return allComponents, nil
}

func (c *Client) DeleteComponent(componentID string) error {
	path := fmt.Sprintf("/service/rest/v1/components/%s", componentID)
	_, err := c.doRequest("DELETE", path)
	return err
}
