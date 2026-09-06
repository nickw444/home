# Agent Instructions

- This is a LIVE Home Assistant configuration directory. A file change can affect the running instance.
- Avoid destructive actions unless the user requests them clearly.

## Configuration Changes

- Check the syntax of changed YAML before you reload or restart Home Assistant.
- Tell the user if a change needs a reload or restart.
- Do not change Home Assistant databases or registry data directly unless the user specifically requests the exact change.

## Internal Data

- Use the Home Assistant MCP for history, triggers, and entity or device registry information.
- Inspect `.storage` only when the MCP cannot provide the required information or the user specifically requests direct inspection.
- Do not write, move, rename, or delete files in `.storage` unless the user specifically requests the exact action.
- Treat the `.storage` file format as unstable.
- Never display `secrets.yaml`, passwords, API keys, tokens, or private credentials.
