from dataclasses import dataclass
from typing import Any, Optional, Dict


@dataclass(frozen=True)
class Error:
    code: int
    message: str

    @classmethod
    def deserialize(cls, o: Dict[str, Any]):
        return Error(**o)


@dataclass(frozen=True)
class Response:
    id: int
    type: str

    @classmethod
    def deserialize(cls, o: Dict[str, Any]) -> 'Response':
        if o['type'] == 'result':
            return ResultResponse.deserialize(o)
        elif o['type'] == 'pong':
            return PongResponse.deserialize(o)
        elif o['type'] == 'event':
            return EventResponse.deserialize(o)
        else:
            raise AssertionError('Unknown response type: ', o['type'])


@dataclass(frozen=True)
class PongResponse(Response):
    @classmethod
    def deserialize(cls, o: Dict[str, Any]):
        return PongResponse(
            id=o['id'],
            type=o['type'],
        )


@dataclass(frozen=True)
class ResultResponse(Response):
    success: Optional[bool]
    result: Optional[Any]
    error: Optional[Error]

    @classmethod
    def deserialize(cls, o: Dict[str, Any]):
        error_data = o.get('error')
        error = error_data and Error.deserialize(error_data)

        return ResultResponse(
            id=o['id'],
            type=o['type'],
            success=o.get('success'),
            result=o.get('result'),
            error=error,
        )


@dataclass(frozen=True)
class Event:
    data: Any
    event_type: str
    time_fired: str
    origin: str


@dataclass(frozen=True)
class EventResponse(Response):
    event: Event

    @classmethod
    def deserialize(cls, o: Dict[str, Any]):
        return EventResponse(
            id=o['id'],
            type=o['type'],
            event=Event(
                data=o['event']['data'],
                event_type=o['event']['event_type'],
                time_fired=o['event']['time_fired'],
                origin=o['event']['origin'],
            )
        )
