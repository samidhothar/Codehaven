import {registerGlobalInitFunc} from '../modules/observer.ts';

function scrollBehavior(): ScrollBehavior {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth';
}

// returns the comment to jump to, comments under the bar count as already passed
export function findSiblingComment(comments: Element[], offset: number, direction: 'prev' | 'next'): Element | null {
  if (direction === 'next') return comments.find((el) => el.getBoundingClientRect().top > offset + 1) ?? null;
  return comments.findLast((el) => el.getBoundingClientRect().top < offset - 1) ?? null;
}

function initIssueStickyBar(bar: HTMLElement) {
  const header = document.querySelector('.issue-title-header');
  if (!header) return;

  const state = document.querySelector('.issue-title-meta .issue-state-label');
  if (state) bar.querySelector('.issue-sticky-bar-state')!.append(state.cloneNode(true));

  const statusButton = document.querySelector<HTMLButtonElement>('#status-button');
  const stickyStatus = bar.querySelector<HTMLButtonElement>('[data-sticky-action="status"]')!;
  if (statusButton) {
    // the status button text follows the comment editor ("Close" or "Close with comment")
    const syncStatus = () => stickyStatus.replaceChildren(...[...statusButton.childNodes].map((node) => node.cloneNode(true)));
    new MutationObserver(syncStatus).observe(statusButton, {subtree: true, childList: true, characterData: true});
    syncStatus();
    stickyStatus.classList.remove('tw-hidden');
  }
  const commentForm = document.querySelector<HTMLFormElement>('#comment-form');
  if (commentForm) bar.querySelector('[data-sticky-action="comment"]')!.classList.remove('tw-hidden');

  new IntersectionObserver(([entry]) => {
    const show = !entry.isIntersecting && entry.boundingClientRect.top < 0;
    bar.classList.toggle('visible', show);
    bar.inert = !show;
  }).observe(header);

  bar.addEventListener('click', (e) => {
    const action = (e.target as Element).closest('[data-sticky-action]')?.getAttribute('data-sticky-action');
    if (!action) return;
    const offset = bar.offsetHeight + 8;
    if (action === 'top') {
      window.scrollTo({top: 0, behavior: scrollBehavior()});
    } else if (action === 'prev' || action === 'next') {
      const comments = [...document.querySelectorAll('.comment-list > .timeline-item.comment:not(.form)')];
      const target = findSiblingComment(comments, offset, action);
      if (target) window.scrollTo({top: window.scrollY + target.getBoundingClientRect().top - offset, behavior: scrollBehavior()});
    } else if (action === 'status') {
      statusButton!.click();
    } else if (action === 'comment') {
      commentForm!.scrollIntoView({block: 'center', behavior: scrollBehavior()});
      commentForm!.querySelector<HTMLTextAreaElement>('textarea')?.focus({preventScroll: true});
    }
  });
}

export function initRepoIssueStickyBar() {
  registerGlobalInitFunc('initIssueStickyBar', initIssueStickyBar);
}
