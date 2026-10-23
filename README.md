# CodeHaven

A modern, keyboard-first take on [Gitea](https://github.com/go-gitea/gitea): self-hosted Git hosting, code review, issues, packages and CI/CD with a command palette, live search, hover cards, auto-saved drafts and smooth page transitions.

CodeHaven is a fork of Gitea (MIT licensed). Everything Gitea does still works the same way; CodeHaven adds interface improvements on top and ships a ready-to-use theme kit you can rebrand in minutes.

## What's new compared to Gitea

| Feature | What you get |
|---|---|
| **Command palette** | Press <kbd>Ctrl</kbd>+<kbd>K</kbd> (<kbd>⌘</kbd>+<kbd>K</kbd> on macOS) anywhere, or click the search icon in the navbar. Fuzzy-search pages, repositories and people; type `#123` in a repository or `owner/repo#123` anywhere to jump straight to an issue or pull request. |
| **Palette actions** | Create an issue or pull request in the current repository, copy the page link, or switch theme, without leaving the keyboard. Only actions you are allowed to use are listed. |
| **Keyboard list navigation** | On issue and pull request lists, <kbd>j</kbd>/<kbd>k</kbd> move between items, <kbd>o</kbd> opens one and <kbd>x</kbd> selects it for bulk actions. Press <kbd>?</kbd> to see every shortcut on the page. |
| **Auto-saved comment drafts** | Unsent comments and new issue or pull request descriptions survive reloads, closed tabs and crashes. Drafts stay in your browser only and are removed when you submit, clear the text, or after 7 days. |
| **User hover cards** | Hover an author, avatar or `@mention` to see the person's avatar, name, bio, location and follower counts. |
| **Page transitions** | Pages cross-fade while the navbar stays in place. Turned off automatically when your system asks for reduced motion. |
| **CodeHaven theme kit** | Accessible light, dark and auto themes (WCAG AA contrast), a redesigned home page and a Docker Compose setup, all applied through Gitea's custom folder without touching the core code. |

## Quick start with Docker

Requirements: Docker with Docker Compose.

```sh
git clone https://github.com/samidhothar/Codehaven.git
cd Codehaven/contrib/codehaven
docker compose up -d --build
```

The first build compiles the frontend and backend from source and takes several minutes. Then open <http://localhost:3000>, finish the installation page, and create your admin account.

| Service | Address |
|---|---|
| Web interface | <http://localhost:3000> |
| Git over SSH | `ssh://git@localhost:2222/<owner>/<repo>.git` |

Data is kept in the `codehaven-data` Docker volume, so it survives restarts and rebuilds. To update after pulling new code, run `docker compose up -d --build` again.

If port 3000 or 2222 is already in use, change the left-hand side of the `ports` entries in [`contrib/codehaven/docker-compose.yml`](contrib/codehaven/docker-compose.yml) and update `GITEA__server__ROOT_URL` and `GITEA__server__SSH_PORT` to match.

## Make it your own

The theme kit lives in [`contrib/codehaven/custom`](contrib/codehaven/custom) and is mounted into the container as Gitea's custom folder, so you can rebrand without rebuilding:

| To change | Edit |
|---|---|
| Site name | `GITEA____APP_NAME` in `docker-compose.yml` |
| Accent colour | the `--color-primary*` values in `custom/public/assets/css/theme-codehaven-light.css` and `theme-codehaven-dark.css` |
| Logo and favicon | add `logo.svg`, `logo.png`, `favicon.svg`, `favicon.png` and `apple-touch-icon.png` to `custom/public/assets/img/` |
| Home page | `custom/templates/home.tmpl` and `custom/public/assets/css/codehaven.css` |
| Default theme | `GITEA__ui__DEFAULT_THEME` in `docker-compose.yml` |

After editing, run `docker compose restart`. Browsers cache assets for a few hours, so use a hard reload (<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>R</kbd>) to see new logos.

Users who registered before you changed the default theme keep their own choice; they can switch from the theme menu in the footer or from the command palette.

## Development

CodeHaven builds and develops exactly like Gitea. You need Go and Node.js with pnpm; run `make help` to list every target. Useful ones:

```sh
make build            # build the gitea binary with embedded frontend assets
make watch            # rebuild and restart on changes
make lint-js lint-css lint-templates lint-go
pnpm exec vitest <path-filter>
```

See [`docs/`](docs) for the build, testing and contribution guidelines inherited from Gitea.

## Documentation

Configuration, administration and the API are unchanged from Gitea, so the [Gitea documentation](https://docs.gitea.com) applies.

## Credits and license

CodeHaven is built on [Gitea](https://about.gitea.com) by the Gitea authors, which is itself a fork of Gogs. It is released under the [MIT license](LICENSE), the same as Gitea. Gitea's name and logo belong to their respective owners, and this project is not affiliated with or endorsed by the Gitea project.
