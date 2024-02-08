import appdaemon.plugins.hass.hassapi as hass

CONF_ALARM_ENTITY = "alarm_entity"
CONF_ARMING_CODE = "arming_code"
CONF_PRESENCE_ENTITY = "presence_entity"
CONF_ENABLE_ENTITY = "enable_entity"
CONF_ENABLE_OVERRIDE_ENTITY = "enable_override_entity"


class AlarmAutoArm(hass.Hass):
    """
    When:
        * Everyone leaves
        * Auto-arm re-enabled (via schedule)
        * When enable input boolean re-enabled
    Conditions:
        * Panel not already armed
        * No one is home
        * Input boolean is enabled
    Then:
        * Arm

    When:
        * Someone arrives home
        * Auto-arm is disabled (via schedule)
        * When input boolean disabled
    Conditions:
        * Panel is armed
    Then:
        * Disarm the alarm panel
    """

    def initialize(self):
        self.listen_state(
            self.on_presence_state_change, self.args[CONF_PRESENCE_ENTITY]
        )
        self.listen_state(self.on_enable_change, self.args[CONF_ENABLE_ENTITY])
        self.listen_state(
            self.on_enable_override_change, self.args[CONF_ENABLE_OVERRIDE_ENTITY]
        )

    def on_presence_state_change(self, entity, attribute, old, new, kwargs):
        if new == "off":
            self._maybe_arm()
        else:
            self._maybe_disarm()

    def on_enable_override_change(self, entity, attribute, old, new, kwargs):
        if new == "on":
            self._maybe_arm()
        else:
            self._maybe_disarm(force=True)

    def on_enable_change(self, entity, attribute, old, new, kwargs):
        if new == "on":
            self._maybe_arm()
        else:
            self._maybe_disarm()

    def _maybe_arm(self):
        if not self._can_arm():
            return

        self.call_service(
            "alarm_control_panel/alarm_arm_away", entity_id=self.args[CONF_ALARM_ENTITY]
        )
        self.fire_event("auto_arm_armed")
        self.call_service(
            "logbook/log",
            name="Alarm Auto Arm",
            message="Automatic arm",
        )

    def _maybe_disarm(self, force=False):
        if not self._can_disarm(force=force):
            return

        self.call_service(
            "alarm_control_panel/alarm_disarm",
            entity_id=self.args[CONF_ALARM_ENTITY],
            code=self.args[CONF_ARMING_CODE],
        )
        self.fire_event("auto_arm_disarmed")
        self.call_service(
            "logbook/log",
            name="Alarm Auto Arm",
            message="Automatic disarm",
        )

    def _can_arm(self):
        alarm_state = self.get_state(self.args[CONF_ALARM_ENTITY])
        presence_state = self.get_state(self.args[CONF_PRESENCE_ENTITY])
        enable_state = self.get_state(self.args[CONF_ENABLE_ENTITY])
        return (
            alarm_state == "disarmed"
            and presence_state == "off"
            and enable_state == "on"
            and self._is_enabled()
        )

    def _can_disarm(self, force):
        alarm_state = self.get_state(self.args[CONF_ALARM_ENTITY])

        return alarm_state != "disarmed" and (force or self._is_enabled())

    def _is_enabled(self):
        return self.get_state(self.args[CONF_ENABLE_OVERRIDE_ENTITY]) == "on"
