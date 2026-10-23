import {GET} from '../modules/fetch.ts';
import {createTippy, getAttachedTippyInstance} from '../modules/tippy.ts';
import {addDelegatedEventListener, createElementFromHTML} from '../utils/dom.ts';
import {html, htmlRaw} from '../utils/html.ts';
import {pathEscape} from '../utils/url.ts';
import {svg} from '../svg.ts';
import type {TimeoutId} from '../types.ts';

export type UserCardInfo = {
  login: string,
  full_name: string,
  avatar_url: string,
  html_url: string,
  description: string,
  location: string,
  followers_count: number,
  following_count: number,
};

const userInfoCache = new Map<string, Promise<UserCardInfo | null>>();

async function fetchUserInfo(name: string): Promise<UserCardInfo | null> {
  const resp = await GET(`${window.config.appSubUrl}/api/v1/users/${pathEscape(name)}`);
  return resp.ok ? resp.json() : null;
}

// explicit "data-user-card" on template user links, or a markdown "@user" mention (team mentions contain "/")
export function getUserCardName(link: HTMLAnchorElement): string {
  const name = link.getAttribute('data-user-card');
  if (name) return name;
  const text = link.textContent.trim();
  if (!link.closest('.markup') || !text.startsWith('@') || text.includes('/')) return '';
  return text.substring(1);
}

export function renderUserCard(user: UserCardInfo, i18n: {followers: string, following: string}): string {
  return html`
    <div class="user-card">
      <a class="user-card-header" href="${user.html_url}">
        <img class="ui avatar" src="${user.avatar_url}" width="48" height="48" alt="">
        <span class="user-card-names">
          ${user.full_name ? htmlRaw(html`<strong>${user.full_name}</strong>`) : ''}
          <span class="user-card-login">@${user.login}</span>
        </span>
      </a>
      ${user.description ? htmlRaw(html`<p class="user-card-description">${user.description}</p>`) : ''}
      <div class="user-card-meta">
        ${user.location ? htmlRaw(html`<span class="flex-text-inline">${htmlRaw(svg('octicon-location', 14))}${user.location}</span>`) : ''}
        <span class="flex-text-inline">${htmlRaw(svg('octicon-people', 14))}<strong>${user.followers_count}</strong> ${i18n.followers} · <strong>${user.following_count}</strong> ${i18n.following}</span>
      </div>
    </div>
  `;
}

async function showUserCard(link: HTMLAnchorElement, name: string) {
  if (!userInfoCache.has(name)) userInfoCache.set(name, fetchUserInfo(name));
  const user = await userInfoCache.get(name)!;
  if (!user) return;
  const {user_followers, user_following} = window.config.i18n;
  const tippy = createTippy(link, {
    theme: 'default',
    content: createElementFromHTML(renderUserCard(user, {followers: user_followers, following: user_following})),
    trigger: 'mouseenter focus',
    placement: 'top-start',
    interactive: true,
    role: 'dialog',
    interactiveBorder: 5,
  });
  link.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') tippy.hide();
  });
  tippy.show();
}

async function showUserCardOrReset(link: HTMLAnchorElement, name: string) {
  try {
    await showUserCard(link, name);
  } catch (err) {
    console.error('Failed to load user info:', err);
    link.removeAttribute('data-user-card-popup');
  }
}

function claimUserCardLink(link: HTMLAnchorElement): string {
  const name = getUserCardName(link);
  if (!name || getAttachedTippyInstance(link)) return '';
  link.setAttribute('data-user-card-popup', '');
  return name;
}

export function initUserCard() {
  // keyboard focus is deliberate, so the card opens without the hover delay
  addDelegatedEventListener<HTMLAnchorElement, FocusEvent>(document, 'focusin', 'a[href]:not([data-user-card-popup])', (link) => {
    const name = claimUserCardLink(link);
    if (name) showUserCardOrReset(link, name);
  });

  addDelegatedEventListener<HTMLAnchorElement, MouseEvent>(document, 'mouseover', 'a[href]:not([data-user-card-popup])', (link) => {
    const name = claimUserCardLink(link);
    if (!name) return;

    // delay so a mouse passing over the link doesn't fire a fetch
    let timer: TimeoutId;
    const cancel = () => {
      clearTimeout(timer);
      link.removeAttribute('data-user-card-popup');
      link.removeEventListener('mouseleave', cancel);
    };
    timer = setTimeout(() => {
      link.removeEventListener('mouseleave', cancel);
      showUserCardOrReset(link, name);
    }, 400);
    link.addEventListener('mouseleave', cancel);
  });
}
