"""Constants for eventsensor."""
import re

from homeassistant.const import CONF_EVENT, CONF_EVENT_DATA, CONF_STATE
from homeassistant.util import slugify

# Base component constants
DOMAIN = "eventsensor"
PLATFORM = "sensor"
DOMAIN_DATA = f"{DOMAIN}_data"

CONF_STATE_MAP = "state_map"

PRESET_AQARA_CUBE = "Aqara Cube"
PRESET_AQARA_CUBE_MAPPING = {
    0: "Wake",
    1: "Shake",
    2: "Drop",
    3: "Flip90",
    4: "Flip180",
    5: "Push",
    6: "DoubleTap",
    7: "RotCW",
    8: "RotCCW",
}
PRESET_AQARA_SMART_BUTTON = "Aqara Smart Button"
PRESET_AQARA_SMART_BUTTON_MAPPING = {
    1000: "click",
    1001: "hold",
    1002: "click_up",
    1003: "hold_up",
    1004: "2_click",
    1005: "3_click",
    1006: "4_click",
    1010: "5_click",
}
PRESET_FOH = "FoH Switch"
PRESET_FOH_MAPPING = {
    16: "left_upper_press",
    20: "left_upper_release",
    17: "left_lower_press",
    21: "left_lower_release",
    18: "right_lower_press",
    22: "right_lower_release",
    19: "right_upper_press",
    23: "right_upper_release",
    100: "double_upper_press",
    101: "double_upper_release",
    98: "double_lower_press",
    99: "double_lower_release",
}
PRESET_HUE_DIMMER = "Hue Dimmer Switch"
PRESET_HUE_DIMMER_MAPPING = {
    1000: "1_click",
    2000: "2_click",
    3000: "3_click",
    4000: "4_click",
    1001: "1_hold",
    2001: "2_hold",
    3001: "3_hold",
    4001: "4_hold",
    1002: "1_click_up",
    2002: "2_click_up",
    3002: "3_click_up",
    4002: "4_click_up",
    1003: "1_hold_up",
    2003: "2_hold_up",
    3003: "3_hold_up",
    4003: "4_hold_up",
}
PRESET_HUE_TAP = "Hue Tap Switch"
PRESET_HUE_TAP_MAPPING = {
    34: "1_click",
    16: "2_click",
    17: "3_click",
    18: "4_click",
}
PRESET_HUE_BUTTON = "Hue Smart Button"
PRESET_HUE_BUTTON_MAPPING = {
    1000: "1_click",
    1001: "1_hold",
    1002: "1_click_up",
    1003: "1_hold_up"
}
_rg_dict_extraction = re.compile(r"({[^{}]+})")


def make_unique_id(sensor_data: dict) -> str:
    """
    Generate an unique id from the listened event + data filters.

    Used for both the the config entry and the sensor entity.
    """
    event: str = sensor_data.get(CONF_EVENT)
    state: str = sensor_data.get(CONF_STATE)
    filter_event: dict = dict(sensor_data.get(CONF_EVENT_DATA, {}))
    state_map: dict = dict(sensor_data.get(CONF_STATE_MAP, {}))
    return "_".join([event, slugify(str(filter_event)), state, slugify(str(state_map))])


# Workaround for config entry data being stored as strings always
def parse_numbers(raw_item):
    """Enable numerical values, like press codes for remotes."""
    if isinstance(raw_item, dict):
        return {parse_numbers(k): parse_numbers(v) for k, v in raw_item.items()}

    try:
        return int(raw_item)
    except ValueError:
        try:
            return float(raw_item)
        except ValueError:
            return raw_item


# Workaround for state extraction from nested data in event
def extract_state_from_event(state_key: str, event_data: dict):
    """
    Extract information from the event data to make a new sensor state.

    Use 'dot syntax' to point for nested attributes, like `service_data.entity_id`
    """
    if state_key in event_data:
        return event_data[state_key]
    elif state_key.split(".")[0] in event_data:
        try:
            nested_data = event_data
            for level in state_key.split("."):
                nested_data = nested_data[level]
            # Don't use dicts as state!
            if isinstance(nested_data, dict):
                return str(nested_data)
            return nested_data
        except (IndexError, TypeError):
            pass
    return "bad_state"


# Workaround lack of UI input field to edit yaml inside a ConfigFlow -> string repr
def make_string_ui_from_dict(data: dict) -> str:
    """
    Generate a readable string for a nested dict to show & edit in on UI.

    Assume syntax like:
    `key1: value1, key2: {subk1: value2, subk2: value3}, ...`
    """
    pairs = []
    for key, value in data.items():
        if isinstance(value, dict):
            value = "{" + make_string_ui_from_dict(value) + "}"
        pairs.append(f"{key}: {value}")

    return ", ".join(pairs)


def _from_str_to_dict(raw_data: str) -> dict:
    """Assume format `key1: value1, key2: value2, ...`."""
    raw_pairs = raw_data.split(",")

    def _parse_item(raw_key: str):
        return raw_key.lstrip(" ").rstrip(" ").rstrip(":")

    data_out = {}
    for pair in raw_pairs:
        if ":" not in pair:
            break
        key, value = pair.split(":", maxsplit=1)
        data_out[_parse_item(key)] = _parse_item(value)

    return data_out


def _walk_nested_dict(container: dict, substitutions: dict):
    for key, value in container.items():
        if isinstance(value, dict):
            _walk_nested_dict(value, substitutions)
        elif value in substitutions:
            new_value = substitutions[value]
            if isinstance(new_value, dict):
                _walk_nested_dict(new_value, substitutions)

            # Making substitution
            container[key] = new_value


def parse_dict_from_ui_string(str_use) -> dict:
    """
    Parse a string field into a nested dict.

    Workaround to enable UI edition of a dict, as a 'YAML' input is not available.

    Assume syntax like:
    `key1: value1, key2: {subk1: value2, subk2: value3}, ...`
    """
    substitutions = {}
    counter_subs = 0
    str_subs = str_use
    count_nesting = 0
    while "{" in str_subs and count_nesting < 10:
        count_nesting += 1
        for found in _rg_dict_extraction.findall(str_subs):
            parsed_piece = _from_str_to_dict(found[1:-1])

            key_sub = f"SUB{counter_subs:03d}"
            substitutions[key_sub] = parsed_piece
            counter_subs += 1

            str_subs = str_subs.replace(found, key_sub, 1)

    # last parse for final root keys
    data = _from_str_to_dict(str_subs)

    # Now to substitute values:
    if substitutions:
        _walk_nested_dict(data, substitutions)

    return data


def check_dict_is_contained_in_another(filter_data: dict, data: dict) -> bool:
    """
    Check if a dict is contained in another one.

    * Works with nested dicts by using _dot notation_ in the filter_data keys
      so a filter with
      `{"base_key.sub_key": "value"}`
      will look for dicts containing
      `{"base_key": {"sub_key": "value"}}`
    """
    for key, value in filter_data.items():
        if key in data:
            if data[key] != value:
                return False
            continue

        if "." in key:
            base_key, sub_key = key.split(".", maxsplit=1)
            if base_key not in data:
                return False

            base_value = data[base_key]
            if not isinstance(base_value, dict):
                return False

            if not check_dict_is_contained_in_another({sub_key: value}, base_value):
                return False

            continue

        return False

    return True
