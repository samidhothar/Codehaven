import {GET} from '../../modules/fetch.ts';
import {createElementFromAttrs} from '../../utils/dom.ts';

export type SlashCommand = {name: string, label: string, value: string, block?: boolean};

const builtinValues: Record<string, string> = {
  table: '| Column 1 | Column 2 |\n| --- | --- |\n|  |  |\n',
  checklist: '- [ ] \n- [ ] \n',
  details: '<details>\n<summary>Details</summary>\n\n\n\n</details>\n', // eslint-disable-line gitea/unescaped-html-literal -- markdown inserted into a textarea, never parsed as HTML
  code: '```\n\n```\n',
  mermaid: '```mermaid\ngraph TD\n  A --> B\n```\n',
  quote: '> ',
};

export function builtinSlashCommands(labels: Record<string, string>): SlashCommand[] {
  return Object.entries(builtinValues).map(([name, value]) => ({name, label: labels[name] ?? name, value, block: true}));
}

export function matchSlashCommands(commands: SlashCommand[], text: string): SlashCommand[] {
  const query = text.toLowerCase();
  return commands.filter((cmd) => cmd.name.startsWith(query) || cmd.label.toLowerCase().includes(query));
}

const savedRepliesCache = new Map<string, Promise<SlashCommand[]>>();

async function fetchSavedReplies(url: string, label: string): Promise<SlashCommand[]> {
  try {
    const resp = await GET(url);
    if (!resp.ok) return [];
    const replies: Array<{title: string, content: string}> = await resp.json();
    return replies.map((reply) => ({name: reply.title.toLowerCase(), label: `${label}: ${reply.title}`, value: reply.content}));
  } catch {
    return []; // the built-in commands still work without saved replies
  }
}

export function loadSavedReplyCommands(url: string, label: string): Promise<SlashCommand[]> {
  if (!savedRepliesCache.has(url)) savedRepliesCache.set(url, fetchSavedReplies(url, label));
  return savedRepliesCache.get(url)!;
}

export function renderSlashSuggestions(commands: SlashCommand[]): HTMLUListElement {
  const ul = createElementFromAttrs<HTMLUListElement>('ul', {class: 'suggestions'});
  for (const cmd of commands) {
    ul.append(createElementFromAttrs('li', {role: 'option', 'data-value': cmd.value, 'data-block': cmd.block},
      createElementFromAttrs('span', {class: 'name'}, `/${cmd.name}`),
      createElementFromAttrs('span', {class: 'fullname'}, cmd.label),
    ));
  }
  return ul;
}
