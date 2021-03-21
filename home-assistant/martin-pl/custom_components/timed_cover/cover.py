import logging
import threading
import time
from typing import Any, Optional

import voluptuous as vol

import homeassistant.helpers.config_validation as cv
from homeassistant.components.cover import (
    CoverEntity, PLATFORM_SCHEMA,
    DOMAIN, ATTR_POSITION, ATTR_CURRENT_POSITION)
from homeassistant.const import (
    CONF_FRIENDLY_NAME, SERVICE_OPEN_COVER,
    ATTR_ENTITY_ID, SERVICE_CLOSE_COVER, SERVICE_STOP_COVER, ATTR_DEVICE_CLASS,
    STATE_CLOSED)
from homeassistant.core import callback
from homeassistant.helpers.event import async_track_state_change_event
from homeassistant.helpers.restore_state import RestoreEntity

_LOGGER = logging.getLogger(__name__)

CONF_TRAVEL_TIME_DOWN = 'travel_time_down'
CONF_TRAVEL_TIME_UP = 'travel_time_up'
CONF_COVER_ENTITY_ID = 'cover_entity_id'

PLATFORM_SCHEMA = PLATFORM_SCHEMA.extend(
    {
        vol.Required(CONF_FRIENDLY_NAME): cv.string,
        vol.Required(CONF_COVER_ENTITY_ID): cv.entity_id,
        vol.Required(CONF_TRAVEL_TIME_UP): cv.positive_float,
        vol.Required(CONF_TRAVEL_TIME_DOWN): cv.positive_float,
    }
)


def setup_platform(hass, config, add_entities, discovery_info=None):
    """Set up the sensor platform."""
    add_entities([
        TimedCover(
            friendly_name=config[CONF_FRIENDLY_NAME],
            cover_entity_id=config[CONF_COVER_ENTITY_ID],
            travel_time_up=config[CONF_TRAVEL_TIME_UP],
            travel_time_down=config[CONF_TRAVEL_TIME_DOWN],
        )
    ])

    return True


class TimedCover(CoverEntity, RestoreEntity):
    def __init__(self, friendly_name: str, cover_entity_id: str,
                 travel_time_up: float, travel_time_down: float):
        self._friendly_name = friendly_name
        self._cover_entity_id = cover_entity_id
        self._travel_time_up = travel_time_up
        self._travel_time_down = travel_time_down

        self._mutex = threading.Lock()
        self._current_position = None
        self._active_movement = None
        self._active_thread: Optional[threading.Thread] = None

    async def async_added_to_hass(self):
        @callback
        def async_state_changed_listener(*_: Any) -> None:
            """Handle child updates."""
            self.async_schedule_update_ha_state(True)

        async_track_state_change_event(
            self.hass, [self._cover_entity_id], async_state_changed_listener
        )

        restored_state = await self.async_get_last_state()
        if restored_state is None:
            return

        restored_curr_position = restored_state.attributes.get(
            ATTR_CURRENT_POSITION)
        self._current_position = restored_curr_position
        _LOGGER.debug("Restored position state: %s", restored_curr_position)

    def is_active_invocation(self):
        return self._active_thread is not None and self._active_thread.ident == threading.get_ident()

    def run_set_position(self, curr_position: Optional[float],
                         target_position: float):
        _LOGGER.debug("Invocation %s started", threading.get_ident())
        with self._mutex:
            _LOGGER.debug("Invocation %s acquired lock", threading.get_ident())
            if not self.is_active_invocation():
                _LOGGER.debug(
                    "Invocation (id: %s) no longer active. (Superseded by %s))",
                    threading.get_ident(),
                    self._active_thread and self._active_thread.ident,
                )
                return

            service, travel_time = (
                (SERVICE_CLOSE_COVER, self._travel_time_down) if
                target_position < curr_position else
                (SERVICE_OPEN_COVER, self._travel_time_up)
            )
            partial_travel_perc = float(
                abs(target_position - curr_position)) / 100
            partial_travel_time = travel_time * partial_travel_perc

            self.hass.services.call(
                DOMAIN, service, {ATTR_ENTITY_ID: self._cover_entity_id})
            self._active_movement = service
            self.async_schedule_update_ha_state(True)

            _LOGGER.debug("Sleeping for %d seconds for %s",
                          partial_travel_time, service)
            time.sleep(partial_travel_time)
            if not self.is_active_invocation():
                _LOGGER.debug(
                    "Invocation (id: %s) no longer active. (Superseded by %s))",
                    threading.get_ident(),
                    self._active_thread and self._active_thread.ident,
                    )
                return

            self.hass.services.call(DOMAIN, SERVICE_STOP_COVER,
                                    {ATTR_ENTITY_ID: self._cover_entity_id})
            self._active_movement = None
            _LOGGER.debug("Movement to %02f finished", target_position)
            self.async_schedule_update_ha_state(True)

    @property
    def name(self) -> str:
        """Return the name of the entity."""
        return self._friendly_name

    @property
    def device_class(self) -> Optional[str]:
        """Return the class of this device, from component DEVICE_CLASSES."""
        state = self.hass.states.get(self._cover_entity_id)
        if state:
            return state.attributes.get(ATTR_DEVICE_CLASS)

    @property
    def current_cover_position(self):
        """Return current position of cover.

        None is unknown, 0 is closed, 100 is fully open.
        """
        return self._current_position

    @property
    def is_opening(self):
        """Return if the cover is opening or not."""
        return self._active_movement == SERVICE_OPEN_COVER

    @property
    def is_closing(self):
        """Return if the cover is closing or not."""
        return self._active_movement == SERVICE_CLOSE_COVER

    @property
    def is_closed(self):
        state = self.hass.states.get(self._cover_entity_id)
        if state:
            return state == STATE_CLOSED

    def open_cover(self, **kwargs: Any) -> None:
        self._current_position = 100
        self._active_movement = None
        self._active_thread = None
        self.hass.services.call(DOMAIN, SERVICE_OPEN_COVER,
                                {ATTR_ENTITY_ID: self._cover_entity_id})

    def close_cover(self, **kwargs: Any) -> None:
        self._current_position = 0
        self._active_movement = None
        self._active_thread = None
        self.hass.services.call(DOMAIN, SERVICE_CLOSE_COVER,
                                {ATTR_ENTITY_ID: self._cover_entity_id})

    def stop_cover(self, **kwargs):
        self._active_movement = None
        self._active_thread = None
        self.hass.services.call(DOMAIN, SERVICE_STOP_COVER,
                                {ATTR_ENTITY_ID: self._cover_entity_id})


    def set_cover_position(self, **kwargs):
        """Move the cover to a specific position."""
        curr_position = self._current_position
        target_position = kwargs[ATTR_POSITION]
        _LOGGER.debug("Setting cover position from: %s to: %s",
                      curr_position, target_position)

        if curr_position is None:
            _LOGGER.warning(
                "Cannot set position when cover position is unknown (entity_id: %s)",
                self.entity_id)
            return

        if curr_position == target_position:
            return

        self._start_set_position(curr_position, target_position, False)

    def _start_set_position(self,
                            curr_position: float,
                            target_position: float,
                            ):
        self._current_position = target_position
        self._active_thread = threading.Thread(
            target=self.run_set_position,
            args=(curr_position, target_position)
        )
        self._active_thread.start()
