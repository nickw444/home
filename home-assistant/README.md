# home-assistant

Configuration for Home Assistant instances living in this mono-repo.

| Directory |
|-----------|
| [`nk/`](nk/) |
| [`wh/`](wh/) |

## Live install (Home Assistant OS)

Home Assistant's config directory is `/config`. This repo is not the config
directory itself — it is cloned beside it and wired in:

1. **Clone** the mono-repo to `/config/.home`
   (`git@github.com:nickw444/home.git`).
2. **Gitfile** at `/config/.git` pointing at the real gitdir:
   ```
   gitdir: /config/.home/.git
   ```
   The worktree remains `/config/.home` (see `core.worktree` in that git
   config), so `git` from `/config` talks to the same repo.
3. **Symlinks** from `/config` into the instance tree, e.g.:
   ```
   /config/configuration.yaml → ./.home/home-assistant/nk/configuration.yaml
   /config/automations        → ./.home/home-assistant/nk/automations
   /config/packages           → /config/.home/home-assistant/nk/packages
   /config/esphome            → ./.home/home-assistant/nk/esphome
   …etc.
   ```

Tracked YAML/layout lives under `home-assistant/<instance>/`. Runtime state
stays on the `/config` volume as real directories/files (for example
`.storage/`, `secrets.yaml`, the recorder DB, and `custom_components/`
managed via `hass-deps`).

## Tools

[`tools/test_config.sh`](tools/test_config.sh) runs `hass --script check_config`
against an instance directory (used by CI).
