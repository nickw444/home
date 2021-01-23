"""
Generate a Home Assistant config from the sprinkle GPIO configuration.
"""
import datetime
import sys

from collections import defaultdict
from ruamel.yaml import YAML, scalarstring

yaml = YAML()


def main():
    file = sys.argv[1]
    f = open(file)
    data = yaml.load(f)

    topic_prefix = data['topic_prefix']

    availability_topic = topic_prefix + '/' + data['status_topic']

    tmpl_switches = {}
    switches = []
    customize = {}

    for entry in data['outputs']:
        base_config = {
            'platform': 'mqtt',
            'availability_topic': availability_topic,
        }
        base_topic = '{}/{}'.format(
            topic_prefix,
            entry['id']
        )
        base_name = 'sprinkle_' + entry['name']

        duration = entry.get('duration')
        icon = entry['icon']

        if duration is None:
            switches.append({
                **base_config,
                'name': base_name,
                'icon': icon,
                'command_topic': base_topic + '/set',
                'state_topic': base_topic,
                'payload_on': scalarstring.SingleQuotedScalarString('ON'),
                'payload_off': scalarstring.SingleQuotedScalarString('OFF'),
                'qos': 1
            })
        else:
            switches.append({
                **base_config,
                'name': base_name + '_power',
                'command_topic': base_topic + '/set',
                'state_topic': base_topic,
                'payload_on': scalarstring.SingleQuotedScalarString('ON'),
                'payload_off': scalarstring.SingleQuotedScalarString('OFF'),
                'qos': 1
            })

            switches.append({
                **base_config,
                'name': base_name + '_trigger',
                'command_topic': base_topic + '/set',
                'payload_on': 'ON:{}'.format(duration),
                'qos': 1
            })

            power_entity_id = 'switch.' + base_name + '_power'
            tmpl_switches[base_name] = {
                'value_template': "{{ is_state('" + power_entity_id + "', 'on') }}",
                'availability_template': "{{ not is_state('" + power_entity_id + "', 'unavailable') }}",
                'icon_template': icon,
                'turn_on': {
                    'service': 'switch.turn_on',
                    'data': {
                        'entity_id': 'switch.' + base_name + '_trigger'
                    }
                },
                'turn_off': {
                    'service': 'switch.turn_off',
                    'data': {
                        'entity_id': 'switch.' + base_name + '_power'
                    }
                }
            }

            customize['switch.{}_power'.format(base_name)] = {'hidden': True}
            customize['switch.{}_trigger'.format(base_name)] = {'hidden': True}

        customize['switch.{}'.format(base_name)] = {
            'friendly_name': ' '.join([x.capitalize() for x in entry['name'].split('_')])
        }

    switches.append({
        'platform': 'template',
        'switches': tmpl_switches,
    })

    yaml.dump({
        'homeassistant': {
            'customize': customize
        },
        'switch': switches
    }, sys.stdout)

main()
