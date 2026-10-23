// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"gitea.dev/models/db"
	system_model "gitea.dev/models/system"
	"gitea.dev/modules/git"
	"gitea.dev/modules/timeutil"
	"gitea.dev/modules/util"

	"xorm.io/builder"
)

var (
	EnvironmentNames = []string{"development", "staging", "production"}
	Triggers         = []string{"manual", "push"}
	Presets          = []string{"static", "compose", "command"}

	// these values end up in shell commands of the deployment pipeline, so only plain names are accepted
	validHost = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,252}$`)
	validUser = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	validPath = regexp.MustCompile(`^/[A-Za-z0-9._/-]{0,250}$`)
)

type LastRun struct {
	Status    string             `json:"status"` // a commit status state: pending, success, failure or error
	CommitSHA string             `json:"commit_sha"`
	BuildURL  string             `json:"build_url,omitempty"`
	Doer      string             `json:"doer"`
	Time      timeutil.TimeStamp `json:"time"`
}

type Environment struct {
	Name    string   `json:"name"`
	Trigger string   `json:"trigger"`
	Branch  string   `json:"branch"`
	Preset  string   `json:"preset"`
	Host    string   `json:"host"`
	Port    int      `json:"port"`
	User    string   `json:"user"`
	Path    string   `json:"path"`
	Command string   `json:"command,omitempty"`
	Last    *LastRun `json:"last,omitempty"`
}

func (env *Environment) Validate() error {
	switch {
	case !slices.Contains(EnvironmentNames, env.Name):
		return util.NewInvalidArgumentErrorf("unknown environment %q", env.Name)
	case !slices.Contains(Triggers, env.Trigger):
		return util.NewInvalidArgumentErrorf("unknown trigger %q", env.Trigger)
	case !slices.Contains(Presets, env.Preset):
		return util.NewInvalidArgumentErrorf("unknown deployment type %q", env.Preset)
	case env.Branch == "" || !git.IsValidRefPattern(env.Branch):
		return util.NewInvalidArgumentErrorf("invalid branch %q", env.Branch)
	case !validHost.MatchString(env.Host):
		return util.NewInvalidArgumentErrorf("invalid host %q", env.Host)
	case env.Port < 1 || env.Port > 65535:
		return util.NewInvalidArgumentErrorf("invalid port %d", env.Port)
	case !validUser.MatchString(env.User):
		return util.NewInvalidArgumentErrorf("invalid SSH user %q", env.User)
	case !validPath.MatchString(env.Path) || strings.Contains(env.Path, ".."):
		return util.NewInvalidArgumentErrorf("invalid target path %q", env.Path)
	case env.Preset == "command" && (strings.TrimSpace(env.Command) == "" || len(env.Command) > 500 || strings.ContainsAny(env.Command, "\r\n")):
		return util.NewInvalidArgumentErrorf("the custom command must be a single line of up to 500 characters")
	}
	if env.Preset != "command" {
		env.Command = ""
	}
	return nil
}

func settingKey(repoID int64) string {
	return "deploy.repo." + strconv.FormatInt(repoID, 10)
}

// environments are read-modify-written as one JSON value, so updates of the same instance are serialized
var environmentsMu sync.Mutex

func GetEnvironments(ctx context.Context, repoID int64) ([]*Environment, error) {
	s, has, err := db.Get[system_model.Setting](ctx, builder.Eq{"setting_key": settingKey(repoID)})
	if err != nil || !has || s.SettingValue == "" {
		return nil, err
	}
	var envs []*Environment
	if err := json.Unmarshal([]byte(s.SettingValue), &envs); err != nil {
		return nil, fmt.Errorf("GetEnvironments: %w", err)
	}
	return envs, nil
}

func GetEnvironment(ctx context.Context, repoID int64, name string) (*Environment, error) {
	envs, err := GetEnvironments(ctx, repoID)
	if err != nil {
		return nil, err
	}
	for _, env := range envs {
		if env.Name == name {
			return env, nil
		}
	}
	return nil, util.NewNotExistErrorf("deployment environment %q does not exist", name)
}

func updateEnvironments(ctx context.Context, repoID int64, update func([]*Environment) ([]*Environment, error)) error {
	environmentsMu.Lock()
	defer environmentsMu.Unlock()
	envs, err := GetEnvironments(ctx, repoID)
	if err != nil {
		return err
	}
	if envs, err = update(envs); err != nil {
		return err
	}
	data, err := json.Marshal(envs)
	if err != nil {
		return err
	}
	return system_model.SetSettings(ctx, map[string]string{settingKey(repoID): string(data)})
}

// SaveEnvironment adds the environment or replaces the one with the same name, keeping its last run
func SaveEnvironment(ctx context.Context, repoID int64, env *Environment) error {
	if err := env.Validate(); err != nil {
		return err
	}
	return updateEnvironments(ctx, repoID, func(envs []*Environment) ([]*Environment, error) {
		for i, existing := range envs {
			if existing.Name == env.Name {
				env.Last = existing.Last
				envs[i] = env
				return envs, nil
			}
		}
		return append(envs, env), nil
	})
}

func DeleteEnvironment(ctx context.Context, repoID int64, name string) error {
	return updateEnvironments(ctx, repoID, func(envs []*Environment) ([]*Environment, error) {
		return slices.DeleteFunc(envs, func(env *Environment) bool { return env.Name == name }), nil
	})
}

func setLastRun(ctx context.Context, repoID int64, name string, last *LastRun) error {
	return updateEnvironments(ctx, repoID, func(envs []*Environment) ([]*Environment, error) {
		for _, env := range envs {
			if env.Name == name {
				env.Last = last
			}
		}
		return envs, nil
	})
}
