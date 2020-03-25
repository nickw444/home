import io
from urllib.parse import urljoin
from urllib.request import urlopen

from appdaemon.plugins.hass import hassapi as hass
from colorthief import ColorThief

# The source the media player needs to be set to in order for the color to be set. Set
# this to your TV's device ID within Spotify (if using Spotify as your media player). Set
# to null if the color should always be set.
CONF_ENABLE_SOURCE = 'enable_source'
# The URL to access the Home Assistant instance. Required in order to download the
# artwork (entity_picture state attribute is a relative path)
CONF_HASS_BASE_URL = 'hass_base_url'
# The entity ID of the media player to obtain the artwork from (using
# entity_picture state attribute)
CONF_MEDIA_PLAYER_ENTITY = 'media_player_entity'
# The entity ID of the light to set the color
CONF_LIGHT_ENTITY = 'light_entity'


class MediaPlayerLight(hass.Hass):
    def initialize(self):
        self.listen_state(
            self.handle_update,
            self.args[CONF_MEDIA_PLAYER_ENTITY],
            attribute='entity_picture'
        )
        current_entity_picture = self.get_state(
            self.args[CONF_MEDIA_PLAYER_ENTITY],
            attribute='entity_picture'
        )
        self.on_entity_picture_changed(current_entity_picture)

    def handle_update(self, entity, attribute, old, new, kwargs):
        self.on_entity_picture_changed(new)

    def on_entity_picture_changed(self, entity_picture):
        enable_source = self.args[CONF_ENABLE_SOURCE]
        state = self.get_state(self.args[CONF_MEDIA_PLAYER_ENTITY])
        if state != 'playing':
            self.log(f'media player is not playing ({state}). Ignoring state change', level="INFO")
            return

        source = self.get_state(self.args[CONF_MEDIA_PLAYER_ENTITY], attribute="source")
        if enable_source is not None and source != enable_source:
            self.log(
                f'the current source ({source}) is not set to the enable source ({enable_source}). '
                f'Ignoring state change',
                level="INFO"
            )
            return

        artwork_url = self.get_fully_qualified_artwork_url(entity_picture)
        rgb_color = extract_dominant_color(artwork_url)
        self.log(
            f'media player entity picture changed. Dominant color is: {rgb_color}',
            level="INFO"
        )

        self.turn_on(self.args[CONF_LIGHT_ENTITY], rgb_color=rgb_color)

    def get_fully_qualified_artwork_url(self, image_path):
        return urljoin(self.args['hass_base_url'], image_path)


def extract_dominant_color(image_url):
    fd = urlopen(image_url)
    f = io.BytesIO(fd.read())
    color_thief = ColorThief(f)
    colors = color_thief.get_palette(3, quality=10)
    return colors[1]
