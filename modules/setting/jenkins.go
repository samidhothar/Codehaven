// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"strings"

	"gitea.dev/modules/log"
)

// Jenkins settings for deploying repositories through a Jenkins server
var Jenkins = struct {
	Enabled bool
	URL     string // used by the server to call the Jenkins API
	// PublicURL is used for links shown to users, when browsers reach Jenkins under another address
	PublicURL string
	User      string
	Token     string
	// CloneBaseURL is how Jenkins reaches this instance to clone repositories, e.g. inside a container network
	CloneBaseURL     string
	SSHCredentialsID string
	PollTimeout      int // seconds to follow a deployment before giving up on its result
}{
	SSHCredentialsID: "codehaven-vm-ssh",
	PollTimeout:      1800,
}

func loadJenkinsFrom(rootCfg ConfigProvider) {
	sec := rootCfg.Section("jenkins")
	Jenkins.URL = strings.TrimSuffix(sec.Key("URL").String(), "/")
	Jenkins.PublicURL = strings.TrimSuffix(sec.Key("PUBLIC_URL").MustString(Jenkins.URL), "/")
	Jenkins.User = sec.Key("USER").String()
	Jenkins.Token = loadSecret(sec, "TOKEN_URI", "TOKEN")
	Jenkins.CloneBaseURL = strings.TrimSuffix(sec.Key("CLONE_BASE_URL").MustString(AppURL), "/")
	Jenkins.SSHCredentialsID = sec.Key("SSH_CREDENTIALS_ID").MustString(Jenkins.SSHCredentialsID)
	Jenkins.PollTimeout = sec.Key("POLL_TIMEOUT").MustInt(Jenkins.PollTimeout)
	Jenkins.Enabled = sec.Key("ENABLED").MustBool(false) && Jenkins.URL != ""
	if sec.Key("ENABLED").MustBool(false) && !Jenkins.Enabled {
		log.Error("[jenkins] ENABLED is true but URL is empty, Jenkins deployments are disabled")
	}
}
