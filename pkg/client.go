package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL        *url.URL
	defaultHeaders http.Header
	client         *http.Client
}

func NewClient(
	baseURL string,
	defaultHeaders http.Header,
	client *http.Client,
) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("empty base URL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid base URL")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	if client == nil {
		client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	if defaultHeaders == nil {
		defaultHeaders = make(http.Header)
	}

	return &Client{
		baseURL:        u,
		defaultHeaders: defaultHeaders.Clone(),
		client:         client,
	}, nil
}

func (c *Client) NewRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
) (*http.Request, error) {
	u, err := url.JoinPath(c.baseURL.String(), path)
	if err != nil {
		return nil, fmt.Errorf("join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer resp.Body.Close()

		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

		return nil, fmt.Errorf(
			"unexpected response: status=%d body=%q",
			resp.StatusCode,
			string(errBody),
		)
	}

	return resp, nil
}

func (c *Client) DoJSON(
	ctx context.Context,
	method string,
	path string,
	body any,
	target any,
) error {
	var bodyReader io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}

		bodyReader = bytes.NewReader(b)
	}

	req, err := c.NewRequest(ctx, method, path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if target == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func (c *Client) Get(
	ctx context.Context,
	path string,
	target any,
) error {
	return c.DoJSON(ctx, http.MethodGet, path, nil, target)
}

func (c *Client) Post(
	ctx context.Context,
	path string,
	body any,
	target any,
) error {
	return c.DoJSON(ctx, http.MethodPost, path, body, target)
}

func (c *Client) Put(
	ctx context.Context,
	path string,
	body any,
	target any,
) error {
	return c.DoJSON(ctx, http.MethodPut, path, body, target)
}

func (c *Client) Patch(
	ctx context.Context,
	path string,
	body any,
	target any,
) error {
	return c.DoJSON(ctx, http.MethodPatch, path, body, target)
}

func (c *Client) Delete(
	ctx context.Context,
	path string,
	target any,
) error {
	return c.DoJSON(ctx, http.MethodDelete, path, nil, target)
}
