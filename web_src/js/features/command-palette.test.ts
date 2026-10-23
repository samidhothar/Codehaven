import {collectPageLinks, filterCommandItems} from './command-palette.ts';

test('collectPageLinks', () => {
  const container = document.createElement('div');
  container.innerHTML = `
    <nav id="navbar">
      <a class="item" href="/issues">Issues</a>
      <a class="item" href="/explore/repos">Explore</a>
      <div class="menu"><a class="item" href="/user/settings"><svg></svg> Settings</a></div>
      <a class="item" href="/user/logout">Sign Out</a>
      <a class="item" href="/user/settings">Settings duplicate</a>
      <a class="item" href="#">Anchor</a>
    </nav>
    <div class="secondary-nav">
      <overflow-menu><a class="item" href="/o/r/issues">Issues <span class="ui label">3</span></a></overflow-menu>
    </div>
  `;
  expect(collectPageLinks(container, 'o/r')).toEqual([
    {label: 'Issues', context: 'o/r', url: '/o/r/issues'},
    {label: 'Issues', context: '', url: '/issues'},
    {label: 'Explore', context: '', url: '/explore/repos'},
    {label: 'Settings', context: '', url: '/user/settings'},
  ]);
});

test('filterCommandItems', () => {
  const items = [
    {label: 'Explore', context: '', url: '/explore/repos'},
    {label: 'Issues', context: 'o/r', url: '/o/r/issues'},
    {label: 'Issues', context: '', url: '/issues'},
  ];
  expect(filterCommandItems(items, '')).toEqual(items);
  expect(filterCommandItems(items, 'xyz')).toEqual([]);
  expect(filterCommandItems(items, 'exp').map((i) => i.url)).toEqual(['/explore/repos']);
  expect(filterCommandItems(items, 'issues o/r')[0].url).toEqual('/o/r/issues');
  expect(filterCommandItems(items, 'o/r iss').map((i) => i.url)).toEqual(['/o/r/issues']);
  expect(filterCommandItems(items, 'iss xyz')).toEqual([]);

  const ranked = [
    {label: 'Switch theme: Auto', context: '', url: '/theme'},
    {label: 'maria', context: 'Maria Santos', url: '/maria'},
    {label: 'Pull Requests', context: '', url: '/pulls'},
  ];
  expect(filterCommandItems(ranked, 'ma').map((i) => i.url)).toEqual(['/maria']); // short terms need contiguous letters
  expect(filterCommandItems(ranked, 'pullreq').map((i) => i.url)).toEqual(['/pulls']); // longer terms may skip letters
  expect(filterCommandItems(ranked, 'theme').map((i) => i.url)).toEqual(['/theme']);
});
