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
        self.listen_state(self.on_enable_change, self.args[CONF_ENABLE_OVERRIDE_ENTITY])

    def on_presence_state_change(self, entity, attribute, old, new, kwargs):
        if new == "not_home":
            self._maybe_arm()
        else:
            self._maybe_disarm()

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

    def _maybe_disarm(self):
        if not self._can_disarm():
            return

        self.call_service(
            "alarm_control_panel/alarm_disarm",
            entity_id=self.args[CONF_ALARM_ENTITY],
            code=self.args[CONF_ARMING_CODE],
        )
        self.fire_event("auto_arm_disarmed")

    def _can_arm(self):
        alarm_state = self.get_state(self.args[CONF_ALARM_ENTITY])
        presence_state = self.get_state(self.args[CONF_PRESENCE_ENTITY])
        enable_state = self.get_state(self.args[CONF_ENABLE_ENTITY])
        return (
            alarm_state == "disarmed"
            and presence_state == "not_home"
            and enable_state == "on"
            and self._is_enabled()
        )

    def _can_disarm(self):
        alarm_state = self.get_state(self.args[CONF_ALARM_ENTITY])

        return alarm_state == "armed_away" and self._is_enabled()

    def _is_enabled(self):
        return self.get_state(self.args[CONF_ENABLE_OVERRIDE_ENTITY]) == "on"
