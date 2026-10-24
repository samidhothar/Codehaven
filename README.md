# CodeHaven

A modern, keyboard-first take on [Gitea](https://github.com/go-gitea/gitea): self-hosted Git hosting, code review, issues, packages and CI/CD with a command palette, live search, hover cards, auto-saved drafts and smooth page transitions.

CodeHaven is a fork of Gitea (MIT licensed). Everything Gitea does still works the same way; CodeHaven adds interface improvements on top and ships a ready-to-use theme kit you can rebrand in minutes.

![CodeHaven home page](contrib/codehaven/screenshots/home.jpg)

## What's new compared to Gitea

| Feature | What you get |
|---|---|
| **Command palette** | Press <kbd>Ctrl</kbd>+<kbd>K</kbd> (<kbd>⌘</kbd>+<kbd>K</kbd> on macOS) anywhere, or click the search icon in the navbar. Fuzzy-search pages, people and (when signed in) repositories; type `#123` in a repository or `owner/repo#123` anywhere to jump straight to an issue or pull request. |
| **Palette actions** | Create an issue or pull request in the current repository, copy the page link, or switch theme, without leaving the keyboard. Only actions you are allowed to use are listed. |
| **Keyboard list navigation** | On issue and pull request lists, <kbd>j</kbd>/<kbd>k</kbd> move between items, <kbd>o</kbd> opens one and <kbd>x</kbd> selects it for bulk actions. Press <kbd>?</kbd> to see every shortcut on the page. |
| **Auto-saved comment drafts** | Unsent comments and new issue or pull request descriptions survive reloads, closed tabs and crashes. Drafts stay in your browser only and are removed when you submit, clear the text, or after 7 days. |
| **User hover cards** | Hover an author, avatar or `@mention` to see the person's avatar, name, bio, location and follower counts. |
| **Page transitions** | Pages cross-fade while the navbar stays in place. Turned off automatically when your system asks for reduced motion. |
| **Jenkins deployments to a VM** | Pick an environment, a branch, when to deploy and how from dropdowns in the repository settings; CodeHaven creates and runs the Jenkins pipeline, deploys over SSH, and shows each environment's status in a **Deploy** menu and next to the commit. No pipeline file to write. |
| **Faster issue and pull request pages** | A bar with the status, comment navigation, Close and Comment stays on screen while you scroll; select text in a comment to quote it; long comments fold behind **Show more**; a **Checkout** button gives the commands to try a pull request locally; **Next unviewed file** jumps through a review. |
| **Smoother buttons and code view** | Buttons respond to presses, save buttons show a spinner and ignore double clicks, long pages get a back-to-top button, and code files have a **Wrap** toggle that is remembered. |
| **Issue search, slash commands and saved replies** | <kbd>Ctrl</kbd>+<kbd>K</kbd> also finds issues and pull requests by title; type `/` in any comment for tables, task lists, diagrams and your saved replies. |
| **CodeHaven theme kit** | Accessible light, dark and auto themes (WCAG AA contrast), a redesigned home page and a Docker Compose setup, all applied through Gitea's custom folder without touching the core code. |

## Screenshots

**Command palette**: actions for the current repository, its pages, and live search, all from <kbd>Ctrl</kbd>+<kbd>K</kbd>.

![Command palette](contrib/codehaven/screenshots/command-palette.jpg)

**Keyboard navigation**: <kbd>j</kbd>/<kbd>k</kbd> highlight and focus issues; <kbd>o</kbd> opens and <kbd>x</kbd> selects.

![Issue list with keyboard navigation](contrib/codehaven/screenshots/keyboard-navigation.jpg)

**User hover cards**: profile details on any author link, avatar or `@mention`.

![User hover card on a mention](contrib/codehaven/screenshots/hover-card.jpg)

**CodeHaven Dark**: every page in the accessible dark theme.

![Issue page in the CodeHaven Dark theme](contrib/codehaven/screenshots/dark-theme.jpg)

## Quick start with Docker

Requirements: Docker with Docker Compose.

```sh
git clone https://github.com/samidhothar/Codehaven.git
cd Codehaven/contrib/codehaven
cp .env.example .env    # then set JENKINS_ADMIN_PASSWORD in .env
docker compose up -d --build
```

The first build compiles the frontend and backend from source and takes several minutes. Then open <http://localhost:3000>, finish the installation page, and create your admin account.

| Service | Address |
|---|---|
| Web interface | <http://localhost:3000> |
| Git over SSH | `ssh://git@localhost:2222/<owner>/<repo>.git` |

Data is kept in the `codehaven-data` Docker volume, so it survives restarts and rebuilds. To update after pulling new code, run `docker compose up -d --build` again.

If port 3000 or 2222 is already in use, change the left-hand side of the `ports` entries in [`contrib/codehaven/docker-compose.yml`](contrib/codehaven/docker-compose.yml) and update `GITEA__server__ROOT_URL` and `GITEA__server__SSH_PORT` to match.

## Deploy to a VM with Jenkins

The Docker setup includes a pre-configured Jenkins and a small test VM, connected automatically:

1. Copy `contrib/codehaven/.env.example` to `.env` and set `JENKINS_ADMIN_PASSWORD`.
2. Run `docker compose up -d --build`. On first start an SSH key pair is generated; Jenkins gets the private key as the credential `codehaven-vm-ssh` and the test VM trusts the public key.
3. In a repository, open **Settings → Deployments** and add an environment: choose Development, Staging or Production, the branch, **On every push** or **Manually**, and the deployment type (static files, Docker Compose or a custom command). For the bundled test VM use host `test-vm`, user `deploy` and a folder such as `/srv/app`.
4. Push to the branch, or use the **Deploy** menu in the repository header. The deployed files of the test VM are served on <http://localhost:8081>.

To deploy to a real machine, add the public key to `~/.ssh/authorized_keys` of the SSH user there (print it with `docker compose exec test-vm cat /run/codehaven-keys/id_ed25519.pub`) and change the host and user of the environment. Jenkins clones over HTTP from CodeHaven, so the repository must be readable without signing in, for example public.

| Service | Address |
|---|---|
| Jenkins | <http://localhost:8080> (sign in with the admin account from `.env`) |
| Test VM web server | <http://localhost:8081> |

## Dependency updates with Renovate

[Renovate](https://docs.renovatebot.com) watches your repositories and opens pull requests when a dependency (npm, Go modules, Docker images, pip, Maven and many more) has a new version. It is included in the Docker setup but only starts when you ask for it:

1. In CodeHaven, register a user for the bot, for example `renovate-bot`.
2. Signed in as that user, open **Settings → Applications** and generate a token with **write** access to *repository* and *issue*, and **read** access to *user*, *organization* and *misc*.
3. Put the token in `.env` as `RENOVATE_TOKEN`.
4. Add `renovate-bot` as a collaborator with **Write** access to each repository it should update (or to an organization team).
5. Run `docker compose --profile renovate up -d`.

On its first run Renovate opens a **Configure Renovate** pull request in each repository; merge it to start receiving update pull requests. Renovate checks again every hour (`RENOVATE_INTERVAL_SECONDS`).

## Security scanning

Each repository has a **Security** tab, next to **Actions**, that lists what [Trivy](https://trivy.dev) found on the default branch: known vulnerabilities in dependencies, committed secrets, and misconfigurations in Dockerfiles, Kubernetes manifests and Terraform. Findings can be filtered by severity and type, and link to the affected file and line.

The scan runs as an Actions workflow, so the Docker setup includes a runner that starts only when you ask for it:

1. In `.env`, set `CODEHAVEN_ROOT_URL` to an address that containers can reach, for example `http://host.docker.internal:3000/` on Docker Desktop. Workflow jobs clone and upload results through it, which does not work with `localhost`.
2. In **Site Administration → Actions → Runners**, click **Create new Runner** and put the token in `.env` as `RUNNER_REGISTRATION_TOKEN`.
3. Run `docker compose --profile actions up -d`.
4. Open a repository's **Security** tab and commit the workflow it shows as `.gitea/workflows/security.yml`.

The tab shows the report of the newest successful run on the default branch that uploaded a `trivy-results` artifact. It is visible to everyone who can see the repository's Actions.

## Team and administration features

These come with CodeHaven and need at most a setting:

| Feature | Where to find it |
|---|---|
| Cross-repository code search | **Explore → Code**. Turned on in the Docker setup with `GITEA__indexer__REPO_INDEXER_ENABLED`. |
| Audit log | **Site Administration → Audit Log** for the whole server, **Settings → Audit Log** in a repository or organization. |
| Time tracking | **Start timer** or **Add time** in the issue sidebar; totals appear on the issue and in milestones. |
| Disable an account but keep its history | **Site Administration → User Accounts → Edit → Disable sign-in**. Commits, issues and comments stay in place. |
| Microsoft Teams | **Settings → Webhooks → Add webhook → Microsoft Teams** in a repository or organization. |
| Email through SendGrid | Set the `MAILER_*` values in `.env` (the example there is for SendGrid), then `docker compose up -d`. |
| Separate storage for Git LFS, packages and artifacts | Give each its own location with the `[lfs]`, `[packages]` and `[actions.artifacts]` sections in `app.ini` (`STORAGE_TYPE` plus a path or S3-compatible bucket). See the [storage documentation](https://docs.gitea.com/administration/config-cheat-sheet#storage-storage). |
| High availability | Run several CodeHaven instances behind a load balancer with an external PostgreSQL or MySQL database, Redis for cache, sessions and queues, and shared or S3-compatible storage. The [Gitea Helm chart](https://gitea.com/gitea/helm-gitea) sets this up on Kubernetes. |

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
