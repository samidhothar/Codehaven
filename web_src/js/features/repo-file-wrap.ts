import {registerGlobalInitFunc} from '../modules/observer.ts';
import {localUserSettings} from '../modules/user-settings.ts';

const settingKey = 'file-view-line-wrap';

export function initRepoFileWrapToggle() {
  registerGlobalInitFunc('initFileWrapToggle', (btn: HTMLElement) => {
    const view = btn.closest('.non-diff-file-content')?.querySelector('.file-view.code-view');
    if (!view) return;
    const apply = (wrap: boolean) => {
      view.classList.toggle('code-nowrap', !wrap);
      btn.classList.toggle('active', wrap);
      btn.setAttribute('aria-pressed', String(wrap));
    };
    apply(localUserSettings.getBoolean(settingKey, true));
    btn.classList.remove('tw-hidden');
    btn.addEventListener('click', () => {
      const wrap = view.classList.contains('code-nowrap');
      localUserSettings.setBoolean(settingKey, wrap);
      apply(wrap);
    });
  });
}
