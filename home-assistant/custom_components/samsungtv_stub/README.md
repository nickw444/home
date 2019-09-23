# samsungtv_stub

A stub custom component to force Home Assistant to load [a fork](https://github.com/eclair4151/samsungctl/tree/websocketssl) 
of samsungctl that is compatible with the Samsung RU8000 series TV.

Unfortunately the underlying samsungctl library incorrectly reports that the tv is "on" when it is in standby mode 
(connected to network but offline). For this reason, I have developed the [samsungwsctl component](../samsungwsctl) 
and associated [samsungwsctl client library](https://github.com/nickw444/samsungwsctl)

## Usage

```yaml
samsungctl_stub:
media_player:
  - platform: samsungtv
    host: tv.local
    port: 8002
    mac: ff:ff:ff:ff:ff:ff
```
