import asyncio
from typing import Generic, Optional, TypeVar

T = TypeVar('T')


class EventWithValue(Generic[T]):
    def __init__(self):
        self._event = asyncio.Event()
        self._value: Optional[T] = None

    async def wait(self) -> T:
        await self._event.wait()
        return self._value

    def set(self, value):
        self._value = value
        self._event.set()

    def clear(self):
        return self._event.clear()

    def is_set(self):
        return self._event.is_set()
