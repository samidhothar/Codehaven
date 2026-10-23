import {registerGlobalInitFunc} from '../modules/observer.ts';
import {createTippy} from '../modules/tippy.ts';

export function initRepoPullCheckoutButton() {
  registerGlobalInitFunc('initPullCheckoutButton', (button: HTMLElement) => {
    const content = button.nextElementSibling as HTMLElement;
    content.classList.remove('tw-hidden');
    createTippy(button, {
      content,
      role: 'dialog',
      theme: 'default',
      trigger: 'click',
      interactive: true,
      hideOnClick: true,
      placement: 'bottom-end',
    });
  });
}
