import {findSiblingComment} from './repo-issue-sticky-bar.ts';

test('findSiblingComment', () => {
  const comments = [-500, 50, 400].map((top) => {
    const el = document.createElement('div');
    el.getBoundingClientRect = () => ({top} as DOMRect);
    return el;
  });
  expect(findSiblingComment(comments, 50, 'next')).toBe(comments[2]); // the comment right under the bar is the current one
  expect(findSiblingComment(comments, 50, 'prev')).toBe(comments[0]);
  expect(findSiblingComment(comments, 450, 'next')).toBeNull();
  expect(findSiblingComment(comments, -600, 'prev')).toBeNull();
});
