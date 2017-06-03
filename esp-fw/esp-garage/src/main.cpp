#include <ESP8266WiFi.h>
#include <ESP8266mDNS.h>
#include <WiFiManager.h>
#include <PubSubClient.h>
#include <Ticker.h>
#include <EEPROM.h>
#include <SimpleMQTT.h>
#include <Timer.h>
#include <Bounce2.h>


#define ESP_LED      13
#define EEPROM_SALT  5734

#define TRIGGER_GPIO  2 // (GPIO2)
#define SENSE_GPIO    3 // (RX)

#define DEBOUNCE_DELAY 100 // Debounce delay in millis.

// How often to transmit a status in millis
#define PUBLISH_EVERY 1000 * 30 // 30 S

enum DOOR_STATE {
  DOOR_OPEN,
  DOOR_CLOSED,
  DOOR_UNKNOWN,
};

static SimpleMQTT mqtt(ESP_LED, EEPROM_SALT);
static DOOR_STATE lastDoorStatus = DOOR_UNKNOWN;
static Timer t;
static Bounce b = Bounce();

void onConnect();
void republish(char * payload, unsigned int length);
void trigger(char * payload, unsigned int length);
void publishStatus();
DOOR_STATE readDoor();
const char * serializeDoorState(DOOR_STATE state);

static int lastDebounceTime = 0;

void setup() {
  Serial.begin(115200);

  pinMode(TRIGGER_GPIO, OUTPUT);
  pinMode(SENSE_GPIO, INPUT);
  digitalWrite(TRIGGER_GPIO, HIGH);

  mqtt.subscribeTo("trigger", trigger);
  mqtt.subscribeTo("republish", republish);

  mqtt.onConnect(onConnect);
  mqtt.beginConfig();

  b.interval(DEBOUNCE_DELAY);
  b.attach(SENSE_GPIO);

  t.every(PUBLISH_EVERY, publishStatus);
}

void loop() {
  mqtt.tick();
  t.update();
  b.update();

  DOOR_STATE currDoorStatus = readDoor();
  if (currDoorStatus != lastDoorStatus) {
    // Door status changed. We should notify.
    Serial.printf("Door status changed from %s to %s\n", serializeDoorState(lastDoorStatus), serializeDoorState(currDoorStatus));
    lastDoorStatus = currDoorStatus;
    publishStatus();
  }
}

void onConnect() {
  Serial.println("Connected to MQTT.");
}

void trigger(char * payload, unsigned int length) {
  Serial.println("Trigger: Setting Low.");
  digitalWrite(TRIGGER_GPIO, LOW);
  delay(500);
  Serial.println("Trigger: Setting High.");
  digitalWrite(TRIGGER_GPIO, HIGH);
}

void republish(char * payload, unsigned int length) {
  publishStatus();
}

void publishStatus() {
  const char * status = serializeDoorState(lastDoorStatus);
  Serial.printf("Publishing status: %s\n", status);
  mqtt.publish("status", status);
}

DOOR_STATE readDoor() {
  int status = b.read();
  if (status == HIGH) {
    return DOOR_OPEN;
  } else {
    return DOOR_CLOSED;
  }
}

const char * serializeDoorState(DOOR_STATE state) {
  switch (state) {
    case DOOR_OPEN:
      return "OPEN";
    case DOOR_CLOSED:
      return "CLOSED";
    case DOOR_UNKNOWN:
      return "UNKNOWN";
  }
}
