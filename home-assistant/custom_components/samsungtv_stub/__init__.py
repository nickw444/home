"""
A stub custom component to force Home Assistant to load a fork of samsungctl that is
compatible with our samsung RU8000 series TV.
"""

REQUIREMENTS = ['https://github.com/eclair4151/samsungctl/archive/websocketssl.zip#samsungctl==0.7.1+1']

def setup(hass, config):
    return True
