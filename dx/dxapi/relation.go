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

type RelationNotFoundError struct {
	Identifier string
}

func (e *RelationNotFoundError) Error() string {
	return fmt.Sprintf("relation not found: %s", e.Identifier)
}

type APIRelation struct {
	Identifier                 string  `json:"identifier"`
	Type                       string  `json:"type"`
	InverseType                string  `json:"inverse_type"`
	Cardinality                string  `json:"cardinality"`
	Description                *string `json:"description"`
	SourceEntityTypeIdentifier string  `json:"source_entity_type_identifier"`
	TargetEntityTypeIdentifier string  `json:"target_entity_type_identifier"`
	CreatedAt                  string  `json:"created_at"`
	UpdatedAt                  string  `json:"updated_at"`
}

type APIRelationResponse struct {
	Ok       bool        `json:"ok"`
	Relation APIRelation `json:"relation"`
}

func (c *Client) CreateRelation(ctx context.Context, payload map[string]interface{}) (*APIRelationResponse, error) {
	tflog.Info(ctx, "Calling CreateRelation")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/catalog.relations.create", c.baseURL)
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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIRelationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from CreateRelation:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) GetRelation(ctx context.Context, identifier string) (*APIRelationResponse, error) {
	reqURL := fmt.Sprintf("%s/catalog.relations.info?identifier=%s", c.baseURL, url.QueryEscape(identifier))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &RelationNotFoundError{Identifier: identifier}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIRelationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from GetRelation:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) UpdateRelation(ctx context.Context, payload map[string]interface{}) (*APIRelationResponse, error) {
	tflog.Info(ctx, "Calling UpdateRelation")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/catalog.relations.update", c.baseURL)
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

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Error response body:\n%s", prettyJSON.String()))
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Error response body (raw):\n%s", string(body)))
		}

		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIRelationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from UpdateRelation:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) DeleteRelation(ctx context.Context, identifier string) (bool, error) {
	tflog.Info(ctx, "Calling DeleteRelation")
	tflog.Info(ctx, fmt.Sprintf("Deleting relation with identifier: %s", identifier))

	payload := map[string]interface{}{"identifier": identifier}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/catalog.relations.delete", c.baseURL)
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

	if resp.StatusCode == http.StatusNotFound {
		return true, nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}
