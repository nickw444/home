"""Adds config flow for eventsensor."""
import logging

import voluptuous as vol

from homeassistant import config_entries
from homeassistant.const import CONF_EVENT, CONF_EVENT_DATA, CONF_NAME, CONF_STATE
from homeassistant.core import callback
from homeassistant.helpers.event import EVENT_STATE_CHANGED

from .common import (
    CONF_STATE_MAP,
    DOMAIN,
    PRESET_AQARA_CUBE,
    PRESET_AQARA_CUBE_MAPPING,
    PRESET_AQARA_SMART_BUTTON,
    PRESET_AQARA_SMART_BUTTON_MAPPING,
    PRESET_FOH,
    PRESET_FOH_MAPPING,
    PRESET_HUE_BUTTON,
    PRESET_HUE_BUTTON_MAPPING,
    PRESET_HUE_DIMMER,
    PRESET_HUE_DIMMER_MAPPING,
    PRESET_HUE_TAP,
    PRESET_HUE_TAP_MAPPING,
    make_string_ui_from_dict,
    make_unique_id,
    parse_dict_from_ui_string,
)

_LOGGER = logging.getLogger(__name__)

CONF_PRESET_CONFIG = "preset_config"
CONF_INTEGRATION = "integration_source"
CONF_IDENTIFIER = "identifier"
CONF_TYPE_IDENTIFIER = "type_identifier"

_EVENT_SOURCE_HUE = "Hue"
_EVENT_SOURCE_DECONZ = "deCONZ"
_EVENT_SOURCE_GENERIC = "Any other"
_IDENTIFIER_ID = "id"
_IDENTIFIER_UNIQUE_ID = "uniqueid"
_PRESET_GENERIC = "Custom state mapping"

STEP_1_INITIAL = vol.Schema(
    {
        vol.Required(CONF_NAME): str,
        vol.Required(CONF_INTEGRATION, default=_EVENT_SOURCE_HUE): vol.In(
            [_EVENT_SOURCE_HUE, _EVENT_SOURCE_DECONZ, _EVENT_SOURCE_GENERIC]
        ),
    },
)
STEP_2_PRECONFIGURED = vol.Schema(
    {
        vol.Required(CONF_TYPE_IDENTIFIER, default=_IDENTIFIER_ID): vol.In(
            [_IDENTIFIER_ID, _IDENTIFIER_UNIQUE_ID]
        ),
        vol.Optional(CONF_IDENTIFIER, default=""): str,
        vol.Required(CONF_PRESET_CONFIG, default=PRESET_HUE_DIMMER): vol.In(
            [
                PRESET_HUE_DIMMER,
                PRESET_HUE_TAP,
                PRESET_HUE_BUTTON,
                PRESET_FOH,
                PRESET_AQARA_SMART_BUTTON,
                PRESET_AQARA_CUBE,
                _PRESET_GENERIC,
            ]
        ),
    },
)
STEP_2_GENERIC_SCHEMA = vol.Schema(
    {
        vol.Required(CONF_EVENT): str,
        vol.Required(CONF_STATE): str,
        vol.Optional(CONF_EVENT_DATA, default=""): str,
        vol.Optional(CONF_STATE_MAP, default=""): str,
    },
)


