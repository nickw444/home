import glob
import json
import os
import shutil
import subprocess
import tempfile
from hashlib import sha1
from os.path import basename, splitext
from typing import List, NamedTuple, Optional, Dict, Literal, Iterable
from urllib.parse import urlparse

import requests
from ruamel.yaml import YAML

yaml = YAML()

import click


@click.group()
@click.option('--config-dir', type=click.Path(exists=True), default='./')
def cli(config_dir: str):
    pass


@cli.command(help="Add a new dependency")
@click.pass_context
@click.argument("dependency")
def add(ctx, dependency):
    pass


class Dependency(NamedTuple):
    source: str
    root_is_custom_components: bool
    include: Optional[List[str]]

    def get_name(self):
        name,_ = splitext(basename(urlparse(self.source).path))
        return name

    def get_src_hash(self):
        return sha1(self.source.encode('utf8')).hexdigest()[:8]

    def is_github(self):
        return urlparse(self.source).hostname == 'github.com'

    def get_github_slug(self):
        parsed = urlparse(self.source)
        if parsed.hostname == 'github.com':
            return parsed.path[1:].rstrip('.git')


class LockedDependency(NamedTuple):
    source: str
    version: str
    is_release: bool
    type: Literal['core', 'lovelace']


def load_dependencies(path: str) -> List[Dependency]:
    dependencies = []
    with open(path) as f:
        for dep in yaml.load(f)['dependencies']:
            if isinstance(dep, str):
                dep = {'source': dep}
            dependencies.append(
                Dependency(
                    source=dep['source'],
                    root_is_custom_components=dep.get(
                        'root_is_custom_components', False),
                    include=dep.get('include')
                )
            )
    return dependencies


def load_locked_dependencies(path: str) -> Dict[str, LockedDependency]:
    dependencies = {}
    with open(path) as f:
        for source, data in yaml.load(f).items():
            dependencies[source] = LockedDependency(
                source=source,
                version=data['version'],
                is_release=data.get('is_release', False),
                type=data['type'],
            )
    return dependencies


def write_locked_dependencies(path: str, locked_dependencies: Dict[
    str, LockedDependency]):
    dumpable = {}
    for source, lock_info in locked_dependencies.items():
        dumpable[source] = {
            'version': lock_info.version,
            'type': lock_info.type,
        }

        if lock_info.is_release:
            dumpable[source]['is_release'] = True

    with open(path, 'w') as f:
        yaml.dump(dumpable, f)


def checkout_dependency_source(dependency: Dependency, ref: str = None):
    tmpdir = tempfile.TemporaryDirectory(suffix='-' + dependency.get_name())
    subprocess.run(
        ['git', 'clone', dependency.source, tmpdir.name],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL)
    if ref is not None:
        subprocess.run(
            ['git', 'checkout', ref],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            cwd=tmpdir.name)

    return tmpdir


def install_dependency(config_root_path: str, dependency: Dependency,
                       lock_info: LockedDependency):
    click.echo(
        click.style(f'Installing: {dependency.get_name()} ', fg='green'))

    if lock_info is not None and lock_info.type == 'lovelace' and lock_info.is_release:
        # Install directly from Github Releases, skip inference logic.
        return install_lovelace_release_dependency(
            config_root_path, dependency, tag_name=lock_info.version)
    else:
        version_ref = lock_info.version if lock_info else None
        with checkout_dependency_source(dependency,
                                        version_ref) as source_path:
            is_core = lock_info.type == 'core' if lock_info else is_core_dependency(
                dependency, source_path)
            if is_core:
                return install_core_dependency(config_root_path, dependency,
                                               source_path)
            else:
                return install_lovelace_dependency(
                    config_root_path, dependency, source_path)


def is_core_dependency(
        dependency: Dependency,
        cloned_path: str) -> bool:
    custom_components_path = os.path.join(cloned_path, 'custom_components')
    if dependency.root_is_custom_components:
        custom_components_path = cloned_path

    return os.path.exists(custom_components_path)


def install_core_dependency(
        config_root_path: str,
        dependency: Dependency,
        cloned_path: str):
    custom_components_path = os.path.join(cloned_path, 'custom_components')
    if dependency.root_is_custom_components:
        custom_components_path = cloned_path

    if not os.path.exists(custom_components_path):
        raise ArtifactNotFoundException()

    for component in os.listdir(custom_components_path):
        component_path = os.path.join(custom_components_path, component)
        if component.startswith('.') or not os.path.isdir(component_path):
            continue

        if dependency.include is not None and component not in dependency.include:
            continue

        destination_path = os.path.join(config_root_path, 'custom_components',
                                        component)
        if os.path.exists(destination_path):
            shutil.rmtree(destination_path)

        shutil.copytree(component_path, destination_path)
        click.echo(f"installed {component}")

    describe_result = subprocess.run(
        ['git', 'describe', '--always'],
        cwd=cloned_path, stdout=subprocess.PIPE)
    return LockedDependency(
        source=dependency.source,
        version=describe_result.stdout.decode('utf-8').strip(),
        is_release=False,
        type='core'
    )

