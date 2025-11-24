package dxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// API model structs for unmarshalling API responses

type APIEntityType struct {
	Identifier  string          `json:"identifier"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Ordering    int64           `json:"ordering"`
	Properties  []*APIProperty  `json:"properties"`
	Aliases     map[string]bool `json:"aliases"`
}

type APIProperty struct {
	Identifier  string                 `json:"identifier"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description *string                `json:"description"`
	Visibility  *string                `json:"visibility"`
	Ordering    *int64                 `json:"ordering"`
	Definition  *APIPropertyDefinition `json:"definition"`
}

type APIPropertyDefinition struct {
	Options          []APIPropertyOption `json:"options"`
	SQL              *string             `json:"sql,omitempty"`
	CallToAction     *string             `json:"call_to_action,omitempty"`
	CallToActionType *string             `json:"call_to_action_type,omitempty"`
}

type APIPropertyOption struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

// APIEntityTypeResponse is the top-level response from the DX API for entity type endpoints.
//
// Example:
//
// ```json
// { "ok": true, "entity_type": { ... } }
// ```.
type APIEntityTypeResponse struct {
	Ok         bool          `json:"ok"`
	EntityType APIEntityType `json:"entity_type"`
}

// APIEntityTypesListResponse is the response from the entityTypes.list endpoint.
type APIEntityTypesListResponse struct {
	Ok               bool            `json:"ok"`
	EntityTypes      []APIEntityType `json:"entity_types"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

func (c *Client) CreateEntityType(ctx context.Context, payload map[string]interface{}) (*APIEntityTypeResponse, error) {
	tflog.Info(ctx, "Calling CreateEntityType")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/entityTypes.create", c.baseURL)
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

	// Decode the response into the APIEntityTypeResponse struct
	var apiResp APIEntityTypeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from CreateEntityType:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	// Return the APIEntityTypeResponse object
	return &apiResp, nil
}

func (c *Client) GetEntityType(ctx context.Context, identifier string) (*APIEntityTypeResponse, error) {
	url := fmt.Sprintf("%s/entityTypes.info?identifier=%s", c.baseURL, identifier)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var apiResp APIEntityTypeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from GetEntityType:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) UpdateEntityType(ctx context.Context, payload map[string]interface{}) (*APIEntityTypeResponse, error) {
	tflog.Info(ctx, "Calling UpdateEntityType")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/entityTypes.update", c.baseURL)
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

	var apiResp APIEntityTypeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from UpdateEntityType:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) DeleteEntityType(ctx context.Context, identifier string) (bool, error) {
	tflog.Info(ctx, "Calling DeleteEntityType")
	tflog.Info(ctx, fmt.Sprintf("Deleting entity type with identifier: %s", identifier))

	payload := map[string]interface{}{"identifier": identifier}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/entityTypes.delete", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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

func (c *Client) ListEntityTypes(ctx context.Context) (*APIEntityTypesListResponse, error) {
	tflog.Info(ctx, "Calling ListEntityTypes")

	url := fmt.Sprintf("%s/entityTypes.list", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var apiResp APIEntityTypesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from ListEntityTypes:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}
