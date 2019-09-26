import base64
import re

import click
from pulses import (
    BroadlinkEncoder, RemoteCode, BlindAction, build_preamble,
    PhaseDuration)
from ruamel.yaml import YAML

yaml = YAML()

# Number of time for the broadlink to repeat the transmission
BROADLINK_REPEATS = 7
# Number of repetitions of the remote payload within the broadlink payload
PAYLOAD_REPEATS = 4


def get_broadlink_send_action(host: str, packet: str):
    return {
        'service': 'broadlink.send',
        'data': {
            'host': host,
            'packet': packet
        }
    }


def encode_packet(encoder: BroadlinkEncoder, remote_code: RemoteCode):
    data = []
    data += build_preamble()

    for _ in range(PAYLOAD_REPEATS):
        data += remote_code.get_phase_durations()
        # Drop a little bit of padding between payloads within a transmission
        data += [PhaseDuration(not data[-1].phase, 5000)]

    packet = encoder.encode(data)
    return base64.b64encode(packet).decode('utf-8')


@click.command(help='Generate Home Assistant config for RAEX blind '
                    'transmission via Broadlink')
@click.option('--input', required=True, type=click.File('r'),
              help='Path to the input config file')
@click.option('--output', required=True, type=click.File('w'),
              help='Destination path for the generated Home Assistant package')
@click.option('--support-pairing/--no-support-pairing', default=False,
              help='Whether switches should be generated to support pairing '
                   'of blinds')
def main(input, output, support_pairing):
    seed_config = yaml.load(input)
    host = seed_config['host_device_ip']
    encoder = BroadlinkEncoder(repeats=BROADLINK_REPEATS)

    action = lambda packet: get_broadlink_send_action(host, packet)

    switches = {}
    covers = {}

    for blind in seed_config['blinds']:
        open_code = RemoteCode(channel=blind['channel'],
                               remote=blind['remote'], action=BlindAction.UP)
        close_code = RemoteCode(channel=blind['channel'],
                                remote=blind['remote'],
                                action=BlindAction.DOWN)
        stop_code = RemoteCode(channel=blind['channel'],
                               remote=blind['remote'], action=BlindAction.STOP)
        pair_code = RemoteCode(channel=blind['channel'],
                               remote=blind['remote'], action=BlindAction.PAIR)
        open_packet = encode_packet(encoder, open_code)
        close_packet = encode_packet(encoder, close_code)
        stop_packet = encode_packet(encoder, stop_code)
        pair_packet = encode_packet(encoder, pair_code)

        cover = {
            'friendly_name': blind['name'],
            'device_class': 'blind',
            'open_cover': action(open_packet),
            'close_cover': action(close_packet),
            'stop_cover': action(stop_packet),
        }
        covers[camelize(blind['name'])] = cover

        pairing_switch = {
            'friendly_name': blind['name'] + ' Blind Pairing',
            'turn_on': action(pair_packet),
            'turn_off': None,
        }
        switches[camelize(pairing_switch['friendly_name'])] = pairing_switch

    package = {
        'cover': [
            {
                'platform': 'template',
                'covers': covers,
            }
        ]
    }

    if support_pairing:
        package['switch'] = {
            'platforms': 'template',
            'switches': switches,
        }

    yaml.dump(package, output)


def camelize(s):
    s = re.sub(r'[^A-Za-z0-9\s]', '', s).lower()
    return '_'.join(s.split(' '))


if __name__ == '__main__':
    main()
