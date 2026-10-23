import {createElementFromHTML} from '../utils/dom.ts';
import {html, htmlRaw} from '../utils/html.ts';
import {svg} from '../svg.ts';
import {tryQuoteSelectedReply} from './repo-issue-edit.ts';

const collapseMinHeight = 720; // only comments much taller than the collapsed height are worth a click

function initLongCommentCollapse() {
  const markups = document.querySelectorAll<HTMLElement>('.comment-list > .timeline-item.comment:not(.issue-content-comment, .form) .comment-container .render-content.markup');
  if (!markups.length) return;
  const {show_more, show_less} = window.config.i18n;

  const observer = new ResizeObserver((entries) => {
    for (const {target} of entries) {
      const markup = target as HTMLElement;
      if (markup.scrollHeight < collapseMinHeight || markup.nextElementSibling?.matches('.comment-collapse-toggle')) continue;
      // a link to this comment must show it entirely
      if (window.location.hash.length > 1 && markup.closest(`.timeline-item#${CSS.escape(window.location.hash.slice(1))}`)) continue;
      markup.classList.add('comment-collapsed');
      const toggle = createElementFromHTML<HTMLButtonElement>(html`<button type="button" class="comment-collapse-toggle" aria-expanded="false">${htmlRaw(svg('octicon-chevron-down', 14))}<span>${show_more}</span></button>`);
      toggle.addEventListener('click', () => {
        const collapsed = markup.classList.toggle('comment-collapsed');
        toggle.setAttribute('aria-expanded', String(!collapsed));
        toggle.replaceChildren(createElementFromHTML(svg(collapsed ? 'octicon-chevron-down' : 'octicon-chevron-up', 14)), createElementFromHTML(html`<span>${collapsed ? show_more : show_less}</span>`));
        if (collapsed) markup.closest('.timeline-item')!.scrollIntoView({block: 'nearest'});
      });
      markup.after(toggle);
      observer.unobserve(markup);
    }
  });
  for (const markup of markups) observer.observe(markup);
}

function selectedQuotableComment(): {range: Range, label: string} | null {
  const selection = window.getSelection();
  if (!selection?.rangeCount || selection.isCollapsed || !selection.toString().trim()) return null;
  const range = selection.getRangeAt(0);
  const node = range.commonAncestorContainer;
  const el = node instanceof Element ? node : node.parentElement;
  const container = el?.closest('.render-content.markup')?.closest('.comment-container');
  const quoteItem = container?.querySelector('[data-shortcut-handler="quoteSelectedReply"]');
  if (!quoteItem) return null;
  return {range, label: quoteItem.textContent.trim()};
}

function initFloatingQuoteReply() {
  if (!document.querySelector('.comment-list [data-shortcut-handler="quoteSelectedReply"]')) return;
  const btn = createElementFromHTML<HTMLButtonElement>(html`<button type="button" class="ui mini primary button floating-quote-reply tw-hidden"></button>`);
  document.body.append(btn);

  let pointerDown = false;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const update = () => {
    const quotable = pointerDown ? null : selectedQuotableComment();
    btn.classList.toggle('tw-hidden', !quotable);
    if (!quotable) return;
    btn.replaceChildren(createElementFromHTML(svg('octicon-quote', 14)), createElementFromHTML(html`<span>${quotable.label}</span>`));
    const rects = quotable.range.getClientRects();
    const rect = rects[rects.length - 1] ?? quotable.range.getBoundingClientRect();
    const left = Math.min(rect.right + window.scrollX, window.scrollX + document.documentElement.clientWidth - btn.offsetWidth - 8);
    btn.style.left = `${Math.max(window.scrollX + 8, left - btn.offsetWidth / 2)}px`;
    btn.style.top = `${rect.bottom + window.scrollY + 6}px`;
  };
  const schedule = () => {
    clearTimeout(timer);
    timer = setTimeout(update, 150);
  };

  document.addEventListener('pointerdown', (e) => {
    if (e.target !== btn && !btn.contains(e.target as Node)) pointerDown = true;
  });
  document.addEventListener('pointerup', () => {
    pointerDown = false;
    schedule();
  });
  document.addEventListener('selectionchange', schedule);
  btn.addEventListener('pointerdown', (e) => e.preventDefault()); // keeps the selection that is going to be quoted
  btn.addEventListener('click', () => {
    tryQuoteSelectedReply();
    btn.classList.add('tw-hidden');
  });
}

export function initRepoIssueCommentTools() {
  if (!document.querySelector('.repository.view.issue .comment-list')) return;
  initLongCommentCollapse();
  initFloatingQuoteReply();
}
