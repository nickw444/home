from appdaemon_testing.pytest import automation_fixture
from apps.light_fade_in_out import LightFadeInOut, CONF_ENABLE_ENTITY, CONF_PRESENCE_ENTITY, \
    CONF_START_TIME_ENTITY, CONF_FADE_IN_DURATION, CONF_FADE_OUT_DURATION, CONF_LIGHTS, \
    CONF_ENTITY_ID, CONF_COLOR_TEMP_KELVIN


def test_callbacks_are_registered(hass_driver, app: LightFadeInOut):
    pass


@automation_fixture(
    LightFadeInOut,
    args={
        CONF_ENABLE_ENTITY: 'input_boolean.enable',
        CONF_PRESENCE_ENTITY: 'device_tracker.iphone',
        CONF_START_TIME_ENTITY: 'input_number.start_time',
        CONF_FADE_IN_DURATION: '00:10:00',
        CONF_FADE_OUT_DURATION: '00:10:00',
        CONF_LIGHTS: [
            {
                CONF_ENTITY_ID: 'light.light',
                CONF_COLOR_TEMP_KELVIN: '2500K',
            }
        ]
    },
)
def app():
    pass
