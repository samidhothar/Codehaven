import {pathEscape} from '../utils/url.ts';

export type IssueSearchResult = {
  number: number,
  title: string,
  state: string,
  pull_request: {merged: boolean} | null,
  repository: {full_name: string},
};

export type IssueResultIcon = 'octicon-issue-opened' | 'octicon-issue-closed' | 'octicon-git-pull-request' | 'octicon-git-pull-request-closed' | 'octicon-git-merge';

export function issueResultIcon(issue: IssueSearchResult): IssueResultIcon {
  if (!issue.pull_request) return issue.state === 'open' ? 'octicon-issue-opened' : 'octicon-issue-closed';
  if (issue.pull_request.merged) return 'octicon-git-merge';
  return issue.state === 'open' ? 'octicon-git-pull-request' : 'octicon-git-pull-request-closed';
}

// issues of the current repository come first, the API orders the rest by relevance
export function toIssueResultItems(issues: IssueSearchResult[], currentRepoFullName: string, appSubUrl: string) {
  const isCurrent = (issue: IssueSearchResult) => Number(issue.repository.full_name === currentRepoFullName);
  return [...issues].sort((a, b) => isCurrent(b) - isCurrent(a)).map((issue) => ({
    label: issue.title,
    context: `${issue.repository.full_name}#${issue.number}`,
    url: `${appSubUrl}/${issue.repository.full_name.split('/').map(pathEscape).join('/')}/issues/${issue.number}`,
    icon: issueResultIcon(issue),
  }));
}
