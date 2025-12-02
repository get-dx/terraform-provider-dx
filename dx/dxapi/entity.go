package dxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// API model structs for unmarshalling API responses

type APIEntity struct {
	Id          string                 `json:"id,omitempty"`
	Identifier  string                 `json:"identifier"`
	Name        *string                `json:"name,omitempty"`
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
	OwnerTeams  []APIOwnerTeam         `json:"owner_teams,omitempty"`
	OwnerUsers  []APIOwnerUser         `json:"owner_users,omitempty"`
	Domain      *APIDomain             `json:"domain,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Aliases     map[string][]APIAlias  `json:"aliases,omitempty"`
	Relations   map[string][]string    `json:"relations,omitempty"`
}

type APIOwnerTeam struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type APIOwnerUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type APIDomain struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type APIAlias struct {
	Identifier string `json:"identifier"`
}

// APIEntityResponse is the top-level response from the DX API for entity endpoints.
//
// Example:
//
// ```json
// { "ok": true, "entity": { ... } }
// ```.
type APIEntityResponse struct {
	Ok     bool      `json:"ok"`
	Entity APIEntity `json:"entity"`
}

func (c *Client) CreateEntity(ctx context.Context, payload map[string]interface{}) (*APIEntityResponse, error) {
	tflog.Info(ctx, "Calling CreateEntity")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/entities.create", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	tflog.Info(ctx, fmt.Sprintf("Request body:\n%s", string(body)))

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	// Decode the response into the APIEntityResponse struct
	var apiResp APIEntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from CreateEntity:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	// Return the APIEntityResponse object
	return &apiResp, nil
}

func (c *Client) GetEntity(ctx context.Context, identifier string) (*APIEntityResponse, error) {
	// Note: The API documentation doesn't show an entities.info endpoint,
	// but we'll try entities.info similar to entityTypes.info
	// If it doesn't exist, we may need to use a different approach
	urlStr := fmt.Sprintf("%s/entities.info?identifier=%s", c.baseURL, url.QueryEscape(identifier))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIEntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from GetEntity:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) UpdateEntity(ctx context.Context, payload map[string]interface{}) (*APIEntityResponse, error) {
	tflog.Info(ctx, "Calling UpdateEntity")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/entities.update", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	tflog.Info(ctx, fmt.Sprintf("Request body:\n%s", string(body)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		// format the JSON nicely
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Error response body:\n%s", prettyJSON.String()))
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Error response body (raw):\n%s", string(body)))
		}

		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIEntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from UpdateEntity:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) DeleteEntity(ctx context.Context, identifier string) (bool, error) {
	tflog.Info(ctx, "Calling DeleteEntity")
	tflog.Info(ctx, fmt.Sprintf("Deleting entity with identifier: %s", identifier))

	urlStr := fmt.Sprintf("%s/entities.delete?identifier=%s", c.baseURL, url.QueryEscape(identifier))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}
