#!/usr/bin/env pipenv run python
import os
import urllib.request
import json
import requests
from requests.auth import HTTPBasicAuth

ESPURNA_ADMIN_PASSWORD = os.environ['ESPURNA_ADMIN_PASSWORD']

def find_latest_release():
    response = urllib.request.urlopen("https://api.github.com/repos/xoseperez/espurna/releases")
    data = json.load(response)
    return next((release for release in data if not release['prerelease'] and not release['draft']), None)


def find_devices(devices_file):
    with open(devices_file) as fh:
        devices_blob = json.load(fh)
        return [device for device in devices_blob if device['ota_method'] == 'ESPURNA']


def download_binary(name, url):
    if not os.path.exists('./firmware'):
        os.mkdir('./firmware')

    out_file_path = os.path.join('./firmware', name)
    if os.path.exists(out_file_path):
        return out_file_path

    resp = requests.get(url)
    if not resp.ok:
        raise Exception('Error when fetching binary: {}'.format(name))

    with open(out_file_path, 'wb') as out_file:
        out_file.write(resp.content)


def main():
    latest_release = find_latest_release()
    version = latest_release['tag_name']
    print('Found latest espurna release: {}'.format(version))

    # Pair devices to firmware binary downloads
    device_assets = []

    for device in find_devices('../esps.json'):
        espurna_model_slug = device['espurna_model_slug']
        firmware_filename = f'espurna-{version}-{espurna_model_slug}.bin'
        asset = next((asset for asset in latest_release['assets'] if asset['name'] == firmware_filename), None)
        device_assets.append((
            device,
            firmware_filename,
            download_binary(firmware_filename, asset['browser_download_url'])
        ))

    print('Flashing devices')
    for (device, firmware_filename, asset_path) in device_assets:
        print('Flashing: {} ({}) with {}'.format(device['name'], device['model'], firmware_filename))
        response = input('Flash this device? [Y/n]: ')
        if response.strip() not in ['Y', 'y', '', 'yes']:
            print('Skipping')
            continue

        upgrade_endpoint = 'http://{}/upgrade'.format(device['ip'])
        files = {
            'filename': open(asset_path, 'rb')
        }

        response = requests.post(
            upgrade_endpoint,
            files=files,
            auth=HTTPBasicAuth('admin', ESPURNA_ADMIN_PASSWORD)
        )

        print(response.text)

if __name__ == '__main__':
    main()
