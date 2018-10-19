#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <ArduinoOTA.h>
#include <Bounce2.h>

#include "../../config.h"
#include "util.h"

#define STATE_GPIO 3 // RX
#define TRIGGER_GPIO 2 // GPIO2
#define STATE_DEBOUNCE_DELAY_MS 100
#define TRIGGER_DURATION_MS 1000

WiFiClient esp_client;
PubSubClient client(esp_client);
Bounce stateBounce = Bounce();

String deviceId = getDeviceId();
String triggerTopic = "/things/garage/" + deviceId + "/trigger";
String statusTopic = "/things/garage/" + deviceId + "/status";
String stateTopic = "/things/garage/" + deviceId + "/state";

static int nextRelayOffTime = 0;

void callback(char* topic, byte* payload, unsigned int length);

void setupWifi() {
  Serial.printf("Connecting to %s\n", WIFI_SSID);
  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  // TODO: Possibly handle ungraceful wifi disconnects.
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("");

  Serial.println("WiFi connected");
  Serial.print("IP address: ");
  Serial.println(WiFi.localIP());
}

void setupMqtt() {
  client.setServer(MQTT_SERVER, 1883);
  client.setCallback(callback);
}

void setupOta() {
  // Intentionally not worrying about username/password protection here –
  // devices are hosted on an isolated internal network, so less surface
  // area for an attacker to re-write the firmware.

  ArduinoOTA.onStart([]() {
    String type;
    if (ArduinoOTA.getCommand() == U_FLASH) {
      type = "sketch";
    } else { // U_SPIFFS
      type = "filesystem";
    }

    // NOTE: if updating SPIFFS this would be the place to unmount SPIFFS using SPIFFS.end()
    Serial.println("Start updating " + type);
  });
  ArduinoOTA.onEnd([]() {
    Serial.println("\nEnd");
  });
  ArduinoOTA.onProgress([](unsigned int progress, unsigned int total) {
    Serial.printf("Progress: %u%%\r", (progress / (total / 100)));
  });
  ArduinoOTA.onError([](ota_error_t error) {
    Serial.printf("Error[%u]: ", error);
    if (error == OTA_AUTH_ERROR) {
      Serial.println("Auth Failed");
    } else if (error == OTA_BEGIN_ERROR) {
      Serial.println("Begin Failed");
    } else if (error == OTA_CONNECT_ERROR) {
      Serial.println("Connect Failed");
    } else if (error == OTA_RECEIVE_ERROR) {
      Serial.println("Receive Failed");
    } else if (error == OTA_END_ERROR) {
      Serial.println("End Failed");
    }
  });
  ArduinoOTA.begin();
}

void callback(char* topic, byte* payload, unsigned int length) {
  if (triggerTopic.equals(topic)) {
    Serial.printf("Receieved message on trigger topic. Triggering for %d ms\n", TRIGGER_DURATION_MS);
    digitalWrite(TRIGGER_GPIO, LOW);
    nextRelayOffTime = millis() + TRIGGER_DURATION_MS;
  } else {
    Serial.printf("Received message on unknown topic: [%s]\n", topic);
  }
}

void mqttReconnect() {
  Serial.println("Attempting MQTT connection...");
  String clientId = "garage-" + deviceId;
  Serial.printf("Connecting with client id: %s\n", clientId.c_str());
  Serial.printf("Status topic: %s\n", statusTopic.c_str());

  if (client.connect(clientId.c_str(), statusTopic.c_str(), 1, true, "offline")) {
    Serial.println("Connected!");
    // Update device status (with retain)
    client.publish(statusTopic.c_str(), "online", true);
    // Subscribe to send_code topic
    client.subscribe(triggerTopic.c_str(), 1);
  } else {
    Serial.print("failed to connect, rc=");
    Serial.println(client.state());
  }
}

int lastConnectionAttempt;
int lastState = 99; // Use junk that is not HIGH or LOW to force update
                         // on boot

void setup() {
  Serial.begin(115200);
  setupWifi();
  setupMqtt();
  setupOta();

  // Setup Trigger GPIO
  pinMode(TRIGGER_GPIO, OUTPUT);
  digitalWrite(TRIGGER_GPIO, HIGH);

  // Setup the state input
  pinMode(STATE_GPIO, INPUT);
  stateBounce.attach(STATE_GPIO);
  stateBounce.interval(STATE_DEBOUNCE_DELAY_MS);

  lastConnectionAttempt = millis();
  mqttReconnect();
}

void loop() {
  int now = millis();

  stateBounce.update();

  if (!client.loop() && now - lastConnectionAttempt > MQTT_CONNECT_RETRY_MS) {
    lastConnectionAttempt = now;
    mqttReconnect();
  }

  int currState = stateBounce.read();
  if (lastState != currState) {
    const char * payload = currState ? "open" : "closed";
    Serial.printf("State changed - payload: %s\n", payload);
    client.publish(stateTopic.c_str(), payload, true);
    lastState = currState;
  }

  if (nextRelayOffTime > 0 && now > nextRelayOffTime) {
    digitalWrite(TRIGGER_GPIO, HIGH);
  }

  ArduinoOTA.handle();
}
