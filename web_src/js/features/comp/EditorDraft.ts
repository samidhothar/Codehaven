import {localUserSettings} from '../../modules/user-settings.ts';
import {debounce} from '../../utils/func.ts';
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
  const draft = loadEditorDrafts(Date.now())[key];
  if (draft) editor.value(draft.text);

  const save = debounce(() => saveEditorDraft(key, editor.value(), Date.now()), 500);
  editor.container.addEventListener(EventEditorContentChanged, () => save());
  form.addEventListener('submit', () => {
    save.cancel();
    saveEditorDraft(key, '', Date.now());
  });
}
