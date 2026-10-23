import {getUserCardName, renderUserCard} from './user-card.ts';

test('getUserCardName', () => {
  const container = document.createElement('div');
  container.innerHTML = `
    <a id="explicit" data-user-card="alice" href="/alice">Alice A</a>
    <div class="markup"><a id="mention" href="/bob">@bob</a><a id="team" href="/org/teams/t">@org/t</a></div>
    <a id="outside" href="/carol">@carol</a>
  `;
  const name = (id: string) => getUserCardName(container.querySelector<HTMLAnchorElement>(`#${id}`)!);
  expect(name('explicit')).toBe('alice');
  expect(name('mention')).toBe('bob');
  expect(name('team')).toBe('');
  expect(name('outside')).toBe('');
});

test('renderUserCard', () => {
  const card = renderUserCard({
    login: 'eve', full_name: '<b>Eve</b>', avatar_url: '/avatar', html_url: '/eve',
    description: '', location: 'Riyadh', followers_count: 3, following_count: 1,
  }, {followers: 'Followers', following: 'Following'});
  expect(card).toContain('&lt;b&gt;Eve&lt;/b&gt;');
  expect(card).not.toContain('user-card-description');
  expect(card).toContain('Riyadh');
  expect(card).toContain('<strong>3</strong> Followers');
});
