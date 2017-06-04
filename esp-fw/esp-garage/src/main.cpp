#include <ESP8266WiFi.h>
#include <ESP8266mDNS.h>
#include <WiFiManager.h>
#include <PubSubClient.h>
#include <Ticker.h>
#include <EEPROM.h>
#include <SimpleMQTT.h>
#include <Timer.h>
#include <Bounce2.h>

#include "data.h"

#define ESP_LED      13
#define EEPROM_SALT  5734

#define TRIGGER_GPIO  2 // (GPIO2)
#define SENSE_GPIO    3 // (RX)

#define DEBOUNCE_DELAY 100 // Debounce delay in millis.

// How often to transmit a status in millis
#define PUBLISH_EVERY 1000 * 30 // 30 S


static SimpleMQTT mqtt(ESP_LED, EEPROM_SALT);
static Timer t;
static Bounce b = Bounce();
static DOOR_STATE lastDoorState = DOOR_UNKNOWN;
static RELAY_STATE lastRelayState = RELAY_UNKNOWN;
static int nextRelayOffTime = -1;

void republish(char * payload, unsigned int length);
void trigger(char * payload, unsigned int length);
void unTrigger();
void publishDoorStatus();
void publishRelayStatus();
DOOR_STATE readDoor();

void setup() {
  Serial.begin(115200);

  pinMode(TRIGGER_GPIO, OUTPUT);
  pinMode(SENSE_GPIO, INPUT);
  digitalWrite(TRIGGER_GPIO, HIGH);

  mqtt.subscribeTo("trigger", trigger);
  mqtt.subscribeTo("republish", republish);

  mqtt.beginConfig();

  b.interval(DEBOUNCE_DELAY);
  b.attach(SENSE_GPIO);

  t.every(PUBLISH_EVERY, publishDoorStatus);
}

void loop() {
  mqtt.tick();
  t.update();
  b.update();

  DOOR_STATE currDoorStatus = readDoor();
  if (currDoorStatus != lastDoorState) {
    // Door status changed. We should notify.
    Serial.printf("Door status changed from %s to %s\n", serializeDoorState(lastDoorState), serializeDoorState(currDoorStatus));
    lastDoorState = currDoorStatus;
    publishDoorStatus();
  }

  if (nextRelayOffTime > 0 && millis() > nextRelayOffTime) {
    unTrigger();
  }
}


void trigger(char * payload, unsigned int length) {
  // Default to 500ms.
  int delayTime = 500;
  if (length >= 2 && payload[0] == 'T') {
    // Payload specifies a trigger length. We will use it.
    char safePayload[10]; // Allow up to 9999NUL.
    memcpy(safePayload, payload + 1, min(9, length - 1));
    safePayload[min(9, length - 1)] = 0; // Add NUL byte.
    int intVal = atoi(safePayload);
    Serial.printf("Converting String %s to Int %d\n", safePayload, intVal);
    delayTime = intVal;
  }

  Serial.printf("Trigger: Setting Low. Will delay for %d\n", delayTime);
  digitalWrite(TRIGGER_GPIO, LOW);
  nextRelayOffTime = millis() + delayTime;

  lastRelayState = RELAY_OPEN;
  publishRelayStatus();
}

void unTrigger() {
  Serial.println("Trigger: Setting High.");
  digitalWrite(TRIGGER_GPIO, HIGH);
  nextRelayOffTime = -1;
  lastRelayState = RELAY_CLOSED;
  publishRelayStatus();
}

void republish(char * payload, unsigned int length) {
  publishDoorStatus();
  publishRelayStatus();
}

void publishDoorStatus() {
  const char * doorState = serializeDoorState(lastDoorState);
  // Legacy support - publish on existing status topic.
  mqtt.publish("status", doorState);
}

void publishRelayStatus() {
  const char * relayState = serializeRelayState(lastRelayState);
  mqtt.publish("relayState", relayState);
}

DOOR_STATE readDoor() {
  int status = b.read();
  if (status == HIGH) {
    return DOOR_OPEN;
  } else {
    return DOOR_CLOSED;
  }
}
