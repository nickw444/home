#include <ESP8266WiFi.h>
#include <ESP8266mDNS.h>
#include <WiFiManager.h>
#include <PubSubClient.h>
#include <Ticker.h>
#include <EEPROM.h>
#include "DHT.h"
#include <SimpleMQTT.h>
#include <Timer.h>
#include <Button.h>
#include <RAEXRemote.h>

#define LED_PIN         13
#define EEPROM_SALT     532
#define TX_PIN          3

static SimpleMQTT mqtt(LED_PIN, EEPROM_SALT);
static Transmitter transmitter = Transmitter(TX_PIN);
static Scheduler scheduler = Scheduler();
static RAEXRemote raexRemote = RAEXRemote(&scheduler, &transmitter);

void controlRaex(char * payload, unsigned int length);

void setup() {
  Serial.begin(115200, SERIAL_8N1, SERIAL_TX_ONLY);
  pinMode(TX_PIN, OUTPUT);

  mqtt.subscribeTo("raex", controlRaex);
  mqtt.beginConfig();
}

void controlRaex(char * payload, unsigned int length) {
  // Expect payload in the format: channel:remote:action:.
  // i.e. 53:64:127:
  int vals[3];
  char buf[10];
  int channel;
  int remote;
  int action;
  size_t vals_pos = 0;
  size_t buf_pos = 0;

  for (size_t i = 0; i < length && vals_pos < 3; i++) {
    char ch = payload[i];
    if (ch == ':') {
      buf[buf_pos] = 0;
      vals[vals_pos] = atoi(buf);

      buf_pos = 0;
      vals_pos++;
    } else if (ch >= '0' && ch <= '9') {
      buf[buf_pos++] = ch;
    } else {
      break;
    }
  }

  if (vals_pos < 3) {
    Serial.println("Bad payload received");
    return;
  }

  channel = vals[0];
  remote = vals[1];
  action = vals[2];

  Serial.print("Channel: ");
  Serial.print(channel);
  Serial.print(", Remote: ");
  Serial.print(remote);
  Serial.print(", Action: ");
  Serial.println(action);

  RAEXRemoteCode raexRemoteCode = RAEXRemoteCode(channel, remote, action);
  raexRemote.transmitCode(&raexRemoteCode);
}

void loop() {
  // Serial.println("MEMes");
  // delay(1000);
  mqtt.tick();
}
