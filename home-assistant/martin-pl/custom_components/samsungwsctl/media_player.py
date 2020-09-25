"""Support for interface with an Samsung TV via samsungwsctl client library."""
import logging
import time
from datetime import timedelta
from typing import Optional, NamedTuple, List

import requests
import voluptuous as vol

import homeassistant.helpers.config_validation as cv
from homeassistant.components.media_player import (
    MediaPlayerEntity,
    PLATFORM_SCHEMA, SUPPORT_PAUSE, SUPPORT_VOLUME_STEP, SUPPORT_VOLUME_MUTE,
    SUPPORT_PREVIOUS_TRACK, SUPPORT_NEXT_TRACK, SUPPORT_PLAY,
    SUPPORT_PLAY_MEDIA)
from homeassistant.components.media_player import (
    SUPPORT_SELECT_SOURCE, SUPPORT_TURN_OFF, SUPPORT_TURN_ON)
from homeassistant.components.media_player.const import (
    MEDIA_TYPE_CHANNEL,
    DOMAIN)
from homeassistant.const import (
    CONF_HOST, CONF_NAME, CONF_TIMEOUT,
    CONF_MAC, STATE_OFF, STATE_ON, CONF_PORT)
from homeassistant.util import Throttle

REQUIREMENTS = ['samsungwsctl=1.0.3']

_LOGGER = logging.getLogger(__name__)

CONF_SECURE = 'secure'
ATTR_KEY = 'key'
ATTR_VERSION = 'version'

DEFAULT_PORT = 8002
DEFAULT_TIMEOUT = 5

PLATFORM_SCHEMA = PLATFORM_SCHEMA.extend(
    {
        vol.Required(CONF_HOST): cv.string,
        vol.Required(CONF_PORT): cv.string,
        vol.Required(CONF_SECURE): cv.boolean,
        vol.Optional(CONF_NAME): cv.string,
        vol.Optional(CONF_MAC): cv.string,
        vol.Optional(CONF_TIMEOUT, default=DEFAULT_TIMEOUT): cv.positive_int,
    }
)

KEY_PRESS_TIMEOUT = 1.2


class Source(NamedTuple):
    name: str
    keys: Optional[List[str]]
    app_id: Optional[str]


SOURCE_TV = Source("TV", keys=["KEY_HOME", "KEY_EXIT", "KEY_EXIT"],
                   app_id=None)
SOURCES = {
    "TV": SOURCE_TV,
    "YouTube": Source("YouTube", app_id="111299001912", keys=None),
    "Netflix": Source("Netflix", app_id="11101200001", keys=None),
    "Spotify": Source("Spotify", app_id="3201606009684", keys=None),
}

SERVICE_SEND_KEY = "send_key"
SERVICE_SCHEMA_SEND_KEY = vol.Schema({
    vol.Required(ATTR_KEY): cv.string
})

SUPPORT_SAMSUNGTV = (
        SUPPORT_PAUSE
        | SUPPORT_VOLUME_STEP
        | SUPPORT_VOLUME_MUTE
        | SUPPORT_PREVIOUS_TRACK
        | SUPPORT_SELECT_SOURCE
        | SUPPORT_NEXT_TRACK
        | SUPPORT_TURN_OFF
        | SUPPORT_PLAY
        | SUPPORT_PLAY_MEDIA
)


def setup_platform(hass, config, add_entities, discovery_info=None):
    from samsungwsctl import Remote

    token_file = hass.config.path('.samsung_token')

    host = config[CONF_HOST]
    port = config[CONF_PORT]
    secure = config[CONF_SECURE]
    name = config.get(CONF_NAME)
    mac = config.get(CONF_MAC)
    timeout = config[CONF_TIMEOUT]

    remote = Remote(host=host, port=port, secure=secure, token_file=token_file,
                    remote_name='HomeAssistant', timeout=timeout)
    tv = SamsungTVDevice(name=name, mac=mac, remote=remote)

    # TODO(NW): Add support for multiple TVs per Home Assistant instance.
    #  Currently only the last added entity will be able to receive commands
    #  as we do not handle entity_id.
    hass.services.register(DOMAIN, SERVICE_SEND_KEY, tv.send_key,
                           schema=SERVICE_SCHEMA_SEND_KEY)

    add_entities([tv])


