const resetDelayMs = 15 * 1000; // a submit that answers with a download never leaves the page

function resetSubmitting(form: HTMLFormElement) {
  form.removeAttribute('data-submitting');
  for (const btn of form.querySelectorAll('.submit-loading')) {
    btn.classList.remove('submit-loading', 'is-loading');
    btn.removeAttribute('aria-disabled');
  }
}

// marks the clicked button as loading for full-page form submits and swallows repeated submits until the page changes
export function initGlobalSubmitLoading() {
  window.addEventListener('submit', (e: SubmitEvent) => {
    const form = e.target as HTMLFormElement;
    if (form.hasAttribute('data-submitting')) {
      e.preventDefault();
      return;
    }
    // prevented submits are handled by scripts (fetch actions, inline editors) which show their own state
    // synthetic submit events (dispatched by scripts) never navigate
    if (!e.isTrusted || e.defaultPrevented || (form.target && form.target !== '_self')) return;

    form.setAttribute('data-submitting', '');
    // disabling the submitter would drop its name and value from the submitted data
    const btn = e.submitter ?? form.querySelector('[type=submit]');
    if (btn?.matches('.ui.button, .btn')) {
      btn.classList.add('submit-loading', 'is-loading');
      btn.setAttribute('aria-disabled', 'true');
    }
    setTimeout(() => resetSubmitting(form), resetDelayMs);
  });

  window.addEventListener('pageshow', (e) => {
    if (!e.persisted) return;
    for (const form of document.querySelectorAll<HTMLFormElement>('form[data-submitting]')) resetSubmitting(form);
  });
}
