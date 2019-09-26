# home-assistant

Configuration for my various home-assistant instances.

## Installation

Both of these configurations are deployed via the [hassio supervisor](https://github.com/home-assistant/hassio). hassio puts it's configuration at `/usr/share/hassio` and within it, it has the `homeassistant` configuration directory.

In a perfect world, a symbolic link from `/usr/share/hassio/homeassistant` to `/path/to/this/repo/home-assistant/313a` would be fine, however in the hassio supervisor container it expects to find `/config/homeassistant`. This means it stumbles upon a symbolic link which it cannot resolve from within it's own world.

To resolve this, we can mount the config directory with `bindfs`:

```sh
sudo mount --bind /usr/share/nickw444_home/home-assistant/martin-pl /usr/share/hassio/homeassistant
```

And to make permanent, can add the following to `/etc/fstab`:

```
/usr/share/nickw444_home/home-assistant/martin-pl   /usr/share/hassio/homeassistant     none    bind                      0       0
```
