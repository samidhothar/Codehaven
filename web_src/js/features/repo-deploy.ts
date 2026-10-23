import {registerGlobalInitFunc} from '../modules/observer.ts';
import {GET} from '../modules/fetch.ts';
import {createTippy} from '../modules/tippy.ts';
import {createElementFromHTML} from '../utils/dom.ts';
import {html} from '../utils/html.ts';

async function loadDeployMenu(url: string, content: HTMLElement) {
  content.classList.add('is-loading');
  try {
    const resp = await GET(url);
    content.replaceChildren(resp.ok ? createElementFromHTML(await resp.text()) : createElementFromHTML(html`<div class="repo-deploy-empty">${resp.statusText}</div>`));
  } finally {
    content.classList.remove('is-loading');
  }
}

export function initRepoDeployMenu() {
  registerGlobalInitFunc('initRepoDeployMenu', (button: HTMLElement) => {
    const content = createElementFromHTML<HTMLElement>(html`<div class="repo-deploy-menu-content"></div>`);
    createTippy(button, {
      content,
      role: 'dialog',
      theme: 'default',
      trigger: 'click',
      interactive: true,
      hideOnClick: true,
      placement: 'bottom-end',
      // reloaded on every open, so the status of a running deployment stays current
      onShow: () => {
        loadDeployMenu(button.getAttribute('data-menu-url')!, content);
      },
    });
  });
}
