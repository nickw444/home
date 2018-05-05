#!/usr/bin/env pipenv run python

import requests
from ruamel.yaml import YAML

yaml = YAML()

def find_latest_release():
    resp = requests.get('https://api.github.com/repos/xoseperez/espurna/releases')
    if not resp.ok:
        raise Exception('Failed to fetch releases')

    data = resp.json()
    return next((release for release in data if not release['prerelease'] and not release['draft']), None)

def find_devices():


def main():
    find_devices()

    # latest_release = find_latest_release()
    # print('Found latest espurna release: {}'.format(latest_release['tag_name']))
    # print('Flashing devices...')

if __name__ == '__main__':
    main()





