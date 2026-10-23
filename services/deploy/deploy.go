// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deploy

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	git_model "gitea.dev/models/git"
	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/commitstatus"
	"gitea.dev/modules/egress/policy"
	"gitea.dev/modules/graceful"
	"gitea.dev/modules/jenkins"
	"gitea.dev/modules/log"
	"gitea.dev/modules/setting"
	"gitea.dev/modules/timeutil"
	"gitea.dev/modules/util"
	commitstatus_service "gitea.dev/services/repository/commitstatus"
)

const pollInterval = 3 * time.Second

func Enabled() bool {
	return setting.Jenkins.Enabled
}

// jenkinsPolicy only lets requests reach the host and port of the configured Jenkins URL
func jenkinsPolicy(jenkinsURL string) (*policy.Policy, error) {
	u, err := url.Parse(jenkinsURL)
	if err != nil {
		return nil, fmt.Errorf("jenkinsPolicy: %w", err)
	}
	port := u.Port()
	if port == "" {
		port = util.Iif(u.Scheme == "https", "443", "80")
	}
	// an allow entry without a port only covers the default web ports, Jenkins usually listens on 8080
	return policy.NewPolicy("jenkins", policy.Strict, policy.WithAllow(net.JoinHostPort(u.Hostname(), port), "jenkins.URL")), nil
}

func newClient() (*jenkins.Client, error) {
	p, err := jenkinsPolicy(setting.Jenkins.URL)
	if err != nil {
		return nil, err
	}
	return jenkins.NewClient(setting.Jenkins.URL, setting.Jenkins.User, setting.Jenkins.Token, p.NewHTTPTransport()), nil
}

// publicURL rewrites a Jenkins URL for browsers, which may reach Jenkins under a different address than this server
func publicURL(jenkinsURL string) string {
	if u, err := url.Parse(jenkinsURL); err == nil {
		return setting.Jenkins.PublicURL + u.RequestURI()
	}
	return jenkinsURL
}

// Deploy starts deploying the commit to the environment, then follows the build in the background
func Deploy(ctx context.Context, repo *repo_model.Repository, doer *user_model.User, env *Environment, commitSHA string) error {
	if !setting.Jenkins.Enabled {
		return util.NewInvalidArgumentErrorf("Jenkins deployments are not configured")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	configXML, err := pipelineConfigXML(setting.Jenkins.SSHCredentialsID)
	if err != nil {
		return err
	}
	name := jobName(repo.ID, env.Name)
	if err := client.CreateOrUpdateJob(ctx, name, configXML); err != nil {
		return fmt.Errorf("Deploy: %w", err)
	}
	queueURL, err := client.TriggerBuild(ctx, name, url.Values{
		"CLONE_URL":      {setting.Jenkins.CloneBaseURL + "/" + repo.FullName() + ".git"},
		"BRANCH":         {env.Branch},
		"COMMIT_SHA":     {commitSHA},
		"PRESET":         {env.Preset},
		"VM_HOST":        {env.Host},
		"VM_PORT":        {strconv.Itoa(env.Port)},
		"VM_USER":        {env.User},
		"VM_PATH":        {env.Path},
		"DEPLOY_COMMAND": {env.Command},
	})
	if err != nil {
		return fmt.Errorf("Deploy: %w", err)
	}

	jobURL := setting.Jenkins.PublicURL + "/" + "job/" + url.PathEscape(name) + "/"
	report(ctx, repo, doer, env.Name, commitSHA, commitstatus.CommitStatusPending, jobURL)
	go follow(graceful.GetManager().ShutdownContext(), client, repo, doer, env.Name, commitSHA, queueURL, jobURL)
	return nil
}

var statusDescriptions = map[commitstatus.CommitStatusState]string{
	commitstatus.CommitStatusPending: "Deployment in progress",
	commitstatus.CommitStatusSuccess: "Deployed",
	commitstatus.CommitStatusFailure: "Deployment failed",
	commitstatus.CommitStatusError:   "Deployment result unknown",
}

// report records the state both as a commit status, so it shows next to the commit, and as the environment's last run
func report(ctx context.Context, repo *repo_model.Repository, doer *user_model.User, envName, commitSHA string, state commitstatus.CommitStatusState, targetURL string) {
	status := &git_model.CommitStatus{State: state, TargetURL: targetURL, Description: statusDescriptions[state], Context: "deploy/" + envName}
	if err := commitstatus_service.CreateCommitStatus(ctx, repo, doer, commitSHA, status); err != nil {
		log.Error("report: CreateCommitStatus for %s %s: %v", repo.FullName(), envName, err)
	}
	last := &LastRun{Status: string(state), CommitSHA: commitSHA, BuildURL: targetURL, Doer: doer.Name, Time: timeutil.TimeStampNow()}
	if err := setLastRun(ctx, repo.ID, envName, last); err != nil {
		log.Error("report: setLastRun for %s %s: %v", repo.FullName(), envName, err)
	}
}

func follow(ctx context.Context, client *jenkins.Client, repo *repo_model.Repository, doer *user_model.User, envName, commitSHA, queueURL, jobURL string) {
	deadline := time.Now().Add(time.Duration(setting.Jenkins.PollTimeout) * time.Second)
	var buildURL string
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollInterval):
		}
		if buildURL == "" {
			started, cancelled, err := client.QueueItem(ctx, queueURL)
			if err != nil {
				log.Warn("follow: %s %s: %v", repo.FullName(), envName, err) // transient errors are retried until the deadline
				continue
			}
			if cancelled {
				report(ctx, repo, doer, envName, commitSHA, commitstatus.CommitStatusError, jobURL)
				return
			}
			if buildURL = started; buildURL == "" {
				continue
			}
			report(ctx, repo, doer, envName, commitSHA, commitstatus.CommitStatusPending, publicURL(buildURL))
		}
		build, err := client.Build(ctx, buildURL)
		if err != nil {
			log.Warn("follow: %s %s: %v", repo.FullName(), envName, err)
			continue
		}
		if build.Building || build.Result == "" {
			continue
		}
		state := commitstatus.CommitStatusFailure
		if build.Result == "SUCCESS" {
			state = commitstatus.CommitStatusSuccess
		}
		report(ctx, repo, doer, envName, commitSHA, state, publicURL(buildURL))
		return
	}
	report(ctx, repo, doer, envName, commitSHA, commitstatus.CommitStatusError, jobURL)
}
