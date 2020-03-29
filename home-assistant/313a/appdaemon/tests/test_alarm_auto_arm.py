import mock
from appdaemon_testing.pytest import automation_fixture

from apps.alarm_auto_arm import (
    AlarmAutoArm,
    CONF_ALARM_ENTITY,
    CONF_ARMING_CODE,
    CONF_ENABLE_ENTITY,
    CONF_ENABLE_OVERRIDE_ENTITY,
    CONF_PRESENCE_ENTITY,
)

ALARM_ENTITY = "alarm_control_panel.alarm"
ARMING_CODE = "123456"
ENABLE_ENTITY = "input_boolean.enable"
ENABLE_OVERRIDE_ENTITY = "input_boolean.enable_override"
PRESENCE_ENTITY = "group.people"


def test_callbacks_are_registered(hass_driver, auto_arm):
    listen_state = hass_driver.get_mock("listen_state")
    assert listen_state.call_count == 3
    listen_state.assert_has_calls(
        [
            mock.call(auto_arm.on_presence_state_change, PRESENCE_ENTITY),
            mock.call(auto_arm.on_enable_change, ENABLE_ENTITY),
            mock.call(auto_arm.on_enable_change, ENABLE_OVERRIDE_ENTITY),
        ]
    )


def test_when_everyone_leaves_then_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "not_home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_called_once_with(
        "alarm_control_panel.alarm_arm_away", entity_id=ALARM_ENTITY
    )


def test_when_everyone_leaves_and_not_enabled_then_do_not_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "off")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "not_home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_not_called()


def test_when_everyone_leaves_and_not_enabled_via_override_then_do_not_arm(
    hass_driver, auto_arm
):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "off")

    hass_driver.set_state(PRESENCE_ENTITY, "not_home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_not_called()


def test_when_everyone_leaves_and_armed_do_not_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "home")
        hass_driver.set_state(ALARM_ENTITY, "armed_away")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "not_home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_not_called()


def test_when_re_enabled_then_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "off")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(ENABLE_ENTITY, "on")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_called_once_with(
        "alarm_control_panel.alarm_arm_away", entity_id=ALARM_ENTITY
    )


def test_when_override_re_enabled_then_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "off")

    hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_called_once_with(
        "alarm_control_panel.alarm_arm_away", entity_id=ALARM_ENTITY
    )


def test_when_re_enabled_but_override_disabled_then_do_not_arm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "disarmed")
        hass_driver.set_state(ENABLE_ENTITY, "off")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "off")

    hass_driver.set_state(ENABLE_ENTITY, "on")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_not_called()


def test_when_someone_home_disarm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "armed_away")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_called_once_with(
        "alarm_control_panel.alarm_disarm", entity_id=ALARM_ENTITY, code=ARMING_CODE
    )


def test_fires_event_when_disarmed(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "armed_away")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "home")

    fire_event = hass_driver.get_mock("fire_event")
    fire_event.assert_called_once_with("auto_arm_disarmed")


def test_when_someone_home_but_override_disabled_dont_disarm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "armed_away")
        hass_driver.set_state(ENABLE_ENTITY, "on")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "off")

    hass_driver.set_state(PRESENCE_ENTITY, "home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_not_called()


def test_when_someone_home_but_disabled_then_disarm(hass_driver, auto_arm):
    with hass_driver.setup():
        hass_driver.set_state(PRESENCE_ENTITY, "not_home")
        hass_driver.set_state(ALARM_ENTITY, "armed_away")
        hass_driver.set_state(ENABLE_ENTITY, "off")
        hass_driver.set_state(ENABLE_OVERRIDE_ENTITY, "on")

    hass_driver.set_state(PRESENCE_ENTITY, "home")

    call_service = hass_driver.get_mock("call_service")
    call_service.assert_called_once_with(
        "alarm_control_panel.alarm_disarm", entity_id=ALARM_ENTITY, code=ARMING_CODE
    )


@automation_fixture(AlarmAutoArm, args={
    CONF_ALARM_ENTITY: ALARM_ENTITY,
    CONF_ARMING_CODE: ARMING_CODE,
    CONF_ENABLE_ENTITY: ENABLE_ENTITY,
    CONF_ENABLE_OVERRIDE_ENTITY: ENABLE_OVERRIDE_ENTITY,
    CONF_PRESENCE_ENTITY: PRESENCE_ENTITY,
})
def auto_arm():
    pass
