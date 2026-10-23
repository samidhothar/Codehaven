import {loadEditorDrafts, saveEditorDraft} from './EditorDraft.ts';
import {localUserSettings} from '../../modules/user-settings.ts';

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
