import {createElementFromHTML} from '../utils/dom.ts';
import {html, htmlRaw} from '../utils/html.ts';
import {svg} from '../svg.ts';

export function initBackToTop() {
  const label = window.config.i18n.back_to_top;
  const btn = createElementFromHTML<HTMLButtonElement>(html`<button type="button" class="back-to-top" aria-label="${label}" data-tooltip-content="${label}" data-tooltip-placement="left">${htmlRaw(svg('octicon-arrow-up'))}</button>`);
  document.body.append(btn);

  let scheduled = false;
  const update = () => {
    scheduled = false;
    btn.classList.toggle('visible', window.scrollY > window.innerHeight * 1.5);
  };
  window.addEventListener('scroll', () => {
    if (scheduled) return;
    scheduled = true;
    requestAnimationFrame(update);
  }, {passive: true});
  update();

  btn.addEventListener('click', () => {
    const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    window.scrollTo({top: 0, behavior: reduceMotion ? 'auto' : 'smooth'});
    btn.blur();
  });
}
