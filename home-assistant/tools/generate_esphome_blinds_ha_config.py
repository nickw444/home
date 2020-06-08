"""
Tool to generate Home Assistant config for RAEX blind transmission via
a custom ESPHome component. See home-assistant/313a/esphome/raex_blind_tx.h
"""
import re

import click
from ruamel.yaml import YAML
from ruamel.yaml.scalarstring import PreservedScalarString

yaml = YAML()

CUSTOMIZE_BASE = {
    'assumed_state': True,
}


def get_service_call(tx_device: str, remote_id: int, channel_id: int, action: str):
    return {
        'service': f'esphome.{tx_device}_transmit',
        'data': {
            'remote_id': remote_id,
            'channel_id': channel_id,
            'action': action,
        }
    }


@click.command(help='Generate Home Assistant config for RAEX blind '
                    'transmission via ESPHome custom component')
@click.option('--input', required=True, type=click.File('r'),
              help='Path to the input config file')
@click.option('--output', required=True, type=click.File('w'),
              help='Destination path for the generated Home Assistant package')
@click.option('--support-pairing/--no-support-pairing', default=True,
              help='Whether switches should be generated to support pairing '
                   'of blinds')
def main(input, output, support_pairing):
    seed_config = yaml.load(input)
    tx_device = seed_config['tx_device']

    switches = {}
    covers = {}
    customize = {}

    availability_template = PreservedScalarString(
        f"{{{{ is_state('binary_sensor.{tx_device}_status', 'on') }}}}"
    )

    for blind in seed_config['blinds']:
        svc_call = lambda action: get_service_call(
            tx_device=tx_device,
            remote_id=blind['remote'],
            channel_id=blind['channel'],
            action=action,
        )

        cover = {
            'friendly_name': blind['name'],
            'device_class': 'blind',
            'open_cover': svc_call("OPEN"),
            'close_cover': svc_call("CLOSE"),
            'stop_cover': svc_call("STOP"),
            'availability_template':  availability_template,
        }
        covers[camelize(blind['name'])] = cover
        customize['cover.' + camelize(blind['name'])] = CUSTOMIZE_BASE

        pairing_switch = {
            'friendly_name': blind['name'] + ' Blind Pairing',
            'value_template': 'off',
            'turn_on': svc_call("PAIR"),
            'turn_off': [],
            'availability_template':  availability_template,
        }
        switches[camelize(pairing_switch['friendly_name'])] = pairing_switch

    package = {
        'homeassistant': {
            'customize': customize,
        },
        'cover': [
            {
                'platform': 'template',
                'covers': covers,
            }
        ]
    }

    if support_pairing:
        package['switch'] = {
            'platform': 'template',
            'switches': switches,
        }

    yaml.dump(package, output)


def camelize(s):
    s = re.sub(r'[^A-Za-z0-9\s]', '', s).lower()
    return '_'.join(s.split(' '))


if __name__ == '__main__':
    main()