@config_entries.HANDLERS.register(DOMAIN)
class EventSensorFlowHandler(config_entries.ConfigFlow):
    """Config flow for eventsensor."""

    VERSION = 1
    CONNECTION_CLASS = config_entries.CONN_CLASS_LOCAL_PUSH

    def __init__(self):
        """Initialize."""
        self._data_steps_config = {}

    async def _create_entry(self):
        event = self._data_steps_config.get(CONF_EVENT)
        if event == EVENT_STATE_CHANGED:
            return self.async_abort(reason="forbidden_event")

        unique_id = make_unique_id(self._data_steps_config)
        await self.async_set_unique_id(unique_id)
        self._abort_if_unique_id_configured()

        name = self._data_steps_config.get(CONF_NAME)
        entry_data = {
            CONF_NAME: name,
            CONF_EVENT: self._data_steps_config.get(CONF_EVENT),
            CONF_EVENT_DATA: self._data_steps_config.get(CONF_EVENT_DATA, {}),
            CONF_STATE: self._data_steps_config.get(CONF_STATE),
            CONF_STATE_MAP: self._data_steps_config.get(CONF_STATE_MAP, {}),
        }

        return self.async_create_entry(title=name, data=entry_data)

    def _parse_dict_fields(self, user_input, field):
        field_map = {}
        raw_field_map = user_input.get(field)
        if raw_field_map:
            field_map = parse_dict_from_ui_string(raw_field_map)
        self._data_steps_config[field] = field_map

    async def async_step_user(self, user_input=None):
        """Handle a flow initialized by the user."""
        if user_input is not None:
            self._data_steps_config[CONF_NAME] = user_input.get(CONF_NAME)

            event_source = user_input.get(CONF_INTEGRATION)
            if event_source == _EVENT_SOURCE_HUE:
                self._data_steps_config[CONF_EVENT] = "hue_event"
                self._data_steps_config[CONF_STATE] = "event"

            elif event_source == _EVENT_SOURCE_DECONZ:
                self._data_steps_config[CONF_EVENT] = "deconz_event"
                self._data_steps_config[CONF_STATE] = "event"

            else:
                return await self.async_step_generic()

            return await self.async_step_preset()

        return self.async_show_form(step_id="user", data_schema=STEP_1_INITIAL)

    async def async_step_preset(self, user_input=None):
        """Handle a flow initialized by the user."""
        if user_input is not None:
            type_id = user_input.get(CONF_TYPE_IDENTIFIER)
            identifier = user_input.get(CONF_IDENTIFIER)
            filter_map = {}
            if identifier:
                filter_map = {type_id: identifier}
            self._data_steps_config[CONF_EVENT_DATA] = filter_map

            preset_map = {}
            preset_config = user_input.get(CONF_PRESET_CONFIG)
            if preset_config == PRESET_HUE_DIMMER:
                preset_map = PRESET_HUE_DIMMER_MAPPING
            elif preset_config == PRESET_HUE_TAP:
                preset_map = PRESET_HUE_TAP_MAPPING
            elif preset_config == PRESET_HUE_BUTTON:
                preset_map = PRESET_HUE_BUTTON_MAPPING
            elif preset_config == PRESET_FOH:
                preset_map = PRESET_FOH_MAPPING
            elif preset_config == PRESET_AQARA_SMART_BUTTON:
                preset_map = PRESET_AQARA_SMART_BUTTON_MAPPING
            elif preset_config == PRESET_AQARA_CUBE:
                preset_map = PRESET_AQARA_CUBE_MAPPING
                self._data_steps_config[CONF_STATE] = "gesture"
            self._data_steps_config[CONF_STATE_MAP] = preset_map

            return await self.async_step_state_mapping()

        return self.async_show_form(step_id="preset", data_schema=STEP_2_PRECONFIGURED)

    async def async_step_generic(self, user_input=None):
        """Handle a flow initialized by the user."""
        if user_input is not None:
            self._data_steps_config[CONF_EVENT] = user_input.get(CONF_EVENT)
            self._data_steps_config[CONF_STATE] = user_input.get(CONF_STATE)
            self._parse_dict_fields(user_input, CONF_EVENT_DATA)
            self._parse_dict_fields(user_input, CONF_STATE_MAP)
            return await self._create_entry()

        return self.async_show_form(
            step_id="generic", data_schema=STEP_2_GENERIC_SCHEMA
        )

    async def async_step_state_mapping(self, user_input=None):
        """Handle a flow initialized by the user."""
        if user_input is not None:
            self._parse_dict_fields(user_input, CONF_STATE_MAP)
            return await self._create_entry()

        state_map_ui = make_string_ui_from_dict(
            self._data_steps_config.get(CONF_STATE_MAP, {})
        )
        return self.async_show_form(
            step_id="state_mapping",
            data_schema=vol.Schema(
                {vol.Optional(CONF_STATE_MAP, default=state_map_ui): str},
            ),
        )

    async def async_step_import(self, import_info):
        """Handle import from YAML config file."""
        self._data_steps_config.update(import_info)
        return await self._create_entry()

    @staticmethod
    @callback
    def async_get_options_flow(config_entry: config_entries.ConfigEntry):
        """Get the options flow for this handler to make a tariff change."""
        return EventSensorOptionsFlowHandler(config_entry)


class EventSensorOptionsFlowHandler(config_entries.OptionsFlow):
    """
    Handle the Options flow for `eventsensor` to edit the configuration.

    **entry.options is used as a container to make changes over entry.data**
    """

    def __init__(self, config_entry: config_entries.ConfigEntry):
        """Initialize the options flow handler with the config entry to modify."""
        self.config_entry = config_entry

    async def async_step_init(self, user_input=None):
        """Manage the options."""
        if user_input is not None:
            # Inverse conversion for mappings shown as strings
            for c in (CONF_EVENT_DATA, CONF_STATE_MAP):
                user_input[c] = parse_dict_from_ui_string(user_input[c])

            new_unique_id = make_unique_id(user_input)
            if self.config_entry.unique_id != new_unique_id:
                # check change of unique_id to prevent collisions
                for entry in filter(
                    lambda x: x.entry_id != self.config_entry.entry_id,
                    self.hass.config_entries.async_entries(DOMAIN),
                ):
                    if entry.unique_id == new_unique_id:
                        _LOGGER.error(
                            "The `unique_id` is already used by another sensor"
                        )
                        return self.async_abort(reason="already_configured")

            return self.async_create_entry(title="", data=user_input)

        # Fill options with entry data
        container = self.config_entry.data
        name = container.get(CONF_NAME)
        event = container.get(CONF_EVENT)
        state = container.get(CONF_STATE)
        filter_ev_str = make_string_ui_from_dict(container.get(CONF_EVENT_DATA, {}))
        state_map_str = make_string_ui_from_dict(container.get(CONF_STATE_MAP, {}))
        return self.async_show_form(
            step_id="init",
            data_schema=vol.Schema(
                {
                    vol.Required(CONF_NAME, default=name): str,
                    vol.Required(CONF_EVENT, default=event): str,
                    vol.Required(CONF_STATE, default=state): str,
                    vol.Optional(CONF_EVENT_DATA, default=filter_ev_str): str,
                    vol.Optional(CONF_STATE_MAP, default=state_map_str): str,
                },
            ),
        )
