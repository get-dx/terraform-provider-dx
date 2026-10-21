package dxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// relationEdgeListPageSize is the maximum page size accepted by catalog.relationEdges.list.
const relationEdgeListPageSize = 50

// Error codes returned by the catalog.relationEdges.* methods that indicate the
// edge (or something it depends on) no longer exists.
const (
	errCodeEdgeNotFound     = "edge_not_found"
	errCodeEntityNotFound   = "entity_not_found"
	errCodeRelationNotFound = "relation_not_found"
)

// RelationEdgeAPIError is returned when a catalog.relationEdges.* method responds
// with `ok: false` or a non-2xx status code.
type RelationEdgeAPIError struct {
	StatusCode int
	Code       string
	Body       string
}

func (e *RelationEdgeAPIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %q (status code %d), response body: %s", e.Code, e.StatusCode, e.Body)
	}
	return fmt.Sprintf("unexpected status code: %d, response body: %s", e.StatusCode, e.Body)
}

// IsNotFound reports whether the error means the edge, one of its entities, or
// its relation definition does not exist.
func (e *RelationEdgeAPIError) IsNotFound() bool {
	switch e.Code {
	case errCodeEdgeNotFound, errCodeEntityNotFound, errCodeRelationNotFound:
		return true
	}
	return false
}

type APIRelationEdgeEntity struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type APIRelationEdgeListResponse struct {
	Ok               bool                    `json:"ok"`
	Error            string                  `json:"error"`
	EntityRelations  []APIRelationEdgeEntity `json:"entity_relations"`
	ResponseMetadata struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"response_metadata"`
}

type apiOkResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func edgePayload(relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier string) map[string]interface{} {
	return map[string]interface{}{
		"relation_identifier":      relationIdentifier,
		"source_entity_identifier": sourceEntityIdentifier,
		"target_entity_identifier": targetEntityIdentifier,
	}
}

// AddRelationEdge creates an edge between two entities. The API is idempotent,
// so adding an edge that already exists succeeds.
func (c *Client) AddRelationEdge(ctx context.Context, relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier string) error {
	tflog.Info(ctx, "Calling AddRelationEdge")
	return c.postRelationEdge(ctx, "catalog.relationEdges.add", edgePayload(relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier))
}

// RemoveRelationEdge deletes an edge between two entities. An edge that is
// already gone (including because an entity or the relation was deleted) is
// treated as successfully removed.
func (c *Client) RemoveRelationEdge(ctx context.Context, relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier string) error {
	tflog.Info(ctx, "Calling RemoveRelationEdge")
	err := c.postRelationEdge(ctx, "catalog.relationEdges.remove", edgePayload(relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier))
	var apiErr *RelationEdgeAPIError
	if errors.As(err, &apiErr) && apiErr.IsNotFound() {
		tflog.Info(ctx, fmt.Sprintf("Relation edge already absent (%s)", apiErr.Code))
		return nil
	}
	return err
}

// RelationEdgeExists reports whether an edge exists by paging through the
// source entity's outgoing edges for the relation. It returns false (and no
// error) if the source entity or the relation does not exist.
func (c *Client) RelationEdgeExists(ctx context.Context, relationIdentifier, sourceEntityIdentifier, targetEntityIdentifier string) (bool, error) {
	cursor := ""
	for {
		page, err := c.listRelationEdges(ctx, relationIdentifier, sourceEntityIdentifier, "outgoing", cursor)
		if err != nil {
			var apiErr *RelationEdgeAPIError
			if errors.As(err, &apiErr) && apiErr.IsNotFound() {
				tflog.Info(ctx, fmt.Sprintf("Relation edge lookup returned %s", apiErr.Code))
				return false, nil
			}
			return false, err
		}

		for _, entity := range page.EntityRelations {
			if entity.Identifier == targetEntityIdentifier {
				return true, nil
			}
		}

		if page.ResponseMetadata.NextCursor == nil || *page.ResponseMetadata.NextCursor == "" {
			return false, nil
		}
		cursor = *page.ResponseMetadata.NextCursor
	}
}

func (c *Client) listRelationEdges(ctx context.Context, relationIdentifier, entityIdentifier, direction, cursor string) (*APIRelationEdgeListResponse, error) {
	query := url.Values{}
	query.Set("entity_identifier", entityIdentifier)
	query.Set("relation_identifier", relationIdentifier)
	query.Set("direction", direction)
	query.Set("limit", strconv.Itoa(relationEdgeListPageSize))
	if cursor != "" {
		query.Set("cursor", cursor)
	}

	reqURL := fmt.Sprintf("%s/catalog.relationEdges.list?%s", c.baseURL, query.Encode())
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var apiResp APIRelationEdgeListResponse
	decodeErr := json.Unmarshal(body, &apiResp)
	if resp.StatusCode != http.StatusOK || (decodeErr == nil && !apiResp.Ok) {
		return nil, &RelationEdgeAPIError{StatusCode: resp.StatusCode, Code: apiResp.Error, Body: string(body)}
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("decoding API response: %w", decodeErr)
	}

	tflog.Info(ctx, fmt.Sprintf("API Response from ListRelationEdges:\n%s", string(body)))

	return &apiResp, nil
}

func (c *Client) postRelationEdge(ctx context.Context, method string, payload map[string]interface{}) error {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	setRequestHeaders(req, c)

	tflog.Info(ctx, fmt.Sprintf("Request body:\n%s", string(body)))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	var apiResp apiOkResponse
	decodeErr := json.Unmarshal(respBody, &apiResp)
	if (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated) || (decodeErr == nil && !apiResp.Ok) {
		return &RelationEdgeAPIError{StatusCode: resp.StatusCode, Code: apiResp.Error, Body: string(respBody)}
	}

	return nil
}
