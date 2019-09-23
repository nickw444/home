# samsungwsctl

An alternative to the official [samsungtv](https://www.home-assistant.io/components/samsungtv/) component intended to 
support newer Samsung smart TVs using the alternative
[samsungwsctl](https://github.com/nickw444/samsungwsctl) client library for communication.

## Usage

```yaml
media_player:
  - platform: samsungwsctl
    host: tv.local
    port: 8002
    secure: true # (whether https & wss) should be used
    mac: ff:ff:ff:ff:ff:ff
    name: "Samsung Smart TV"
```

## Features

* Source list (with some predefined hardcoded apps):
    * Netflix
    * YouTube
    * Spotify
    * TV
* Register `media_player.send_key` service
* Report `off` state correctly when TV is in standby mode
* Correctly reports current app name & source by polling TV API

## `send_key` service:

**service**: `media_player.send_key` 

**payload**

```yaml
key: KEY_UP
```
