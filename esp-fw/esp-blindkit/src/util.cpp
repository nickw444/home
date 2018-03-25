#include "util.h"
#include <ESP8266WiFi.h>

String get_device_id() {
  String clientMac = "";
  unsigned char mac[6];
  WiFi.macAddress(mac);

  for (int i = 0; i < 6; ++i) {
    clientMac += String(mac[i], 16);
  }
  return clientMac;
}
