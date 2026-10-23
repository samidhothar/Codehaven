<script lang="ts" setup>
import {computed, nextTick, ref, shallowRef, useTemplateRef, type ShallowRef} from 'vue';
import SvgIcon from './SvgIcon.vue';
import {GET, POST} from '../modules/fetch.ts';
import {hideModal} from '../modules/modal.ts';
import {copyToClipboard} from '../modules/clipboard.ts';
import {showErrorToast, showSuccessToast} from '../modules/toast.ts';
import {debounce} from '../utils/func.ts';
import {generateElemId} from '../utils/dom.ts';
import {pathEscape, urlQueryEscape} from '../utils/url.ts';
import {parseIssueListQuickGotoLink} from '../features/common-issue-list.ts';
import type {CommandPaletteItem, CommandPaletteLocale, CommandPaletteFilter} from '../features/command-palette.ts';
import {toIssueResultItems, type IssueResultIcon, type IssueSearchResult} from '../features/command-palette-issues.ts';

const props = defineProps<{
  pageLinks: CommandPaletteItem[],
  repoLink: string,
  signedIn: boolean,
  locale: CommandPaletteLocale,
  filterItems: CommandPaletteFilter,
}>();

type ResultIcon = IssueResultIcon | 'octicon-link' | 'octicon-copy' | 'octicon-eye' | 'octicon-repo' | 'octicon-lock' | '';
type ResultItem = CommandPaletteItem & {icon: ResultIcon, avatar?: string, run?: () => Promise<void> | void, isHint?: boolean};
type ResultGroup = {name: string, items: ResultItem[]};

const {appSubUrl} = window.config;
const maxPageLinks = 8;
const remoteLimit = 5;

const refInput = useTemplateRef('input') as Readonly<ShallowRef<HTMLInputElement>>;
const refList = useTemplateRef<HTMLElement>('list');
const listId = generateElemId('command-palette-list-');

const query = ref('');
const selectedIndex = ref(0);
const isLoading = ref(false);
const repoResults = shallowRef<ResultItem[]>([]);
const userResults = shallowRef<ResultItem[]>([]);
const issueResults = shallowRef<ResultItem[]>([]);
const themeItems = shallowRef<ResultItem[]>([]);
let searchSeq = 0;

const actionItems = computed<ResultItem[]>(() => {
  const items: ResultItem[] = [];
  const hasRepoTab = (tab: string) => props.pageLinks.some((link) => link.url === `${props.repoLink}/${tab}`);
  if (props.repoLink && hasRepoTab('issues')) {
    items.push({label: props.locale.newIssue, context: '', url: `${props.repoLink}/issues/new`, icon: 'octicon-issue-opened'});
  }
  if (props.repoLink && hasRepoTab('pulls')) {
    items.push({label: props.locale.newPull, context: '', url: `${props.repoLink}/compare`, icon: 'octicon-git-pull-request'});
  }
  items.push({label: props.locale.copyLink, context: '', url: window.location.href, icon: 'octicon-copy', run: copyPageLink});
  return [...items, ...themeItems.value];
});

async function copyPageLink() {
  const {copy_success, copy_error} = window.config.i18n;
  const ok = await copyToClipboard(window.location.href);
  if (ok) showSuccessToast(copy_success); else showErrorToast(copy_error);
}

async function loadThemeItems() {
  if (themeItems.value.length) return;
  const resp = await GET(`${appSubUrl}/-/web-theme/list`);
  const results: Array<{name: string, value: string}> = (await resp.json()).results ?? [];
  themeItems.value = results.map((theme) => {
    // "name" is server-rendered theme item HTML, only its text is wanted
    const displayName = new DOMParser().parseFromString(theme.name, 'text/html').body.textContent.trim();
    return {
      label: props.locale.theme.replace('%s', () => displayName), context: '', url: '', icon: 'octicon-eye',
      run: async () => {
        await POST(`${appSubUrl}/-/web-theme/apply?theme=${urlQueryEscape(theme.value)}`);
        window.location.reload();
      },
    };
  });
}

