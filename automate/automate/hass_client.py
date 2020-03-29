import asyncio
import json
import logging
from asyncio import Future
from typing import Optional, Dict, Callable, Union
from urllib.parse import urljoin

import websockets

from .event_with_value import EventWithValue
from .hass_response import Response, PongResponse, ResultResponse, EventResponse

_LOGGER = logging.getLogger(__name__)


class HassClient:
    def __init__(self, host: str, access_token: str):
        self._host = host
        self._access_token = access_token

        self._conn: Optional[websockets.WebSocketClientProtocol] = None
        self._listen_task: Optional[Future] = None
        self._keepalive_task: Optional[Future] = None

        self._last_req_id = 0
        self._subscriptions: Dict[int, Callable] = {}
        self._pending_events: Dict[int, EventWithValue[Response]] = {}

    async def connect(self):
        self._conn = await websockets.connect(urljoin(self._host, '/api/websocket'))
        await self.authenticate()

        # Begin listening
        self._listen_task = asyncio.ensure_future(self._listen())
        self._keepalive_task = asyncio.ensure_future(self._keepalive())

    async def authenticate(self):
        while True:
            auth_resp = json.loads(await self._conn.recv())
            if auth_resp['type'] == 'auth_required':
                await self._conn.send(json.dumps({
                    'type': 'auth',
                    'access_token': self._access_token,
                }))
            elif auth_resp['type'] == 'auth_ok':
                return
            elif auth_resp['type'] == 'auth_invalid':
                raise ValueError('Invalid credentials: ' + auth_resp)
            else:
                raise AssertionError('Unexpected response: ' + auth_resp)

    async def run_forever(self):
        return await asyncio.wait([self._listen_task, self._keepalive_task])

    async def _listen(self):
        async for raw_message in self._conn:
            response = Response.deserialize(json.loads(raw_message))
            if isinstance(response, EventResponse):
                handler = self._subscriptions[response.id]
                try:
                    handler(response)
                except:
                    _LOGGER.error('Exception whilst calling event handler', exc_info=True)
            elif isinstance(response, PongResponse) or isinstance(response, ResultResponse):
                e = self._pending_events[response.id]
                del self._pending_events[response.id]
                e.set(response)
            else:
                raise AssertionError('Unexpected response: ', response)

    async def _keepalive(self):
        while True:
            resp = await self.ping()
            _LOGGER.debug("Got ping response: %s", resp)
            await asyncio.sleep(5)

    def _get_request_id(self):
        self._last_req_id += 1
        return self._last_req_id

    async def _send(self, data) -> Union[ResultResponse, PongResponse]:
        req_id = self._get_request_id()

        await self._conn.send(json.dumps({
            'id': req_id,
            **data
        }))

        ev = EventWithValue()
        self._pending_events[req_id] = ev

        resp = await ev.wait()
        if isinstance(resp, ResultResponse) and resp.error is not None:
            raise Exception(resp.error.message)

        return resp

    async def subscribe_events(self, *, handler: Callable, event_type: str = None) -> ResultResponse:
        data = {}
        if event_type is not None:
            data['event_type'] = event_type

        resp = await self._send({
            'type': 'subscribe_events',
            **data,
        })
        self._subscriptions[resp.id] = handler
        return resp

    async def unsubscribe_events(self, *, subscription: int) -> ResultResponse:
        resp = await self._send({
            'type': 'unsubscribe_events',
            'subscription': subscription
        })
        del self._subscriptions[subscription]
        return resp

    async def call_service(self, *, domain: str, service: str, service_data=None,
                           entity_id=None) -> ResultResponse:
        if entity_id is not None and service_data is not None:
            raise ValueError('Cannot provide both service_data and entity_id')

        if entity_id is not None:
            service_data = {
                'entity_id': entity_id
            }

        return await self._send({
            'type': 'call_service',
            'domain': domain,
            'service': service,
            'service_data': service_data,
        })

    async def get_states(self) -> ResultResponse:
        return await self._send({'type': 'get_states'})

    async def get_config(self) -> ResultResponse:
        return await self._send({'type': 'get_config'})

    async def get_services(self) -> ResultResponse:
        return await self._send({'type': 'get_services'})

    async def get_panels(self) -> ResultResponse:
        return await self._send({'type': 'get_panels'})

    async def camera_thumbnail(self, *, entity_id: str) -> ResultResponse:
        return await self._send({
            'type': 'camera_thumbnail',
            'entity_id': entity_id,
        })

    async def media_player_thumbnail(self, *, entity_id: str) -> ResultResponse:
        return await self._send({
            'type': 'media_player_thumbnail',
            'entity_id': entity_id,
        })

    async def ping(self) -> PongResponse:
        return await self._send({'type': 'ping'})
