"""
Generate a Home Assistant config from a blinds.yaml configuration.
"""

import sys
import re
from ruamel.yaml import YAML, scalarstring

yaml = YAML()

def main():
    file = sys.argv[1]
    f = open(file)
    data = yaml.load(f)

    blinds = []
    send_topic = data['send_topic']
    status_topic = data['status_topic']

    customize = {}

    for trans_name, trans_info in data['transmitters'].items():
        trans_base_topic = '/things/blindkit/{}/'.format(trans_info['mac'])
        trans_send_topic = trans_base_topic + send_topic
        trans_status_topic = trans_base_topic + status_topic

        for blind_info in trans_info['blinds']:
            name = camelize(blind_info['name'])
            payload_prefix = '{},{},'.format(
                blind_info['remote'],
                blind_info['channel']
            )

            payload_open = payload_prefix + 'OPEN'
            payload_close = payload_prefix + 'CLOSE'
            payload_stop = payload_prefix + 'STOP'

            blinds.append({
                'platform': 'mqtt',
                'name': name,
                'command_topic': trans_send_topic,
                'availability_topic': trans_status_topic,
                'qos': 1,
                'payload_open': payload_open,
                'payload_close': payload_close,
                'payload_stop': payload_stop,
                'set_position_topic': trans_send_topic,
                'set_position_template': scalarstring.PreservedScalarString(
                    '{% if position > 70 %}' + payload_open + '\n'
                    '{% elif position < 30 %}' + payload_close + '\n'
                    '{% else %}' + payload_stop + '\n{% endif %}'
                )
            })

            customize['cover.{}'.format(name)] = {
                'friendly_name': blind_info['name']
            }

    yaml.dump(blinds, sys.stdout)
    yaml.dump(customize, sys.stdout)


def camelize(s):
    s = re.sub(r'[^A-Za-z0-9\s]','', s).lower()
    return 'blindkit_' + '_'.join(s.split(' '))


main()
