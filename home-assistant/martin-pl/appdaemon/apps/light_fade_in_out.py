from typing import Literal

import voluptuous as vol
from appdaemon.plugins.hass import hassapi as hass

from .config_validation import time_delta, parse_time

CONF_ENABLE_ENTITY = 'enable_entity'
CONF_PRESENCE_ENTITY = 'presence_entity'
CONF_START_TIME_ENTITY = 'start_time_entity'
CONF_FADE_IN_DURATION = 'fade_in_duration'
CONF_FADE_OUT_DURATION = 'fade_out_duration'
CONF_ENABLE_DAYS = 'enable_days'
CONF_LIGHTS = 'lights'
CONF_ENTITY_ID = 'entity_id'
CONF_COLOR_TEMP_KELVIN = 'color_temp_kelvin'

WEEKDAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thur', 'Fri', 'Sat', 'Sun']

CONFIG_SCHEMA = vol.Schema({
    vol.Required(CONF_ENABLE_ENTITY): str,
    vol.Optional(CONF_PRESENCE_ENTITY): str,
    vol.Required(CONF_START_TIME_ENTITY): str,
    vol.Required(CONF_FADE_IN_DURATION): time_delta,
    vol.Optional(CONF_FADE_OUT_DURATION): time_delta,
    vol.Optional(CONF_ENABLE_DAYS): [vol.Any(WEEKDAY_NAMES)],
    vol.Required(CONF_LIGHTS): [vol.Schema({
        CONF_ENTITY_ID: str,
        CONF_COLOR_TEMP_KELVIN: vol.Optional(str),
    })]
}, extra=True, required=True)


class LightFadeInOut(hass.Hass):
    def initialize(self):
        self._config = CONFIG_SCHEMA(self.args)

        if self._config[CONF_PRESENCE_ENTITY]:
            self.listen_state(self._on_presence_detected, self._config[CONF_PRESENCE_ENTITY],
                              new='home')

        self.listen_state(self._on_enable, self.args[CONF_ENABLE_ENTITY], new='on')
        self.listen_state(self._on_start_time_change, self.args[CONF_START_TIME_ENTITY])

    def _on_presence_detected(self, *args):
        self._resume()

    def _on_enable(self, *args):
        self._resume()

    def _on_start_time_change(self, entity, attribute, old ,new, **kwargs):
        if self._start_task:
            self.cancel_timer(self._start_task)

        self._start_task = self.run_daily(self._resume, new)

    def _on_minute(self):
        pass

    def _resume(self):
        curr_day_name = WEEKDAY_NAMES[self.date().weekday()]
        if self._config[CONF_ENABLE_DAYS] is not None and curr_day_name not in self._config[
            CONF_ENABLE_DAYS]:
            return

        curr_time = self.time()
        start_time_str = self.get_state(self._config[CONF_START_TIME_ENTITY])
        start_time = parse_time(start_time_str)

        fadein_end = start_time + self._config[CONF_FADE_IN_DURATION]
        fadeout_end = fadein_end + self._config[CONF_FADE_OUT_DURATION]

        if start_time < curr_time <= fadein_end:
            self._toggle('on', self._config[CONF_FADE_IN_DURATION].seconds())
        elif fadein_end < curr_time < fadeout_end:
            self._toggle('off', self._config[CONF_FADE_IN_DURATION].seconds())

    def _toggle(self, state: Literal['on', 'off'], transition_duration: int):
        impl = self.turn_on if state == 'on' else self.turn_off

        for light_conf in self._config[CONF_LIGHTS]:
            impl(light_conf[CONF_ENTITY_ID],
                 kelvin=light_conf[CONF_COLOR_TEMP_KELVIN],
                 transition=transition_duration)
