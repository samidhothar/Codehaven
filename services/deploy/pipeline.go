// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deploy

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// pipelineParams are passed on every build, so user input never becomes part of the pipeline script
var pipelineParams = []string{"CLONE_URL", "BRANCH", "COMMIT_SHA", "PRESET", "VM_HOST", "VM_PORT", "VM_USER", "VM_PATH", "DEPLOY_COMMAND"}

// the remote "cd" quotes VM_PATH with single quotes, Validate guarantees the path contains none
const pipelineScript = `pipeline {
  agent any
  options {
    timeout(time: 30, unit: 'MINUTES')
    disableConcurrentBuilds()
  }
  stages {
    stage('Checkout') {
      steps {
        sh '''
          rm -rf src && git init -q src
          git -C src fetch -q "$CLONE_URL" "+refs/heads/$BRANCH:refs/remotes/origin/$BRANCH"
          git -C src checkout -q "$COMMIT_SHA"
        '''
      }
    }
    stage('Deploy') {
      steps {
        sshagent(credentials: ['%s']) {
          sh '''
            ssh_vm() { ssh -p "$VM_PORT" -o StrictHostKeyChecking=accept-new -o BatchMode=yes "$VM_USER@$VM_HOST" "$@"; }
            ssh_vm "mkdir -p '$VM_PATH'"
            tar -C src --exclude=.git -czf - . | ssh_vm "tar -C '$VM_PATH' -xzf -"
            case "$PRESET" in
              compose) ssh_vm "cd '$VM_PATH' && docker compose up -d --build" ;;
              command) ssh_vm "cd '$VM_PATH' && $DEPLOY_COMMAND" ;;
            esac
          '''
        }
      }
    }
  }
}
`

var validCredentialsID = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func jobName(repoID int64, env string) string {
	return fmt.Sprintf("codehaven-%d-%s", repoID, env)
}

func pipelineConfigXML(sshCredentialsID string) (string, error) {
	if !validCredentialsID.MatchString(sshCredentialsID) {
		return "", fmt.Errorf("pipelineConfigXML: invalid SSH credentials ID %q", sshCredentialsID)
	}
	esc := func(s string) string {
		var buf bytes.Buffer
		_ = xml.EscapeText(&buf, []byte(s))
		return buf.String()
	}
	var params strings.Builder
	for _, name := range pipelineParams {
		fmt.Fprintf(&params, "<hudson.model.StringParameterDefinition><name>%s</name><trim>true</trim></hudson.model.StringParameterDefinition>", name)
	}
	return `<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <description>Managed by CodeHaven, manual changes are overwritten by the next deployment.</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <hudson.model.ParametersDefinitionProperty><parameterDefinitions>` + params.String() + `</parameterDefinitions></hudson.model.ParametersDefinitionProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script>` + esc(fmt.Sprintf(pipelineScript, sshCredentialsID)) + `</script>
    <sandbox>true</sandbox>
  </definition>
  <disabled>false</disabled>
</flow-definition>
`, nil
}
