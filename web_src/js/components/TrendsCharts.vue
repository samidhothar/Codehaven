<script lang="ts" setup>
import {Chart, Legend, LineElement, PointElement, BarElement, Tooltip, type ChartData, type ChartOptions} from 'chart.js';
import ChartCanvas from './ChartCanvas.vue';
import {chartJsColors} from '../utils/color.ts';

Chart.register(Legend, LineElement, PointElement, BarElement, Tooltip);

export type TrendsWeek = {
  start: number,
  issuesOpened: number,
  issuesClosed: number,
  pullsOpened: number,
  pullsMerged: number,
  pushes: number,
  contributors: number,
};

const props = defineProps<{
  weeks: TrendsWeek[],
  locale: Record<'issues' | 'pulls' | 'activity' | 'issuesOpened' | 'issuesClosed' | 'pullsOpened' | 'pullsMerged' | 'pushes' | 'contributors', string>,
}>();

const points = (key: keyof TrendsWeek) => props.weeks.map((w) => ({x: w.start * 1000, y: w[key]}));

const line = (label: string, key: keyof TrendsWeek, color: string) => ({
  label, data: points(key), borderColor: color, backgroundColor: color, pointRadius: 3, cubicInterpolationMode: 'monotone' as const,
});

const issuesData = {datasets: [
  line(props.locale.issuesOpened, 'issuesOpened', chartJsColors.additions),
  line(props.locale.issuesClosed, 'issuesClosed', chartJsColors.deletions),
]} as ChartData;

const pullsData = {datasets: [
  line(props.locale.pullsOpened, 'pullsOpened', chartJsColors.additions),
  line(props.locale.pullsMerged, 'pullsMerged', chartJsColors.commits),
]} as ChartData;

const activityData = {datasets: [
  {type: 'bar', label: props.locale.pushes, data: points('pushes'), backgroundColor: chartJsColors.commits},
  {...line(props.locale.contributors, 'contributors', chartJsColors.additions), type: 'line'},
]} as ChartData;

const options: ChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {mode: 'index', intersect: false},
  plugins: {legend: {display: true}},
  scales: {
    x: {type: 'time', time: {unit: 'week'}, grid: {display: false}, ticks: {maxRotation: 0, maxTicksLimit: 8}},
    y: {beginAtZero: true, ticks: {precision: 0, maxTicksLimit: 6}},
  },
};
</script>

<template>
  <div class="trends-charts">
    <div class="trends-chart">
      <h3>{{ locale.issues }}</h3>
      <div class="trends-chart-canvas"><ChartCanvas type="line" :data="issuesData" :options="options"/></div>
    </div>
    <div class="trends-chart">
      <h3>{{ locale.pulls }}</h3>
      <div class="trends-chart-canvas"><ChartCanvas type="line" :data="pullsData" :options="options"/></div>
    </div>
    <div class="trends-chart trends-chart-wide">
      <h3>{{ locale.activity }}</h3>
      <div class="trends-chart-canvas"><ChartCanvas type="bar" :data="activityData" :options="options"/></div>
    </div>
  </div>
</template>
