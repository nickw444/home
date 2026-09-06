# Packages

Home Assistant loads every YAML file in this tree recursively through:

```yaml
homeassistant:
  packages: !include_dir_named packages
```

The filename, without `.yaml`, is the package name. Filenames must therefore
be unique across the entire tree, including across different subdirectories.

Packages are organized primarily by area:

- Area directories, such as `kitchen/` or `laundry/`: configuration and
  workflows for that area. New packages should go into an area wherever
  possible.
- `home/`: a catch-all for whole-home or cross-area behavior.
- `system/`: a catch-all for Home Assistant-wide loaders and platform
  configuration.
- `legacy/`: existing packages awaiting migration. Do not add new packages
  here.

A package may contain multiple integration domains so related helpers,
templates, scripts, and automations can live together.
