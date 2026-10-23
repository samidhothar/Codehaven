import {moveListNav} from './list-keyboard-nav.ts';

test('moveListNav', () => {
  const container = document.createElement('div');
  container.innerHTML = `<div id="issue-list">
    <div class="item"><a class="list-item-large-title" href="/a">A</a></div>
    <div class="item"><a class="list-item-large-title" href="/b">B</a></div>
  </div>`;
  document.body.append(container);
  onTestFinished(() => container.remove());
  const active = () => document.querySelector('#issue-list > .item.list-nav-active a')?.textContent;

  expect(moveListNav(-1)).toBe(true);
  expect(active()).toBe('B'); // starting backwards selects the last item
  expect(moveListNav(1)).toBe(true);
  expect(active()).toBe('B'); // clamped at the end
  moveListNav(-1);
  expect(active()).toBe('A');
  expect(document.activeElement?.textContent).toBe('A');
  expect(document.querySelectorAll('.list-nav-active')).toHaveLength(1);
});
