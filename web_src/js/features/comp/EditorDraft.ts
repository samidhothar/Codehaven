import {localUserSettings} from '../../modules/user-settings.ts';
import {debounce} from '../../utils/func.ts';
import {createElementFromHTML} from '../../utils/dom.ts';
import {html, htmlRaw} from '../../utils/html.ts';
import {svg} from '../../svg.ts';
import {EventEditorContentChanged} from './EditorMarkdown.ts';
import type {ComboMarkdownEditor} from './ComboMarkdownEditor.ts';

type EditorDrafts = Record<string, {text: string, time: number}>;

const settingKey = 'editor-drafts';
const maxDrafts = 20;
const maxAgeMs = 7 * 24 * 60 * 60 * 1000;

export function loadEditorDrafts(now: number): EditorDrafts {
  const drafts = localUserSettings.getJsonObject<EditorDrafts>(settingKey, {});
  const kept = Object.entries(drafts).filter(([_, draft]) => now - draft.time < maxAgeMs);
  kept.sort((a, b) => b[1].time - a[1].time);
  return Object.fromEntries(kept.slice(0, maxDrafts));
}

export function saveEditorDraft(key: string, text: string, now: number) {
  const drafts = loadEditorDrafts(now);
  if (text.trim()) {
    drafts[key] = {text, time: now};
  } else {
    delete drafts[key];
  }
  localUserSettings.setJsonObject(settingKey, drafts);
}

// only new content is drafted, editors that start with content (editing an existing comment) are left alone
export function initEditorDraft(editor: ComboMarkdownEditor) {
  const form = editor.textarea.closest('form');
  if (!form || editor.value()) return;

  const key = `${form.getAttribute('action') || window.location.pathname}#${editor.textarea.name}`;
  const save = debounce(() => saveEditorDraft(key, editor.value(), Date.now()), 500);
  const draft = loadEditorDrafts(Date.now())[key];
  // a visible notice, so restored text is never submitted without the user noticing it
  let notice: HTMLElement | null = null;
  if (draft) {
    editor.value(draft.text);
    const {editor_draft_restored, editor_draft_discard} = window.config.i18n;
    notice = createElementFromHTML(html`
      <div class="editor-draft-notice flex-text-block" role="status">
        ${htmlRaw(svg('octicon-history', 14))}<span>${editor_draft_restored}</span>
        <button type="button" class="ui mini basic button">${editor_draft_discard}</button>
      </div>
    `);
    notice.querySelector('button')!.addEventListener('click', () => {
      save.cancel();
      editor.value('');
      saveEditorDraft(key, '', Date.now());
      notice!.remove();
      editor.focus();
    });
    editor.container.before(notice);
  }

  editor.container.addEventListener(EventEditorContentChanged, () => save());
  form.addEventListener('submit', () => {
    save.cancel();
    saveEditorDraft(key, '', Date.now());
    notice?.remove();
  });
}
