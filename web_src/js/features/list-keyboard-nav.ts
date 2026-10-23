import {registerShortcutHandler} from '../modules/shortcut.ts';

const activeClass = 'list-nav-active';

function listItems(): HTMLElement[] {
  return Array.from(document.querySelectorAll<HTMLElement>('#issue-list > .item'));
}

function activeItem(): HTMLElement | null {
  return document.querySelector<HTMLElement>(`#issue-list > .item.${activeClass}`);
}

export function moveListNav(delta: number): boolean {
  const items = listItems();
  if (!items.length) return false;
  const current = activeItem();
  const currentIdx = current ? items.indexOf(current) : -1;
  const nextIdx = currentIdx === -1 ? (delta > 0 ? 0 : items.length - 1) : Math.min(Math.max(currentIdx + delta, 0), items.length - 1);
  current?.classList.remove(activeClass);
  const next = items[nextIdx];
  next.classList.add(activeClass);
  // focusing the title link lets screen readers announce the item and native Enter open it
  next.querySelector<HTMLElement>('.list-item-large-title')!.focus({preventScroll: true});
  next.scrollIntoView({block: 'nearest'});
  return true;
}

export function initListKeyboardNav() {
  registerShortcutHandler('listNavNext', () => moveListNav(1));
  registerShortcutHandler('listNavPrevious', () => moveListNav(-1));
  registerShortcutHandler('listNavOpen', () => {
    const link = activeItem()?.querySelector<HTMLAnchorElement>('.list-item-large-title');
    if (!link) return false;
    window.location.assign(link.href);
    return true;
  });
  registerShortcutHandler('listNavSelect', () => {
    const checkbox = activeItem()?.querySelector<HTMLInputElement>('.issue-checkbox');
    if (!checkbox) return false;
    checkbox.click();
    return true;
  });
}
