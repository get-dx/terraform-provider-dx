package dxapi

import (
	"net/http"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	version    string
}

func NewClient(baseURL, token, version string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: http.DefaultClient,
		version:    version,
	}
}
