// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dxapi

import (
	"net/http"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: http.DefaultClient,
	}
}
