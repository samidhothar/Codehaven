import {createApp} from 'vue';
import {registerGlobalInitFunc} from '../modules/observer.ts';
import {showModal} from '../modules/modal.ts';
import {calcMatchedWeight, strSubMatch} from './repo-findfile.ts';

export type CommandPaletteItem = {
  label: string,
  context: string,
  url: string,
};

export type CommandPaletteLocale = {
  placeholder: string,
  noResults: string,
  pages: string,
  repositories: string,
  people: string,
  goToIssue: string,
  hint: string,
  actions: string,
  newIssue: string,
  newPull: string,
  copyLink: string,
  theme: string,
  resultOne: string,
  results: string,
  signInRepos: string,
};

function firstTextContent(el: Element): string {
  const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT);
  for (let node = walker.nextNode(); node; node = walker.nextNode()) {
    const text = node.textContent!.trim();
    if (text) return text;
  }
  return '';
}

// Links already rendered on the page carry the server's permission checks and translations
export function collectPageLinks(doc: ParentNode, repoFullName: string): CommandPaletteItem[] {
  // the current repo's tabs are the most relevant, so they come first
  const groups: Array<[string, string]> = [
    ['.secondary-nav overflow-menu a.item[href], .overflow-menu-popup a.item[href]', repoFullName],
    ['#navbar a.item[href]', ''],
  ];
  const seen = new Set<string>();
  const items: CommandPaletteItem[] = [];
  for (const [selector, context] of groups) {
    for (const link of doc.querySelectorAll<HTMLAnchorElement>(selector)) {
      const url = link.getAttribute('href')!;
      if (url.startsWith('#') || url.endsWith('/user/logout') || seen.has(url)) continue;
      const label = link.getAttribute('aria-label') || firstTextContent(link);
      if (!label) continue;
      seen.add(url);
      items.push({label, context, url});
    }
  }
  return items;
}

export type CommandPaletteFilter = typeof filterCommandItems;

// returns -1 when the term doesn't match
function termMatchWeight(haystack: string, term: string): number {
  const idx = haystack.toLowerCase().indexOf(term);
  if (idx >= 0) {
    const atWordStart = idx === 0 || /[\s/_.:-]/.test(haystack[idx - 1]);
    return term.length * term.length * (atWordStart ? 4 : 2);
  }
  // scattered letters only count for longer terms, otherwise "ma" matches any label with an "m" before an "a"
  if (term.length < 4) return -1;
  const res = strSubMatch(haystack, term);
  const matchedSegments = Math.floor(res.length / 2);
  if (!matchedSegments || matchedSegments > Math.ceil(term.length / 2)) return -1;
  return calcMatchedWeight(res);
}

export function filterCommandItems<T extends CommandPaletteItem>(items: T[], query: string): T[] {
  const terms = query.trim().toLowerCase().split(/\s+/).filter(Boolean);
  if (!terms.length) return items;
  const matched: Array<{item: T, weight: number}> = [];
  for (const item of items) {
    // every term must match, in any order, so "demo issues" and "issues demo" both work
    let weight = 0;
    for (const term of terms) {
      const termWeight = termMatchWeight(`${item.label} ${item.context}`, term);
      if (termWeight < 0) {
        weight = -1;
        break;
      }
      weight += termWeight;
    }
    if (weight >= 0) matched.push({item, weight});
  }
  return matched.sort((a, b) => b.weight - a.weight).map((m) => m.item);
}

type PaletteHandle = {dialog: HTMLDialogElement, instance: {onOpen: () => void}};

let palettePromise: Promise<PaletteHandle> | null = null;

async function createCommandPalette(trigger: HTMLElement): Promise<PaletteHandle> {
  const {default: CommandPalette} = await import('../components/CommandPalette.vue');
  const dialog = document.createElement('dialog');
  dialog.className = 'modal-dialog command-palette-modal';
  dialog.setAttribute('aria-label', trigger.getAttribute('aria-label')!);
  const locale: CommandPaletteLocale = {
    placeholder: trigger.getAttribute('data-locale-placeholder')!,
    noResults: trigger.getAttribute('data-locale-no-results')!,
    pages: trigger.getAttribute('data-locale-pages')!,
    repositories: trigger.getAttribute('data-locale-repositories')!,
    people: trigger.getAttribute('data-locale-people')!,
    goToIssue: trigger.getAttribute('data-locale-go-to-issue')!,
    hint: trigger.getAttribute('data-locale-hint')!,
    actions: trigger.getAttribute('data-locale-actions')!,
    newIssue: trigger.getAttribute('data-locale-new-issue')!,
    newPull: trigger.getAttribute('data-locale-new-pull')!,
    copyLink: trigger.getAttribute('data-locale-copy-link')!,
    theme: trigger.getAttribute('data-locale-theme')!,
    resultOne: trigger.getAttribute('data-locale-result-one')!,
    results: trigger.getAttribute('data-locale-results')!,
    signInRepos: trigger.getAttribute('data-locale-sign-in-repos')!,
  };
  const instance = createApp(CommandPalette, {
    pageLinks: collectPageLinks(document, trigger.getAttribute('data-repo-full-name') ?? ''),
    repoLink: trigger.getAttribute('data-repo-link') ?? '',
    signedIn: trigger.hasAttribute('data-signed-in'),
    locale,
    filterItems: filterCommandItems,
  }).mount(dialog) as unknown as PaletteHandle['instance'];
  return {dialog, instance};
}

async function openCommandPalette(trigger: HTMLElement) {
  palettePromise ||= createCommandPalette(trigger);
  const {dialog, instance} = await palettePromise;
  showModal(dialog, {onShow: () => instance.onOpen()});
}

export function initCommandPalette() {
  registerGlobalInitFunc('initCommandPalette', (trigger: HTMLElement) => {
    trigger.addEventListener('click', () => openCommandPalette(trigger));
    document.addEventListener('keydown', (e) => {
      if (e.defaultPrevented || e.isComposing || e.repeat || e.altKey || e.shiftKey) return;
      if (!(e.ctrlKey || e.metaKey) || e.key.toLowerCase() !== 'k') return;
      if ((e.target as Element).closest('.cm-editor') || document.querySelector('.modal-dialog[open]')) return; // CodeMirror has its own Mod-k chords
      e.preventDefault();
      openCommandPalette(trigger);
    });
  });
}
