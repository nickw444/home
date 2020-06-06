import datetime
from typing import Any, Optional

import voluptuous as vol


def time_delta(value: Any) -> datetime.timedelta:
    parsed = [int(x) for x in value.split(":")]

    if len(parsed) == 2:
        hour, minute = parsed
        second = 0
    elif len(parsed) == 3:
        hour, minute, second = parsed
    else:
        raise vol.Invalid('Invalid time delta')

    return datetime.timedelta(hours=hour, minutes=minute, seconds=second)


def time(value: Any) -> datetime.time:
    """Validate and transform a time."""
    if isinstance(value, datetime.time):
        return value

    try:
        time_val = parse_time(value)
    except TypeError:
        raise vol.Invalid("Not a parseable type")

    if time_val is None:
        raise vol.Invalid(f"Invalid time specified: {value}")

    return time_val


def parse_time(time_str: str) -> Optional[datetime.time]:
    parts = str(time_str).split(":")
    if len(parts) < 2:
        return None
    try:
        hour = int(parts[0])
        minute = int(parts[1])
        second = int(parts[2]) if len(parts) > 2 else 0
        return datetime.time(hour, minute, second)
    except ValueError:
        # ValueError if value cannot be converted to an int or not in range
        return None
