import mock
import pytest
from appdaemon_testing.pytest import automation_fixture

from apps.media_player_light import (
    MediaPlayerLight,
    CONF_MEDIA_PLAYER_ENTITY,
    CONF_LIGHT_ENTITY,
    CONF_HASS_BASE_URL,
    CONF_ENABLE_SOURCE,
)

ENABLE_SOURCE = "Spotify"
MEDIA_PLAYER_ENTITY = "media_player.spotify"
LIGHT_ENTITY = "light.led_strip"


def test_callbacks_are_registered(hass_driver, media_player_light: MediaPlayerLight):
    listen_state = hass_driver.get_mock("listen_state")
    assert listen_state.call_count == 1
    listen_state.assert_called_once_with(
        media_player_light.handle_update,
        MEDIA_PLAYER_ENTITY,
        attribute="entity_picture",
    )


def test_picture_changed_and_playing_set_color(
    hass_driver, media_player_light: MediaPlayerLight, mock_extract_color
):
    with hass_driver.setup():
        hass_driver.set_state(MEDIA_PLAYER_ENTITY, "playing")
        hass_driver.set_state(
            MEDIA_PLAYER_ENTITY, ENABLE_SOURCE, attribute_name="source"
        )

    hass_driver.set_state(
        MEDIA_PLAYER_ENTITY, "/album.jpg", attribute_name="entity_picture"
    )

    turn_on = hass_driver.get_mock("turn_on")
    turn_on.assert_called_once_with("light.led_strip", rgb_color=(11, 22, 33))


@pytest.fixture()
def mock_extract_color():
    with mock.patch("apps.media_player_light.extract_dominant_color") as fn:
        fn.return_value = (11, 22, 33)
        yield fn


@automation_fixture(
    MediaPlayerLight,
    args={
        CONF_ENABLE_SOURCE: ENABLE_SOURCE,
        CONF_HASS_BASE_URL: "https://hass.local/",
        CONF_MEDIA_PLAYER_ENTITY: MEDIA_PLAYER_ENTITY,
        CONF_LIGHT_ENTITY: LIGHT_ENTITY,
    },
)
def media_player_light():
    pass
