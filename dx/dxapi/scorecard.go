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

type APIScorecard struct {
	// Required fields
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	EntityFilterType    string `json:"entity_filter_type"`
	EvaluationFrequency int32  `json:"evaluation_frequency_hours"`

	// Conditionally required fields for levels based scorecards
	EmptyLevelLabel *string     `json:"empty_level_label"`
	EmptyLevelColor *string     `json:"empty_level_color"`
	Levels          []*APILevel `json:"levels"`

	// Conditionally required fields for points based scorecards
	CheckGroups []*APICheckGroup `json:"check_groups"`

	// Optional fields
	Description                 *string     `json:"description"`
	Published                   bool        `json:"published"`
	EntityFilterTypeIdentifiers []*string   `json:"entity_filter_type_identifiers"`
	EntityFilterSql             *string     `json:"entity_filter_sql"`
	Checks                      []*APICheck `json:"checks"`
}

type APILevel struct {
	Key   *string `json:"key"`
	Id    *string `json:"id"`
	Name  *string `json:"name"`
	Color *string `json:"color"`
	Rank  *int32  `json:"rank"`
}

type APICheckGroup struct {
	Key      *string `json:"key"`
	Id       *string `json:"id"`
	Name     *string `json:"name"`
	Ordering *int32  `json:"ordering"`
}

type APICheck struct {
	Id                     *string                 `json:"id"`
	Name                   *string                 `json:"name"`
	Description            *string                 `json:"description"`
	Ordering               int32                   `json:"ordering"`
	Sql                    *string                 `json:"sql"`
	FilterSql              *string                 `json:"filter_sql"`
	FilterMessage          *string                 `json:"filter_message"`
	OutputEnabled          bool                    `json:"output_enabled"`
	OutputType             *string                 `json:"output_type"`
	OutputAggregation      *string                 `json:"output_aggregation"`
	OutputCustomOptions    *APIOutputCustomOptions `json:"output_custom_options"`
	EstimatedDevDays       *float32                `json:"estimated_dev_days"`
	ExternalUrl            *string                 `json:"external_url"`
	Published              bool                    `json:"published"`
	ScorecardLevelKey      *string                 `json:"scorecard_level_key"`
	Level                  *APILevel               `json:"level"`
	ScorecardCheckGroupKey *string                 `json:"scorecard_check_group_key"`
	CheckGroup             *APICheckGroup          `json:"check_group"`
	Points                 *int32                  `json:"points"`
}

type APIOutputCustomOptions struct {
	Unit     string `json:"unit"`
	Decimals *int32 `json:"decimals"` // TODO: "auto" or number
}

// APIResponse is the top-level response from the DX API for scorecard endpoints.
//
// Example:
//
// ```json
// { "ok": true, "scorecard": { ... } }
// ```
type APIResponse struct {
	Ok        bool         `json:"ok"`
	Scorecard APIScorecard `json:"scorecard"`
}

func (c *Client) CreateScorecard(ctx context.Context, payload map[string]interface{}) (*APIResponse, error) {
	tflog.Info(ctx, "Calling CreateScorecard")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/scorecards.create", c.baseURL)
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

	// Decode the response into the APIResponse struct
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from CreateScorecard:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	// Return the APIResponse object
	return &apiResp, nil
}

func (c *Client) GetScorecard(ctx context.Context, id string) (*APIResponse, error) {
	url := fmt.Sprintf("%s/scorecards.info?id=%s", c.baseURL, id)
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

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Debug(ctx, "API Response from CreateScorecard", map[string]interface{}{
			"response": string(respJson),
		})
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) UpdateScorecard(ctx context.Context, payload map[string]interface{}) (*APIResponse, error) {
	tflog.Info(ctx, "Calling UpdateScorecard")

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/scorecards.update", c.baseURL)
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

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	// Log the API response for debugging
	if respJson, err := json.MarshalIndent(apiResp, "", "  "); err == nil {
		tflog.Info(ctx, fmt.Sprintf("API Response from UpdateScorecard:\n%s", string(respJson)))
	} else {
		tflog.Debug(ctx, "Could not marshal API response", map[string]interface{}{
			"error": err,
		})
	}

	return &apiResp, nil
}

func (c *Client) DeleteScorecard(ctx context.Context, id string) (bool, error) {
	payload := map[string]interface{}{"id": id}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf("%s/scorecards.delete", c.baseURL)
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

func setRequestHeaders(req *http.Request, client *Client) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.token)
	req.Header.Set("X-Client-Type", "terraform-provider-dx")
	req.Header.Set("X-Client-Version", client.version)
}