class SamsungTVDevice(MediaPlayerEntity):
    def __init__(self, name: Optional[str], mac: Optional[str], remote):
        from samsungwsctl import GetInfoResponse

        self._name = name
        self._mac = mac
        self._remote = remote

        self._info: Optional[GetInfoResponse] = None
        self._info_failed = None
        self._is_powering_off = False
        self._current_source: Optional[Source] = None

    @property
    def name(self) -> Optional[str]:
        """Return the name of the entity."""
        if self._name:
            return self._name

        if self._info is not None:
            return self._info.name

    @property
    def unique_id(self) -> Optional[str]:
        """Return a unique ID."""
        if self._info is not None:
            return self._info.id

    @property
    def state(self):
        if self._is_powering_off:
            return STATE_OFF

        if self._info is not None:
            if self._info.power_state == 'on':
                return STATE_ON
            elif self._info.power_state == 'standby':
                return STATE_OFF

        if self._info_failed:
            return STATE_OFF

    @property
    def assumed_state(self) -> bool:
        """Return True if unable to access real state of the entity."""
        return self._info_failed or self._is_powering_off

    @property
    def app_id(self):
        """ID of the current running app."""
        if self._current_source:
            return self._current_source.app_id

    @property
    def app_name(self):
        """Name of the current running app."""
        if self._current_source:
            return self._current_source.name

    @property
    def source(self):
        """Name of the current input source."""
        if self._current_source:
            return self._current_source.name

    @property
    def source_list(self):
        """List of available input sources."""
        return list(SOURCES.keys())

    @property
    def supported_features(self):
        """Flag media player features that are supported."""
        if self._mac:
            return SUPPORT_SAMSUNGTV | SUPPORT_TURN_ON
        return SUPPORT_SAMSUNGTV

    def update(self):
        _LOGGER.debug("Updating")
        is_online = self._update_info()
        if is_online:
            self._update_current_source()

    def _update_info(self) -> bool:
        try:
            self._info = self._remote.get_info()
            self._is_powering_off = False
            self._info_failed = False
            _LOGGER.debug("Fetched remote info: %s", self._info)
            return True

        except (requests.exceptions.ConnectionError,
                requests.exceptions.ConnectTimeout,
                requests.exceptions.ReadTimeout) as e:
            _LOGGER.debug(
                'Failed to fetch remote info. Assuming device is powered off')
            self._info_failed = True

    @Throttle(min_time=timedelta(minutes=1))
    def _update_current_source(self):
        for source in SOURCES.values():
            if source.app_id is not None:
                try:
                    app_info = self._remote.get_app_info(source.app_id)
                    if app_info.running and app_info.visible:
                        self._current_source = source
                        return
                except (requests.exceptions.ConnectionError,
                        requests.exceptions.ConnectTimeout,
                        requests.exceptions.ReadTimeout) as e:
                    # TV is probably offline
                    break

        # Current source couldn't be determined
        _LOGGER.debug("Current source could not be determined. Assuming TV")
        self._current_source = SOURCE_TV

    def turn_on(self):
        _LOGGER.debug("Sending WOL magic packet mac %s to wake device",
                      self._mac)
        import wakeonlan
        wakeonlan.send_magic_packet(self._mac)

    def turn_off(self):
        self._remote.send_key("KEY_POWER")
        self._remote.disconnect()
        self._is_powering_off = True

    def volume_up(self):
        """Volume up the media player."""
        self._remote.send_key("KEY_VOLUP")

    def volume_down(self):
        """Volume down media player."""
        self._remote.send_key("KEY_VOLDOWN")

    def mute_volume(self, mute):
        """Send mute command."""
        self._remote.send_key("KEY_MUTE")

    def media_play(self):
        """Send play command."""
        self._remote.send_key("KEY_PLAY")

    def media_pause(self):
        """Send media pause command to media player."""
        self._remote.send_key("KEY_PAUSE")

    def media_next_track(self):
        """Send next track command."""
        self._remote.send_key("KEY_FF")

    def media_previous_track(self):
        """Send the previous track command."""
        self._remote.send_key("KEY_REWIND")

    def play_media(self, media_type, media_id, **kwargs):
        """Support changing a channel."""
        if media_type != MEDIA_TYPE_CHANNEL:
            _LOGGER.error("Unsupported media type: %s", media_type)
            return

        # media_id should only be a channel number
        try:
            cv.positive_int(media_id)
        except vol.Invalid:
            _LOGGER.error("Media ID must be positive integer")
            return

        for digit in media_id:
            self._remote.send_key(f'KEY_{digit}')
            time.sleep(KEY_PRESS_TIMEOUT)

        self._remote.send_key(f'KEY_ENTER')

    def select_source(self, source):
        """Select input source."""
        selected_source = SOURCES.get(source)
        if selected_source is None:
            _LOGGER.error("Unknown source: %s", source)
            return

        if selected_source.app_id is not None:
            self._remote.start_app(selected_source.app_id)
        elif selected_source.keys is not None:
            for key in selected_source.keys:
                self._remote.send_key(key)
                time.sleep(KEY_PRESS_TIMEOUT)

        self._current_source = selected_source

    def send_key(self, call):
        self._remote.send_key(call.data[ATTR_KEY])
