import {createApp} from 'vue';

export async function initTrendsCharts() {
  const el = document.querySelector<HTMLElement>('#trends-charts');
  if (!el) return;
  const {default: TrendsCharts} = await import('../components/TrendsCharts.vue');
  const attr = (name: string) => el.getAttribute(`data-locale-${name}`)!;
  createApp(TrendsCharts, {
    weeks: JSON.parse(el.getAttribute('data-weeks')!),
    locale: {
      issues: attr('issues'),
      pulls: attr('pulls'),
      activity: attr('activity'),
      issuesOpened: attr('issues-opened'),
      issuesClosed: attr('issues-closed'),
      pullsOpened: attr('pulls-opened'),
      pullsMerged: attr('pulls-merged'),
      pushes: attr('pushes'),
      contributors: attr('contributors'),
    },
  }).mount(el);
}