const groups = computed<ResultGroup[]>(() => {
  const issueUrl = parseIssueListQuickGotoLink(props.repoLink, query.value);
  const issueLabel = props.locale.goToIssue.replace('%s', () => query.value.trim());
  const pages = props.filterItems(props.pageLinks, query.value).slice(0, maxPageLinks);
  // themes are only listed once the user searches, to keep the default view short
  const actions = query.value.trim() ? props.filterItems(actionItems.value, query.value) : actionItems.value.filter((item) => !themeItems.value.includes(item));
  const result: ResultGroup[] = [
    {name: '', items: issueUrl ? [{label: issueLabel, context: '', url: issueUrl, icon: 'octicon-issue-opened'}] : []},
    {name: props.locale.actions, items: actions.slice(0, maxPageLinks)},
    {name: props.locale.pages, items: pages.map((item): ResultItem => ({...item, icon: 'octicon-link'}))},
    {name: props.locale.issues, items: issueResults.value},
    {name: props.locale.repositories, items: props.signedIn ? repoResults.value : signInForReposItems()},
    {name: props.locale.people, items: userResults.value},
  ];
  return result.filter((group) => group.items.length);
});

// repository search requires signing in, so explain the missing results instead of silently showing none
function signInForReposItems(): ResultItem[] {
  if (!query.value.trim()) return [];
  const redirectTo = urlQueryEscape(`${window.location.pathname}${window.location.search}`);
  return [{label: props.locale.signInRepos, context: '', url: `${appSubUrl}/user/login?redirect_to=${redirectTo}`, icon: 'octicon-lock', isHint: true}];
}

const flatItems = computed(() => groups.value.flatMap((group) => group.items));

const resultsAnnouncement = computed(() => {
  if (!query.value.trim() || isLoading.value) return '';
  const count = flatItems.value.filter((item) => !item.isHint).length;
  if (!count) return props.locale.noResults;
  return count === 1 ? props.locale.resultOne : props.locale.results.replace('%d', () => String(count));
});

async function readResultData(resp: Response | null) {
  if (!resp?.ok) return [];
  try {
    return (await resp.json()).data ?? [];
  } catch {
    return []; // anonymous user search may redirect to the sign-in page
  }
}

async function readIssueResults(resp: Response | null): Promise<IssueSearchResult[]> {
  if (!resp?.ok) return [];
  try {
    return await resp.json();
  } catch {
    return [];
  }
}

async function fetchRemote(q: string) {
  const seq = ++searchSeq;
  isLoading.value = true;
  try {
    // a bare "#123" is handled by the issue jump, searching titles for it only adds noise
    const searchIssues = q && !/^#?\d+$/.test(q);
    const [repoResp, userResp, issueResp] = await Promise.all([
      props.signedIn ? GET(`${appSubUrl}/repo/search?q=${urlQueryEscape(q)}&limit=${remoteLimit}&sort=updated&order=desc`) : null,
      q ? GET(`${appSubUrl}/user/search_candidates?q=${urlQueryEscape(q)}&orgs=true`) : null,
      searchIssues ? GET(`${appSubUrl}/api/v1/repos/issues/search?q=${urlQueryEscape(q)}&state=all&limit=${remoteLimit}`) : null,
    ]);
    const [repoData, userData, issueData] = await Promise.all([readResultData(repoResp), readResultData(userResp), readIssueResults(issueResp)]);
    if (seq !== searchSeq || q !== query.value.trim()) return;
    issueResults.value = toIssueResultItems(issueData, props.repoLink.substring(appSubUrl.length + 1), appSubUrl);
    repoResults.value = repoData.map(({repository: repo}: {repository: {full_name: string, link: string, private: boolean}}) => ({
      label: repo.full_name, context: '', url: repo.link, icon: repo.private ? 'octicon-lock' : 'octicon-repo',
    }));
    userResults.value = userData.slice(0, remoteLimit).map((user: {login: string, full_name: string, avatar_url: string}) => ({
      label: user.login, context: user.full_name, url: `${appSubUrl}/${pathEscape(user.login)}`, icon: '', avatar: user.avatar_url,
    }));
  } finally {
    if (seq === searchSeq) isLoading.value = false;
  }
}

const fetchRemoteDebounced = debounce(fetchRemote, 250);

function onInput() {
  query.value = refInput.value.value;
  selectedIndex.value = 0;
  fetchRemoteDebounced(query.value.trim());
}

function selectIndex(idx: number) {
  selectedIndex.value = idx;
  nextTick(() => refList.value?.querySelector('[aria-selected="true"]')?.scrollIntoView({block: 'nearest'}));
}

