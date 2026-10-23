// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	user       string
	token      string
	httpClient *http.Client
}

// NewClient keeps cookies because Jenkins binds the CSRF crumb to the session when authenticating with a password
func NewClient(baseURL, user, token string, transport http.RoundTripper) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		user:       user,
		token:      token,
		httpClient: &http.Client{Transport: transport, Jar: jar, Timeout: 30 * time.Second},
	}
}

// resolve keeps only the path of absolute URLs from Jenkins, they use its configured root URL which may differ from baseURL
func (c *Client) resolve(pathOrURL string) string {
	if u, err := url.Parse(pathOrURL); err == nil && u.IsAbs() {
		pathOrURL = u.RequestURI()
	}
	return c.baseURL + "/" + strings.TrimPrefix(pathOrURL, "/")
}

func (c *Client) newRequest(ctx context.Context, method, pathOrURL string, body io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolve(pathOrURL), body)
	if err != nil {
		return nil, err
	}
	if c.user != "" {
		req.SetBasicAuth(c.user, c.token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func (c *Client) getJSON(ctx context.Context, pathOrURL string, v any) (int, error) {
	req, err := c.newRequest(ctx, http.MethodGet, pathOrURL, nil, "")
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) post(ctx context.Context, pathOrURL string, body io.Reader, contentType string) (*http.Response, error) {
	var crumb struct {
		CrumbRequestField string `json:"crumbRequestField"`
		Crumb             string `json:"crumb"`
	}
	status, err := c.getJSON(ctx, "crumbIssuer/api/json", &crumb)
	if err != nil {
		return nil, fmt.Errorf("post: fetch crumb: %w", err)
	}
	if status != http.StatusOK && status != http.StatusNotFound { // 404: CSRF protection disabled
		return nil, fmt.Errorf("post: fetch crumb: unexpected status %d", status)
	}
	req, err := c.newRequest(ctx, http.MethodPost, pathOrURL, body, contentType)
	if err != nil {
		return nil, err
	}
	if crumb.CrumbRequestField != "" {
		req.Header.Set(crumb.CrumbRequestField, crumb.Crumb)
	}
	return c.httpClient.Do(req)
}

func (c *Client) postExpect(ctx context.Context, pathOrURL string, body io.Reader, contentType string, okStatus ...int) (*http.Response, error) {
	resp, err := c.post(ctx, pathOrURL, body, contentType)
	if err != nil {
		return nil, err
	}
	for _, s := range okStatus {
		if resp.StatusCode == s {
			return resp, nil
		}
	}
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	resp.Body.Close()
	return nil, fmt.Errorf("postExpect: %s returned %d: %s", pathOrURL, resp.StatusCode, strings.TrimSpace(string(detail)))
}

func jobPath(name string) string {
	return "job/" + url.PathEscape(name)
}

// CreateOrUpdateJob creates the job, or replaces the configuration of an existing one
func (c *Client) CreateOrUpdateJob(ctx context.Context, name, configXML string) error {
	var job struct{}
	status, err := c.getJSON(ctx, jobPath(name)+"/api/json?tree=name", &job)
	if err != nil {
		return fmt.Errorf("CreateOrUpdateJob: %w", err)
	}
	target := "createItem?name=" + url.QueryEscape(name)
	if status == http.StatusOK {
		target = jobPath(name) + "/config.xml"
	}
	resp, err := c.postExpect(ctx, target, strings.NewReader(configXML), "application/xml", http.StatusOK)
	if err != nil {
		return fmt.Errorf("CreateOrUpdateJob: %w", err)
	}
	return resp.Body.Close()
}

// TriggerBuild queues a build and returns the URL of its queue item
func (c *Client) TriggerBuild(ctx context.Context, name string, params url.Values) (string, error) {
	resp, err := c.postExpect(ctx, jobPath(name)+"/buildWithParameters", strings.NewReader(params.Encode()), "application/x-www-form-urlencoded", http.StatusCreated)
	if err != nil {
		return "", fmt.Errorf("TriggerBuild: %w", err)
	}
	resp.Body.Close()
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("TriggerBuild: Jenkins did not return a queue item location")
	}
	return location, nil
}

type QueueItem struct {
	Cancelled  bool `json:"cancelled"`
	Executable *struct {
		URL string `json:"url"`
	} `json:"executable"`
}

// QueueItem returns the build URL once the queued build has started, or "" while it is still waiting
func (c *Client) QueueItem(ctx context.Context, queueURL string) (buildURL string, cancelled bool, _ error) {
	var item QueueItem
	status, err := c.getJSON(ctx, strings.TrimSuffix(queueURL, "/")+"/api/json", &item)
	if err != nil {
		return "", false, fmt.Errorf("QueueItem: %w", err)
	}
	if status != http.StatusOK {
		return "", false, fmt.Errorf("QueueItem: unexpected status %d", status)
	}
	if item.Executable != nil {
		buildURL = item.Executable.URL
	}
	return buildURL, item.Cancelled, nil
}

type Build struct {
	Building bool   `json:"building"`
	Result   string `json:"result"` // SUCCESS, FAILURE, UNSTABLE, ABORTED or empty while building
	Number   int    `json:"number"`
}

func (c *Client) Build(ctx context.Context, buildURL string) (*Build, error) {
	var build Build
	status, err := c.getJSON(ctx, strings.TrimSuffix(buildURL, "/")+"/api/json?tree=building,result,number", &build)
	if err != nil {
		return nil, fmt.Errorf("Build: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Build: unexpected status %d", status)
	}
	return &build, nil
}
