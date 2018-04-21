"""
Generate a Home Assistant config from the sprinkle GPIO configuration.
"""
import datetime
import sys

from collections import defaultdict
from ruamel.yaml import YAML, scalarstring

yaml = YAML()
config = defaultdict(lambda: (datetime.timedelta(minutes=10), 'mdi:water'))
config['tree_lights'] = (None, 'mdi:tree')

def main():
    file = sys.argv[1]
    f = open(file)
    data = yaml.load(f)

    topic_prefix = data['mqtt']['topic_prefix']

    availability_topic = topic_prefix + '/' + data['mqtt']['status_topic']

    tmpl_switches = {}
    switches = []
    customize = {}

    for entry in data['digital_outputs']:
        base_config = {
            'platform': 'mqtt',
            'availability_topic': availability_topic,
        }
        base_topic = topic_prefix + '/output/' + entry['name']
        base_name = 'sprinkle_' + entry['name']

        on_time, icon = config[entry['name']]

        if on_time is None:
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
            on_time_ms = round(on_time.total_seconds() * 1000)
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
                'command_topic': base_topic + '/set_on_ms',
                'payload_on': on_time_ms,
                'qos': 1
            })

            tmpl_switches[base_name] = {
                'value_template': '{{ states.switch.' + base_name + '_power.state }}',
                'icon': icon,
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

    yaml.dump(switches, sys.stdout)
    yaml.dump(customize, sys.stdout)

main()
