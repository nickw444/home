"""
Support for Ness D8X/D16X devices.

For more details about this component, please refer to the documentation at
https://home-assistant.io/components/ness_alarm/
"""
import asyncio
import logging
from collections import namedtuple

import voluptuous as vol

from homeassistant.components.binary_sensor import DEVICE_CLASSES
from homeassistant.const import EVENT_HOMEASSISTANT_STOP
from homeassistant.helpers import config_validation as cv
from homeassistant.helpers.discovery import async_load_platform
from homeassistant.helpers.dispatcher import async_dispatcher_send

REQUIREMENTS = ['nessclient==0.9.2']

_LOGGER = logging.getLogger(__name__)

DOMAIN = 'ness_alarm'
DATA_NESS = 'ness_alarm'

CONF_DEVICE_HOST = 'host'
CONF_DEVICE_PORT = 'port'
CONF_ZONES = 'zones'
CONF_ZONE_NAME = 'name'
CONF_ZONE_TYPE = 'type'
CONF_ZONE_ID = 'id'
CONF_CODE = 'code'
CONF_OUTPUT_ID = 'output_id'
CONF_STATE = 'state'

SIGNAL_ZONE_CHANGED = 'ness_alarm.zone_changed'
SIGNAL_ARMING_STATE_CHANGED = 'ness_alarm.arming_state_changed'

ZoneChangedData = namedtuple('ZoneChangedData', ['zone_id', 'state'])

DEFAULT_ZONE_TYPE = 'motion'
ZONE_SCHEMA = vol.Schema({
    vol.Required(CONF_ZONE_NAME): cv.string,
    vol.Required(CONF_ZONE_ID): cv.positive_int,
    vol.Optional(CONF_ZONE_TYPE, default=DEFAULT_ZONE_TYPE):
        vol.In(DEVICE_CLASSES)})

CONFIG_SCHEMA = vol.Schema({
    DOMAIN: vol.Schema({
        vol.Required(CONF_DEVICE_HOST): cv.string,
        vol.Required(CONF_DEVICE_PORT): cv.port,
        vol.Optional(CONF_ZONES): vol.All(cv.ensure_list, [ZONE_SCHEMA]),
    }),
}, extra=vol.ALLOW_EXTRA)

SERVICE_PANIC = 'panic'
SERVICE_AUX = 'aux'

SERVICE_SCHEMA_PANIC = vol.Schema({
    vol.Required(CONF_CODE): cv.string,
})
SERVICE_SCHEMA_AUX = vol.Schema({
    vol.Required(CONF_OUTPUT_ID): cv.positive_int,
    vol.Optional(CONF_STATE, default=True): cv.boolean,
})


async def async_setup(hass, config):
    """Set up the Ness Alarm platform."""
    from nessclient import Client, ArmingState
    conf = config[DOMAIN]

    zones = conf.get(CONF_ZONES)
    host = conf.get(CONF_DEVICE_HOST)
    port = conf.get(CONF_DEVICE_PORT)

    client = Client(host=host, port=port, loop=hass.loop)
    hass.data[DATA_NESS] = client

    async def _close():
        client.close()

    hass.bus.async_listen_once(EVENT_HOMEASSISTANT_STOP, _close())

    task_zones = hass.async_create_task(
        async_load_platform(
            hass, 'binary_sensor', DOMAIN, {CONF_ZONES: zones}, config))
    task_control_panel = hass.async_create_task(
        async_load_platform(
            hass, 'alarm_control_panel', DOMAIN, conf, config))

    await asyncio.wait([task_zones, task_control_panel])

    def on_zone_change(zone_id: int, state: bool):
        """Receives and propagates zone state updates."""
        async_dispatcher_send(hass, SIGNAL_ZONE_CHANGED, ZoneChangedData(
            zone_id=zone_id,
            state=state,
        ))

    def on_state_change(arming_state: ArmingState):
        """Receives and propagates arming state updates."""
        async_dispatcher_send(hass, SIGNAL_ARMING_STATE_CHANGED, arming_state)

    client.on_zone_change(on_zone_change)
    client.on_state_change(on_state_change)

    # Force update for current arming status and current zone states
    hass.loop.create_task(client.keepalive())
    hass.loop.create_task(client.update())

    async def handle_panic(call):
        await client.panic(call.data[CONF_CODE])

    async def handle_aux(call):
        await client.aux(call.data[CONF_OUTPUT_ID], call.data[CONF_STATE])

    hass.services.async_register(DOMAIN, 'panic', handle_panic,
                                 schema=SERVICE_SCHEMA_PANIC)
    hass.services.async_register(DOMAIN, 'aux', handle_aux,
                                 schema=SERVICE_SCHEMA_AUX)

    return True
