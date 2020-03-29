import asyncio
from collections import defaultdict
from typing import Optional, Callable, NamedTuple

from .hass_client import HassClient
from .hass_response import EventResponse


class StateSpy(NamedTuple):
    entity_id: str
    attribute: Optional[str]
    handler: Callable


class Mate:
    def __init__(self, client: HassClient):
        self.client = client

        self.client.subscribe_events(handler=self._handle_state_change, event_type='state_change')
        self._cached_states = {}
        self._spys = defaultdict(lambda: [])

    async def poll_states(self):
        while True:
            resp = await self.client.get_states()
            for state_info in resp.result:
                entity_id = state_info['entity_id']
                cached_state = self._cached_states.get(entity_id)
                entity_spys = self._spys[entity_id]

                for spy in entity_spys:
                    if spy.attribute is None and cached_state['state'] !=


                print(state['entity_id'], state['state'])
                self._cached_states =

            await asyncio.sleep(60)

    def _handle_state_change(self, message: EventResponse):
        # message.event['entity_id']
        pass

    def on_state_change(self, handler, entity_id, attribute=None):
        self._spys[entity_id].append(StateSpy(
            entity_id=entity_id,
            attribute=attribute,
            handler=handler
        ))

    def listen_state(self):
        pass

    def get_state(self):
        pass

    def call_service(self):
        pass