class ArtifactNotFoundException(Exception):
    pass


def find_source_artifacts(dependency: Dependency, cloned_path: str,
                          filename_hint: str):
    artifacts = []
    for artifact in glob.glob(
            os.path.join(cloned_path, '**/*{}*'.format(filename_hint)),
            recursive=True):
        artifacts.append(artifact)

    return artifacts


def get_github_release(dependency: Dependency, tag_name: Optional[str]):
    github_slug = dependency.get_github_slug()
    if github_slug is None:
        # Cannot check Github if the dependency doesn't come from Github!
        return None

    url = f'https://api.github.com/repos/{github_slug}/releases/latest'
    if tag_name is not None:
        url = f'https://api.github.com/repos/{github_slug}/releases/tags/{tag_name}'

    resp = requests.get(url)
    resp.raise_for_status()
    return resp.json()


def find_github_releases_artifacts(dependency: Dependency, release_data):
    artifacts = []
    for asset in release_data['assets']:
        if asset['name'].endswith('.js') or asset['name'].endswith('.map'):
            artifacts.append(asset['browser_download_url'])

    return artifacts


def install_lovelace_release_dependency(
        config_root_path: str,
        dependency: Dependency,
        tag_name: Optional[str]
):
    release_data = get_github_release(dependency, tag_name=tag_name)
    github_artifacts = find_github_releases_artifacts(dependency, release_data)
    if len(github_artifacts):
        destination_path = os.path.join(
            config_root_path, 'www', dependency.get_name())
        if os.path.isdir(destination_path):
            shutil.rmtree(destination_path)
        os.mkdir(destination_path)

        for artifact in github_artifacts:
            artifact_basename = os.path.basename(urlparse(artifact).path)
            artifact_destination_path = os.path.join(destination_path,
                                                     artifact_basename)
            resp = requests.get(artifact)
            resp.raise_for_status()
            with open(artifact_destination_path, 'wb') as f:
                f.write(resp.content)
            print(f"installed {artifact_basename}")

        return LockedDependency(
            source=dependency.source,
            version=release_data['tag_name'],
            is_release=True,
            type='lovelace',
        )

    raise ArtifactNotFoundException()


def install_lovelace_dependency(
        config_root_path: str,
        dependency: Dependency,
        cloned_path: Optional[str],
) -> LockedDependency:
    hacs_json_path = os.path.join(cloned_path, 'hacs.json')
    if os.path.exists(hacs_json_path):
        hacs_config = json.load(open(hacs_json_path))
        asset_filename = hacs_config.get('filename') or hacs_config['name']

        source_artifacts = find_source_artifacts(dependency, cloned_path,
                                                 asset_filename)
        if len(source_artifacts):
            destination_path = os.path.join(
                config_root_path, 'www', dependency.get_name())
            if os.path.isdir(destination_path):
                shutil.rmtree(destination_path)
            os.mkdir(destination_path)

            for artifact in source_artifacts:
                artifact_basename = os.path.basename(artifact)
                artifact_destination_path = os.path.join(destination_path, artifact_basename)
                shutil.copy(artifact, artifact_destination_path)
                print(f"installed {artifact_basename}")

            describe_result = subprocess.run(
                ['git', 'describe', '--always'],
                cwd=cloned_path, stdout=subprocess.PIPE)
            return LockedDependency(
                source=dependency.source,
                version=describe_result.stdout.decode('utf-8').strip(),
                is_release=False,
                type='lovelace',
            )

    # No source candidates found, try check Github Releases.
    return install_lovelace_release_dependency(config_root_path, dependency,
                                               None)


@cli.command(help="Install dependencies from hass-deps.yaml")
@click.option('--force', help="Force reinstallation of all dependencies")
@click.pass_context
def install(ctx):
    config_dir = ctx.parent.params['config_dir']
    dependencies_path = os.path.join(config_dir, "hass-deps.yaml")
    dependencies_lock_path = os.path.join(config_dir, "hass-deps.lock")
    dependencies = load_dependencies(dependencies_path)
    locked_deps = {}
    if os.path.exists(dependencies_lock_path):
        locked_deps = load_locked_dependencies(dependencies_lock_path)

    for dependency in dependencies:
        lock_info = locked_deps.get(dependency.source)
        updated_lock_info = install_dependency(
            './demo-config', dependency, lock_info)

        if lock_info is None:
            # Only update locked dependency if no lock was previously specified
            locked_deps[dependency.source] = updated_lock_info

    write_locked_dependencies(dependencies_lock_path, locked_deps)


@cli.command(help="Upgrade dependencies to the latest version/release")
@click.argument('dependency', nargs=-1)
def upgrade(dependency: Iterable[str]):
    print(dependency)


if __name__ == '__main__':
    cli()
