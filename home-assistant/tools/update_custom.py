from sys import path

from ruamel.yaml import YAML

yaml = YAML()

import click


@click.command
@click.argument('config-dir', type=click.Path(exists=True))
def cli(config_dir):
    manifest_path = path.join(config_dir, "custom_components.manifest.yaml")
    manifest = yaml.load(open(manifest_path))

    for

if __name__ == '__main__':
    cli()
