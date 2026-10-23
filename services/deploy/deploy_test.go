// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deploy

import (
	"net/url"
	"strings"
	"testing"

	"gitea.dev/models/unittest"
	"gitea.dev/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validEnv() *Environment {
	return &Environment{Name: "staging", Trigger: "push", Branch: "main", Preset: "command", Host: "vm.example.com", Port: 22, User: "deploy", Path: "/srv/app", Command: "systemctl restart app"}
}

func TestEnvironmentValidate(t *testing.T) {
	require.NoError(t, validEnv().Validate())

	static := validEnv()
	static.Preset = "static"
	require.NoError(t, static.Validate())
	assert.Empty(t, static.Command, "the command only applies to the custom command preset")

	for name, mutate := range map[string]func(*Environment){
		"unknown environment": func(e *Environment) { e.Name = "qa" },
		"unknown trigger":     func(e *Environment) { e.Trigger = "cron" },
		"unknown preset":      func(e *Environment) { e.Preset = "ansible" },
		"empty branch":        func(e *Environment) { e.Branch = "" },
		"shell in host":       func(e *Environment) { e.Host = "vm;rm -rf /" },
		"option as host":      func(e *Environment) { e.Host = "-oProxyCommand=x" },
		"bad port":            func(e *Environment) { e.Port = 70000 },
		"shell in user":       func(e *Environment) { e.User = "root$(id)" },
		"quote in path":       func(e *Environment) { e.Path = "/srv/'app" },
		"relative path":       func(e *Environment) { e.Path = "srv/app" },
		"parent in path":      func(e *Environment) { e.Path = "/srv/../etc" },
		"multi-line command":  func(e *Environment) { e.Command = "a\nb" },
		"empty command":       func(e *Environment) { e.Command = " " },
	} {
		env := validEnv()
		mutate(env)
		assert.ErrorIs(t, env.Validate(), util.ErrInvalidArgument, name)
	}
}

func TestPipelineConfigXML(t *testing.T) {
	xml, err := pipelineConfigXML("codehaven-vm-ssh")
	require.NoError(t, err)
	assert.Contains(t, xml, "sshagent(credentials: [&#39;codehaven-vm-ssh&#39;])")
	assert.Contains(t, xml, "<name>DEPLOY_COMMAND</name>")
	assert.Contains(t, xml, "<sandbox>true</sandbox>")
	assert.Contains(t, xml, "rm -rf src &amp;&amp; git init")
	assert.NotContains(t, xml, "&& ", "the script must be XML-escaped")

	_, err = pipelineConfigXML("x']) { sh 'id' } //")
	assert.Error(t, err)
}

func TestJenkinsPolicy(t *testing.T) {
	check := func(policyURL, target string) error {
		p, err := jenkinsPolicy(policyURL)
		require.NoError(t, err)
		u, err := url.Parse(target)
		require.NoError(t, err)
		return p.CheckHost(u)
	}
	require.NoError(t, check("http://jenkins:8080", "http://jenkins:8080/job/x/api/json"))
	require.NoError(t, check("https://ci.example.com", "https://ci.example.com/api/json"))
	require.Error(t, check("http://jenkins:8080", "http://jenkins:9000/"), "other ports of the Jenkins host are not reachable")
	require.Error(t, check("http://jenkins:8080", "http://gitea:3000/"), "other hosts are not reachable")
}

func TestEnvironmentsStorage(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	ctx := t.Context()

	envs, err := GetEnvironments(ctx, 1)
	require.NoError(t, err)
	assert.Empty(t, envs)

	require.NoError(t, SaveEnvironment(ctx, 1, validEnv()))
	require.NoError(t, setLastRun(ctx, 1, "staging", &LastRun{Status: "success", CommitSHA: "abc"}))

	changed := validEnv()
	changed.Host = "vm2.example.com"
	require.NoError(t, SaveEnvironment(ctx, 1, changed))
	env, err := GetEnvironment(ctx, 1, "staging")
	require.NoError(t, err)
	assert.Equal(t, "vm2.example.com", env.Host)
	assert.Equal(t, "abc", env.Last.CommitSHA, "editing an environment keeps its last run")

	invalid := validEnv()
	invalid.Host = "bad host"
	assert.ErrorIs(t, SaveEnvironment(ctx, 1, invalid), util.ErrInvalidArgument)

	require.NoError(t, DeleteEnvironment(ctx, 1, "staging"))
	_, err = GetEnvironment(ctx, 1, "staging")
	assert.ErrorIs(t, err, util.ErrNotExist)
	assert.True(t, strings.HasPrefix(jobName(1, "staging"), "codehaven-1-"))
}
