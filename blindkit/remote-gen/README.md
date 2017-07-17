## RAEX Blind Remote Code Generator

A tool to identify and generate new RAEX remote codes.

```
usage: RAEX Remote Code Generator Tool [<flags>] <command> [<args> ...]

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.


  info [<flags>] <code>
    Extract information about a captured remote code

    --validate  Verify that the checksum matches our guessed checksum value.

  create --channel=CHANNEL --remote=REMOTE [<flags>]
    Create new remote codes

    --channel=CHANNEL    Channel to broadcast on (uint8)
    --remote=REMOTE      Remote value to broadcast (uint16)
    --action=ACTION ...  Actions to create
    --verbose            Output additional information about the generated codes

```