function openItem(item: ResultItem, newTab: boolean) {
  if (item.run) {
    hideModal(refInput.value.closest('dialog')!);
    item.run();
    return;
  }
  if (newTab) {
    window.open(item.url, '_blank', 'noopener');
    hideModal(refInput.value.closest('dialog')!);
    return;
  }
  window.location.assign(item.url);
}

function onKeyDown(e: KeyboardEvent) {
  if (e.isComposing) return;
  const count = flatItems.value.length;
  if (e.key === 'ArrowDown' && count) {
    e.preventDefault();
    selectIndex((selectedIndex.value + 1) % count);
  } else if (e.key === 'ArrowUp' && count) {
    e.preventDefault();
    selectIndex((selectedIndex.value - 1 + count) % count);
  } else if (e.key === 'Enter') {
    e.preventDefault();
    const item = flatItems.value[selectedIndex.value];
    if (item) openItem(item, e.ctrlKey || e.metaKey);
  }
}

function indexOf(item: ResultItem) {
  return flatItems.value.indexOf(item);
}

function onOpen() {
  refInput.value.value = '';
  query.value = '';
  selectedIndex.value = 0;
  refInput.value.focus();
  fetchRemote('');
  loadThemeItems();
}

defineExpose({onOpen});
</script>

<template>
  <div class="command-palette">
    <div class="command-palette-input">
      <SvgIcon name="octicon-search"/>
      <input
        ref="input" type="text" autocomplete="off" spellcheck="false" :placeholder="locale.placeholder" :aria-label="locale.placeholder"
        role="combobox" aria-autocomplete="list" aria-expanded="true" :aria-controls="listId"
        :aria-activedescendant="flatItems.length ? `${listId}-${selectedIndex}` : undefined"
        @input="onInput" @keydown="onKeyDown"
      >
      <div v-if="isLoading" class="is-loading command-palette-spinner"/>
    </div>
    <div :id="listId" ref="list" role="listbox" class="command-palette-results">
      <div v-for="group in groups" :key="group.name" role="group" :aria-label="group.name || undefined">
        <div v-if="group.name" class="command-palette-group">{{ group.name }}</div>
        <a
          v-for="item in group.items" :key="`${item.url}|${item.label}`" :id="`${listId}-${indexOf(item)}`" :href="item.run ? undefined : item.url"
          role="option" :aria-selected="indexOf(item) === selectedIndex" class="command-palette-item"
          tabindex="-1" @mousemove="selectedIndex = indexOf(item)" @click.prevent="openItem(item, $event.ctrlKey || $event.metaKey)"
        >
          <img v-if="item.avatar" class="ui avatar command-palette-icon" :src="item.avatar" width="16" height="16" alt="">
          <SvgIcon v-else-if="item.icon" :name="item.icon" class="command-palette-icon"/>
          <span class="gt-ellipsis">{{ item.label }}</span>
          <span v-if="item.context" class="command-palette-context gt-ellipsis">{{ item.context }}</span>
        </a>
      </div>
      <div v-if="!flatItems.length && !isLoading" class="command-palette-empty">{{ locale.noResults }}</div>
    </div>
    <div class="command-palette-hint">{{ locale.hint }}</div>
    <div class="tw-sr-only" aria-live="polite">{{ resultsAnnouncement }}</div>
  </div>
</template>

<style scoped>
.command-palette-input {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-secondary);
}

.command-palette-input input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--color-text);
  font-size: 16px;
}

.command-palette-spinner {
  width: 16px;
  height: 16px;
}

.command-palette-results {
  max-height: min(60vh, 420px);
  overflow-y: auto;
  padding: 4px 0;
}

.command-palette-group {
  padding: 8px 16px 4px;
  font-size: 12px;
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-light-2);
}

.command-palette-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  color: var(--color-text);
}

.command-palette-item[aria-selected="true"] {
  background: var(--color-hover);
  color: var(--color-text);
  box-shadow: inset 3px 0 0 var(--color-primary);
}

.command-palette-icon {
  flex-shrink: 0;
  color: var(--color-text-light);
}

.command-palette-context {
  margin-left: auto;
  padding-left: 16px;
  color: var(--color-text-light-2);
  font-size: 12px;
}

.command-palette-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-light);
}

.command-palette-hint {
  padding: 6px 16px;
  border-top: 1px solid var(--color-secondary);
  font-size: 12px;
  color: var(--color-text-light-2);
}
</style>
