package turf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Client is an HTTP client wrapper for the JRA website.
// It transparently decodes responses from Shift-JIS to UTF-8,
// and resolves relative request paths against a configurable base URL.
type Client struct {
	client *http.Client

	userAgent string
	baseURL   *url.URL
}

// NewClient creates a new Client for making requests to the JRA website.
// If httpClient is nil, a default http.Client is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	c := &Client{
		client:    httpClient,
		userAgent: "turf/" + Version,
	}
	c.baseURL, _ = url.Parse("https://www.jra.go.jp/")
	return c
}

// SetBaseURL overrides the base URL used to resolve request paths.
func (c *Client) SetBaseURL(u *url.URL) {
	c.baseURL = u
}

// NewRequest resolves urlStr against baseURL and returns a new http.Request with the specified method and body.
func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	if !strings.HasSuffix(c.baseURL.Path, "/") {
		return nil, fmt.Errorf("baseURL must have a trailing slash, but %q does not", c.baseURL)
	}

	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// NewFormRequest resolves urlStr against baseURL and returns a POST request
// with Content-Type set to application/x-www-form-urlencoded.
func (c *Client) NewFormRequest(urlStr string, form *url.Values) (*http.Request, error) {
	req, err := c.NewRequest(http.MethodPost, urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// Do executes the request and returns the response body decoded from Shift-JIS to UTF-8.
func (c *Client) Do(ctx context.Context, req *http.Request) (io.Reader, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, resp.Body)

	tr := transform.NewReader(buf, japanese.ShiftJIS.NewDecoder())
	return tr, nil
}
