# Agent Instructions

- This is a LIVE Home Assistant configuration directory. A file change can affect the running instance.
- Avoid destructive actions unless the user requests them clearly.

## Git Layout

- `/config` is not a Git work tree. The `/config/.git` pointer can make it look like one.
- Most tracked files in `/config` are symlinks to `/config/.home/home-assistant/nk/`. An edit to the symlink changes its tracked target.
- Git uses paths from `/config/.home`. For example, `/config/AGENTS.md` is `home-assistant/nk/AGENTS.md`. Do not assume that Git tracks all files in `/config`.

## Configuration Changes

- Use the `ha` CLI to check configuration before reloads. Restart Home Assistant only when explicitly requested.
- Tell the user if a change needs a reload or restart.
- Do not change Home Assistant databases or registry data directly unless the user specifically requests the exact change.

## Package Structure

- Put new YAML configuration in `/config/packages/`. Use the most relevant area folder, such as `/config/packages/kitchen/`, even if the configuration affects multiple areas.
- Use `home/` only for configuration that applies to the full home. Use `system/` for system configuration. Do not add files to `legacy/`.
- Keep related helpers and automations in one package. Each YAML filename must be unique in the full `packages/` tree. Subdirectories do not make filenames unique.

## Internal Data

- Use the Home Assistant MCP for history, triggers, and entity or device registry information.
- Inspect `.storage` only when the MCP cannot provide the required information or the user specifically requests direct inspection.
- Do not write, move, rename, or delete files in `.storage` unless the user specifically requests the exact action.
- Treat the `.storage` file format as unstable.
- Never display `secrets.yaml`, passwords, API keys, tokens, or private credentials.
