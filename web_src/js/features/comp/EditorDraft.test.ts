import {initEditorDraft, loadEditorDrafts, saveEditorDraft} from './EditorDraft.ts';
import {localUserSettings} from '../../modules/user-settings.ts';
import type {ComboMarkdownEditor} from './ComboMarkdownEditor.ts';

test('initEditorDraft restores with a discardable notice', () => {
  onTestFinished(() => localUserSettings.setJsonObject('editor-drafts', {}));
  window.config.i18n.editor_draft_restored = 'Unsent draft restored';
  window.config.i18n.editor_draft_discard = 'Discard';
  const form = document.createElement('form');
  form.setAttribute('action', '/o/r/issues/1/comments');
  form.innerHTML = `<div class="combo-markdown-editor"><textarea name="content"></textarea></div>`;
  document.body.append(form);
  onTestFinished(() => form.remove());
  const textarea = form.querySelector('textarea')!;
  const editor = {
    textarea, container: form.querySelector('.combo-markdown-editor'), focus: () => textarea.focus(),
    value: (v?: string) => (v === undefined ? textarea.value : (textarea.value = v)),
  } as unknown as ComboMarkdownEditor;

  saveEditorDraft('/o/r/issues/1/comments#content', 'old text', Date.now());
  initEditorDraft(editor);
  expect(textarea.value).toBe('old text');
  const notice = form.querySelector('.editor-draft-notice')!;
  expect(notice.textContent).toContain('Unsent draft restored');

  notice.querySelector('button')!.click();
  expect(textarea.value).toBe('');
  expect(form.querySelector('.editor-draft-notice')).toBeNull();
  expect(loadEditorDrafts(Date.now())).toEqual({});
});

test('EditorDraft', () => {
  onTestFinished(() => localUserSettings.setJsonObject('editor-drafts', {}));
  const day = 24 * 60 * 60 * 1000;

  saveEditorDraft('/issues/1/comments#content', 'hello', 0);
  expect(loadEditorDrafts(day)['/issues/1/comments#content'].text).toBe('hello');
  expect(loadEditorDrafts(8 * day)).toEqual({}); // expired

  saveEditorDraft('/issues/1/comments#content', '  ', day);
  expect(loadEditorDrafts(day)).toEqual({}); // blank text removes the draft

  for (let i = 0; i < 25; i++) saveEditorDraft(`key-${i}`, 'text', i);
  const drafts = loadEditorDrafts(30);
  expect(Object.keys(drafts)).toHaveLength(20);
  expect(drafts['key-24']).toBeTruthy(); // newest drafts are kept
  expect(drafts['key-0']).toBeUndefined();
});
