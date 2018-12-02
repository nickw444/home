#include <Arduino.h>
#include <PubSubClient.h>
#include <Ethernet.h>
#include <avr/wdt.h>


#include "../../config.h"

#define RELAY_UPDATE_INTERVAL 1000
// Re-publish every 5 minutes
#define PUBLISH_INTERVAL 300000

#define STATUS_PAYLOAD_ONLINE "online"
#define STATUS_PAYLOAD_OFFLINE "offline"

uint8_t MAC_ADDRESS[6] = {0x12, 0xe0, 0x14, 0xec, 0x13, 0x82};
String clientId = "sprinkle";
String statusTopic = clientId + "/status";
String uptimeTopic = clientId + "/uptime";
String ipTopic = clientId + "/ip";
String setTopic = clientId + "/+/set";

EthernetClient ethClient;
PubSubClient mqttClient;

enum outputState {
  OUTPUT_STATE_ON,
  OUTPUT_STATE_OFF,
  OUTPUT_STATE_TIMER,
};

struct output {
  uint8_t id;
  uint8_t pin;
  outputState state;
  unsigned long offTime;
};

static uint8_t AVAILABLE_OUTPUTS[] = {
  2, 3, 4, 5, 6, 7, 8, 9,
  // D10, D11, D12, D13 reserved for ethernet
  A0, A1, A2, A3, A4, A5, A6, A7,
};

static const uint8_t NUM_OUTPUTS = sizeof(AVAILABLE_OUTPUTS);
static struct output outputs[NUM_OUTPUTS];

bool isOnState(outputState state) {
  return state == OUTPUT_STATE_ON || state == OUTPUT_STATE_TIMER;
}

void publishOutput(output *o) {
  char topic[20];
  snprintf(topic, sizeof(topic), "%s/%d", clientId.c_str(), o->id);
  mqttClient.publish(topic, isOnState(o->state) ? "ON" : "OFF", true);
}

void publish() {
  char buf[16];

  Serial.println(F("Publishing state"));
  mqttClient.publish(statusTopic.c_str(), STATUS_PAYLOAD_ONLINE, true);

  // Publish Uptime
  snprintf(buf, sizeof(buf), "%lu", millis());
  mqttClient.publish(uptimeTopic.c_str(), buf, true);

  // Publish IP
  IPAddress ip = Ethernet.localIP();
  snprintf(buf, sizeof(buf), "%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3]);
  mqttClient.publish(ipTopic.c_str(), buf, true);

  for (size_t i = 0; i < NUM_OUTPUTS; i++) {
    output *o = &outputs[i];
    publishOutput(o);
  }
}

void writeOutput(output *o) {
  digitalWrite(o->pin, isOnState(o->state) ? LOW : HIGH);
}


void callback(char* topic, byte* payload, unsigned int length) {
  unsigned int outputId;
  sscanf(topic + clientId.length() + 1, "%u/set", &outputId);

  output *o;
  for (size_t i = 0; i < NUM_OUTPUTS; i++) {
    if (outputs[i].id == outputId) {
      o = &outputs[i];
      break;
    }
  }

  if (o == NULL) {
    Serial.println(F("Invalid output."));
    return;
  }

  char buf[10];
  uint8_t end = min(sizeof(buf) - 1, length);
  memcpy(buf, payload, end);
  buf[end] = 0;

  if (strncmp(buf, "OFF", 3) == 0) {
    o->offTime = 0;
    o->state = OUTPUT_STATE_OFF;
  } else if (strncmp(buf, "ON", 2) == 0) {
    unsigned long duration = 0;
    if(sscanf(buf, "ON:%lu", &duration) == 1) {
      o->offTime = millis() + duration * 1000;
      o->state = OUTPUT_STATE_TIMER;
    } else {
      o->offTime = 0;
      o->state = OUTPUT_STATE_ON;
    }
  } else {
    Serial.println(F("Invalid payload."));
    return;
  }

  writeOutput(o);
  publishOutput(o);
}

void mqttReconnect() {
  Serial.println(F("Attempting MQTT connection..."));
  if (mqttClient.connect(clientId.c_str(), statusTopic.c_str(), 1, true, STATUS_PAYLOAD_OFFLINE)) {
    Serial.println(F("MQTT Connected!"));
    // Update device status (with retain)
    mqttClient.publish(statusTopic.c_str(), STATUS_PAYLOAD_ONLINE, true);
    // Subscribe to all set topics.
    mqttClient.subscribe(setTopic.c_str(), 1);
  } else {
    Serial.print(F("failed to connect, rc="));
    Serial.println(mqttClient.state());
  }
}

/** Find expired timers and update the output state */
void updateOutputs(unsigned long now) {
  for (size_t i = 0; i < NUM_OUTPUTS; i++) {
    output *o = &outputs[i];
    if (o->state == OUTPUT_STATE_TIMER && o->offTime < now) {
      Serial.print(F("Timer for output "));
      Serial.print(o->id);
      Serial.println(F(" has expired"));
      o->state = OUTPUT_STATE_OFF;
      o->offTime = 0;
      publishOutput(o);
      writeOutput(o);
    }
  }
}

void initOutputs() {
  Serial.println(F("Initializing Outputs.."));

  for (size_t i = 0; i < NUM_OUTPUTS; i++) {
    output *o = &outputs[i];
    o->pin = AVAILABLE_OUTPUTS[i];
    o->id = i + 1;
    o->offTime = 0;
    o->state = OUTPUT_STATE_OFF;
    pinMode(o->pin, OUTPUT);
    writeOutput(o);
  }
}

void initNetwork() {
  Serial.println(F("Ethernet connecting.."));
  while (Ethernet.begin(MAC_ADDRESS) == 0) {
    wdt_reset();
    Serial.println(F("DHCP failed. Retrying.."));
  }
  Serial.print(F("Ethernet connected. IP: "));
  Serial.println(Ethernet.localIP());
}

void initMqtt() {
  mqttClient.setClient(ethClient);
  mqttClient.setServer(MQTT_SERVER, 1883);
  mqttClient.setCallback(callback);
}

unsigned long lastConnectionAttempt;
unsigned long lastRelayUpdateTime;
unsigned long lastPublishTime;

void setup() {
  wdt_disable();
  // Force a pause to allow an IDE to reconnect/flash.
  delay(2L * 1000L);
  wdt_enable(WDTO_8S);

  Serial.begin(115200);

  initOutputs();

  initNetwork();
  initMqtt();

  lastConnectionAttempt = millis();
  lastRelayUpdateTime = millis();
  // Request a publish immediately on connect.
  lastPublishTime = 0;

  wdt_reset();
  mqttReconnect();
}

void loop() {
  wdt_reset();
  
  unsigned long now = millis();

  if (!mqttClient.loop() && now - lastConnectionAttempt > MQTT_CONNECT_RETRY_MS) {
    lastConnectionAttempt = now;
    mqttReconnect();
  }

  if (now - lastRelayUpdateTime > RELAY_UPDATE_INTERVAL) {
    lastRelayUpdateTime = now;
    updateOutputs(now);
  }

  if ((now - lastPublishTime > PUBLISH_INTERVAL || lastPublishTime == 0) && mqttClient.connected()) {
    lastPublishTime = now;
    publish();
  }
}
