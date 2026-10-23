import {builtinSlashCommands, matchSlashCommands, renderSlashSuggestions} from './SlashCommands.ts';

test('matchSlashCommands', () => {
  const commands = [...builtinSlashCommands({table: 'Table', checklist: 'Task list'}), {name: 'lgtm', label: 'Saved reply: LGTM', value: 'Looks good to me.'}];
  expect(matchSlashCommands(commands, 'ta').map((c) => c.name)).toEqual(['table', 'checklist', 'details']); // name prefix or label substring
  expect(matchSlashCommands(commands, 'saved').map((c) => c.name)).toEqual(['lgtm']);
  expect(matchSlashCommands(commands, 'xyz')).toEqual([]);
  expect(matchSlashCommands(commands, '')).toHaveLength(commands.length);
});

test('renderSlashSuggestions', () => {
  const ul = renderSlashSuggestions([{name: 'quote', label: '<b>Quote</b>', value: '> '}]);
  const li = ul.querySelector('li')!;
  expect(li.getAttribute('data-value')).toBe('> ');
  expect(li.hasAttribute('data-block')).toBe(false);
  expect(renderSlashSuggestions(builtinSlashCommands({})).querySelector('li')!.hasAttribute('data-block')).toBe(true);
  expect(li.textContent).toBe('/quote<b>Quote</b>'); // labels are text, never parsed as HTML
});
