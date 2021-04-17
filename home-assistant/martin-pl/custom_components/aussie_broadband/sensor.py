import logging
from datetime import timedelta
from typing import Optional

import voluptuous as vol
from aussiebb import AussieBB

import homeassistant.helpers.config_validation as cv
from homeassistant.components.sensor import (PLATFORM_SCHEMA)
from homeassistant.const import CONF_SCAN_INTERVAL
from homeassistant.helpers.update_coordinator import (
    DataUpdateCoordinator,
    CoordinatorEntity)

_LOGGER = logging.getLogger(__name__)

CONF_USERNAME = 'username'
CONF_PASSWORD = 'password'
CONF_SERVICE_ID = 'service_id'

DEFAULT_SCAN_INTERVAL = timedelta(minutes=30)

PLATFORM_SCHEMA = PLATFORM_SCHEMA.extend(
    {
        vol.Required(CONF_USERNAME): cv.string,
        vol.Required(CONF_PASSWORD): cv.string,
        vol.Required(CONF_SERVICE_ID): cv.positive_int,
        vol.Optional(CONF_SCAN_INTERVAL, default=DEFAULT_SCAN_INTERVAL): vol.All(
            cv.time_period, cv.positive_timedelta
        )
    }
)


async def async_setup_platform(hass, config, add_entities, discovery_info=None):
    """Set up the sensor platform."""
    def create_client():
      return AussieBB(config[CONF_USERNAME], config[CONF_PASSWORD])

    client = await hass.async_add_executor_job(create_client)
    service_id = config[CONF_SERVICE_ID]

    async def async_update_data():
        return await hass.async_add_executor_job(client.get_usage, service_id)

    coordinator = DataUpdateCoordinator(hass, _LOGGER, name="sensor",
                                        update_interval=config[
                                            CONF_SCAN_INTERVAL],
                                        update_method=async_update_data)
    await coordinator.async_refresh()

    add_entities([
        CounterEntity(coordinator, service_id, "Total Usage", "usedMb", "MB"),
        CounterEntity(coordinator, service_id, "Downloaded", "downloadedMb", "MB"),
        CounterEntity(coordinator, service_id, "Uploaded", "uploadedMb", "MB"),
        CounterEntity(coordinator, service_id, "Billing Cycle Length",
                      "daysTotal", "days"),
        CounterEntity(coordinator, service_id, "Billing Cycle Remaining",
                      "daysRemaining", "days"),
    ])

    return True


class CounterEntity(CoordinatorEntity):
    def __init__(self, coordinator: DataUpdateCoordinator, service_id: int, name: str, attribute: str, unit_of_measurement: str):
        super(CounterEntity, self).__init__(coordinator)

        self._service_id = service_id
        self._name = name
        self._attribute = attribute
        self._unit_of_measurement = unit_of_measurement

    @property
    def name(self):
        """Return the name of the sensor."""
        return self._name

    @property
    def unit_of_measurement(self):
        """Return unit of measurement."""
        return self._unit_of_measurement

    @property
    def state(self):
        """Return the state of the device."""
        return self.coordinator.data[self._attribute]

    @property
    def unique_id(self) -> Optional[str]:
        return 'abb-' + str(self._service_id) + ':' + self._attribute
