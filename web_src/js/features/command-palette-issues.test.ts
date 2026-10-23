import {issueResultIcon, toIssueResultItems, type IssueSearchResult} from './command-palette-issues.ts';

const issue = (fullName: string, number: number, extra: Partial<IssueSearchResult> = {}): IssueSearchResult => ({
  number, title: `t${number}`, state: 'open', pull_request: null, repository: {full_name: fullName}, ...extra,
});

test('issueResultIcon', () => {
  expect(issueResultIcon(issue('o/r', 1))).toBe('octicon-issue-opened');
  expect(issueResultIcon(issue('o/r', 1, {state: 'closed'}))).toBe('octicon-issue-closed');
  expect(issueResultIcon(issue('o/r', 1, {pull_request: {merged: false}}))).toBe('octicon-git-pull-request');
  expect(issueResultIcon(issue('o/r', 1, {state: 'closed', pull_request: {merged: false}}))).toBe('octicon-git-pull-request-closed');
  expect(issueResultIcon(issue('o/r', 1, {state: 'closed', pull_request: {merged: true}}))).toBe('octicon-git-merge');
});

test('toIssueResultItems', () => {
  const items = toIssueResultItems([issue('other/x', 7), issue('o/r', 3)], 'o/r', '/sub');
  expect(items.map((i) => i.url)).toEqual(['/sub/o/r/issues/3', '/sub/other/x/issues/7']); // current repo first
  expect(items[0]).toMatchObject({label: 't3', context: 'o/r#3'});
});
