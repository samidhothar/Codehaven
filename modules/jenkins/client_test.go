// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package jenkins

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	jobs := map[string]string{}
	var triggered url.Values
	mux := http.NewServeMux()
	requireAuthAndCrumb := func(w http.ResponseWriter, r *http.Request) bool {
		user, pass, _ := r.BasicAuth()
		cookie, _ := r.Cookie("JSESSIONID")
		if user != "admin" || pass != "secret" || r.Method == http.MethodPost && (r.Header.Get("Jenkins-Crumb") != "c1" || cookie == nil) {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		return true
	}
	mux.HandleFunc("GET /crumbIssuer/api/json", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "s1", Path: "/"})
		_, _ = io.WriteString(w, `{"crumbRequestField":"Jenkins-Crumb","crumb":"c1"}`)
	})
	mux.HandleFunc("GET /job/{name}/api/json", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := jobs[r.PathValue("name")]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `{}`)
	})
	mux.HandleFunc("POST /createItem", func(w http.ResponseWriter, r *http.Request) {
		if requireAuthAndCrumb(w, r) {
			body, _ := io.ReadAll(r.Body)
			jobs[r.URL.Query().Get("name")] = "created:" + string(body)
		}
	})
	mux.HandleFunc("POST /job/{name}/config.xml", func(w http.ResponseWriter, r *http.Request) {
		if requireAuthAndCrumb(w, r) {
			body, _ := io.ReadAll(r.Body)
			jobs[r.PathValue("name")] = "updated:" + string(body)
		}
	})
	mux.HandleFunc("POST /job/{name}/buildWithParameters", func(w http.ResponseWriter, r *http.Request) {
		if requireAuthAndCrumb(w, r) {
			_ = r.ParseForm()
			triggered = r.PostForm
			// Jenkins builds absolute URLs from its configured root URL, not from the address it was called on
			w.Header().Set("Location", "http://jenkins.public:8080/queue/item/7/")
			w.WriteHeader(http.StatusCreated)
		}
	})
	mux.HandleFunc("GET /queue/item/7/api/json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"cancelled":false,"executable":{"url":"http://jenkins.public:8080/job/app/3/"}}`)
	})
	mux.HandleFunc("GET /job/app/3/api/json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"building":false,"result":"SUCCESS","number":3}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx := t.Context()
	c := NewClient(srv.URL+"/", "admin", "secret", http.DefaultTransport)

	require.NoError(t, c.CreateOrUpdateJob(ctx, "app", "<v1/>"))
	assert.Equal(t, "created:<v1/>", jobs["app"])
	require.NoError(t, c.CreateOrUpdateJob(ctx, "app", "<v2/>"))
	assert.Equal(t, "updated:<v2/>", jobs["app"])

	queueURL, err := c.TriggerBuild(ctx, "app", url.Values{"COMMIT_SHA": {"abc"}})
	require.NoError(t, err)
	assert.Equal(t, "abc", triggered.Get("COMMIT_SHA"))

	buildURL, cancelled, err := c.QueueItem(ctx, queueURL)
	require.NoError(t, err)
	assert.False(t, cancelled)
	build, err := c.Build(ctx, buildURL)
	require.NoError(t, err)
	assert.Equal(t, &Build{Result: "SUCCESS", Number: 3}, build)

	_, err = NewClient(srv.URL, "admin", "wrong", http.DefaultTransport).TriggerBuild(ctx, "app", nil)
	assert.ErrorContains(t, err, "returned 403")
}
