"""Event sensor."""
import logging
from typing import Any, Callable, List

import voluptuous as vol

from homeassistant.components.sensor import PLATFORM_SCHEMA
from homeassistant.config_entries import ConfigEntry, SOURCE_IMPORT
from homeassistant.const import CONF_EVENT, CONF_EVENT_DATA, CONF_NAME, CONF_STATE
from homeassistant.core import callback
from homeassistant.helpers.config_validation import string
from homeassistant.helpers.event import Event
from homeassistant.helpers.restore_state import RestoreEntity
from homeassistant.helpers.typing import (
    ConfigType,
    DiscoveryInfoType,
    HomeAssistantType,
)

from .common import (
    CONF_STATE_MAP,
    DOMAIN,
    DOMAIN_DATA,
    check_dict_is_contained_in_another,
    extract_state_from_event,
    make_unique_id,
    parse_numbers,
)

_LOGGER = logging.getLogger(__name__)

PLATFORM_SCHEMA = PLATFORM_SCHEMA.extend(
    {
        vol.Required(CONF_NAME): string,
        vol.Required(CONF_STATE): string,
        vol.Required(CONF_EVENT): string,
        vol.Optional(CONF_EVENT_DATA): dict,
        vol.Optional(CONF_STATE_MAP): dict,
    },
    extra=vol.ALLOW_EXTRA,
)


async def async_setup_platform(
    hass: HomeAssistantType,
    config: ConfigType,
    async_add_entities: Callable[[List[Any], bool], None],
    discovery_info: DiscoveryInfoType = None,
):
    """
    Set up event sensors from configuration.yaml as a sensor platform.

    Left just to read deprecated manual configuration.
    """
    if config:
        hass.async_create_task(
            hass.config_entries.flow.async_init(
                DOMAIN, data=config, context={"source": SOURCE_IMPORT}
            )
        )
        _LOGGER.warning(
            "Manual yaml config is deprecated. "
            "You can remove it now, as it has been migrated to config entry, "
            "handled in the Integrations menu [Sensor %s, event: %s]",
            config.get(CONF_NAME),
            config.get(CONF_EVENT),
        )

    return True


async def async_setup_entry(
    hass: HomeAssistantType,
    config_entry: ConfigEntry,
    async_add_entities: Callable[[List[Any], bool], None],
):
    """Set up the component sensors from a config entry."""
    if DOMAIN_DATA not in hass.data:
        hass.data[DOMAIN_DATA] = {}

    async_add_entities([EventSensor(config_entry.unique_id, config_entry.data)], False)

    # add an update listener to enable edition by OptionsFlow
    hass.data[DOMAIN_DATA][config_entry.entry_id] = config_entry.add_update_listener(
        update_listener
    )


async def update_listener(hass: HomeAssistantType, entry: ConfigEntry):
    """Update when config_entry options update."""
    changes = len(entry.options) > 1 and entry.data != entry.options
    if changes:
        # update entry replacing data with new options, and updating unique_id and title
        _LOGGER.debug(
            f"Config entry update with {entry.options} and unique_id:{entry.unique_id}"
        )
        hass.config_entries.async_update_entry(
            entry,
            title=entry.options[CONF_NAME],
            data=entry.options,
            options={},
            unique_id=make_unique_id(entry.options),
        )
        hass.async_create_task(hass.config_entries.async_reload(entry.entry_id))


class EventSensor(RestoreEntity):
    """Sensor to store information originated with events."""

    should_poll = False
    icon = "mdi:bullseye-arrow"

    def __init__(self, unique_id: str, sensor_data: dict):
        """Set up a new sensor mirroring some event."""
        self._unique_id = unique_id
        self._name = sensor_data.get(CONF_NAME)
        self._event = sensor_data.get(CONF_EVENT)
        self._state_key = sensor_data.get(CONF_STATE)
        self._event_data = parse_numbers(sensor_data.get(CONF_EVENT_DATA, {}))
        self._state_map = parse_numbers(sensor_data.get(CONF_STATE_MAP, {}))

        self._event_listener = None
        self._state = None
        self._attributes = {}

    @property
    def name(self):
        """Return the name of the entity."""
        return self._name

    @property
    def unique_id(self) -> str:
        """Return a unique ID, made with the event name and data filters."""
        return self._unique_id

    @property
    def state(self):
        """Return the state of the entity."""
        return self._state

    @property
    def state_attributes(self):
        """Return the state attributes."""
        return self._attributes

    async def async_added_to_hass(self) -> None:
        """Add event listener when adding entity to Home Assistant."""
        # Recover last state
        last_state = await self.async_get_last_state()
        if last_state is not None:
            self._state = last_state.state
            self._attributes = dict(last_state.attributes)

        @callback
        def async_update_sensor(event: Event):
            """Update state when event is received."""
            if check_dict_is_contained_in_another(self._event_data, event.data):
                # Extract new state
                new_state = extract_state_from_event(self._state_key, event.data)

                # Apply custom state mapping
                if new_state in self._state_map:
                    new_state = self._state_map[new_state]

                self._state = new_state
                self._attributes = {
                    **event.data,
                    "origin": event.origin.name,
                    "time_fired": event.time_fired,
                }
                _LOGGER.debug("%s: New state: %s", self.entity_id, self._state)
                self.async_write_ha_state()

        # Listen for event
        self._event_listener = self.hass.bus.async_listen(
            self._event, async_update_sensor
        )
        _LOGGER.debug(
            "%s: Added sensor listening to '%s' with event data: %s",
            self.entity_id,
            self._event,
            self._event_data,
        )

    async def async_will_remove_from_hass(self):
        """Remove listeners when removing entity from Home Assistant."""
        if self._event_listener is not None:
            self._event_listener()
            self._event_listener = None
            _LOGGER.debug("%s: Removed event listener", self.entity_id)
