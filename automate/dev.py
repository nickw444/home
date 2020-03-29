import asyncio
import logging
import os

from automate.hass_client import HassClient
from automate.mate import Mate


class MediaPlayerLight:
    def __init__(self, mate: Mate, hass: HassClient, config):
        self.mate = mate
        self.hass = hass
        self.config = config

    def initialize(self):
        self.mate.on_state_change(
            self.handle_update,
            entity_id='light.led_strip',
        )

    def handle_update(self, entity_id, attribute, previous_state, state):
        pass


async def main():
    client = HassClient(
        host='wss://' + os.environ['HASS_HOST'],
        access_token=os.environ['HASS_ACCESS_TOKEN']
    )
    await client.connect()

    mate = Mate(client)
    app = MediaPlayerLight(mate, client, config={})

    # def handle_event(resp: EventResponse):
    #     print("New event:", resp.event.data['entity_id'], resp.event.data['new_state']['state'])
    #
    asyncio.ensure_future(mate.poll_states())
    await client.run_forever()

    # resp = await client.subscribe_events(handler=handle_event, event_type='state_changed')
    # print("subscribe resp: ", resp)


if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    logging.getLogger('automate.hass_client').setLevel(logging.DEBUG)

    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
    loop.close()
