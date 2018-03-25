#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <ArduinoOTA.h>

#include "../../config.h"
#include "transmit.h"
#include "util.h"

#define TX_PIN 3

WiFiClient espClient;
PubSubClient client(espClient);

String deviceId = get_device_id();
String sendCodeTopic = "/things/blindkit/" + deviceId + "/send_code";
String statusTopic = "/things/blindkit/" + deviceId + "/status";

void callback(char* topic, byte* payload, unsigned int length);


void setupWifi() {
  Serial.printf("Connecting to %s\n", WIFI_SSID);
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
  if (sendCodeTopic.equals(topic)) {
    Serial.printf("Send Code Topic!\n");

    char payloadStr[length + 1];
    strncpy(payloadStr, (char *)payload, length);
    payloadStr[length] = 0;

    uint16_t remote;
    uint8_t channel;
    raex_action_t action;
    int idx = 0;

    char * token = strtok(payloadStr, ",");
    while(token != NULL) {
      Serial.printf("Tok[%d]: [%s]\n", idx, token);
      if (idx == 0) {
        remote = atoi(token);
      } else if (idx == 1) {
        channel = atoi(token);
      } else if (idx == 2) {
        if (strcmp("UP", token) == 0) {
          action = TX_RAEX_ACTION_UP;
        } else if (strcmp("DOWN", token) == 0) {
          action = TX_RAEX_ACTION_DOWN;
        } else if (strcmp("STOP", token) == 0) {
          action = TX_RAEX_ACTION_STOP;
        } else if (strcmp("PAIR", token) == 0) {
          action = TX_RAEX_ACTION_PAIR;
        } else {
          Serial.printf("Malformed payload received. Unknown action [%s]\n", token);
          return;
        }
      } else {
        Serial.println("Malformed payload received.");
        return;
      }

      idx++;
      token = strtok(NULL, ",");
    }
    Serial.printf("Data - remote [%d], channel [%d], action [%d]\n", remote, channel, action);

    // TODO NW: Maybe implement send queue within the loop to avoid blocking
    // additional MQTT messages.
    txPrepare(TX_PIN, 200);
    txRaexSend(TX_PIN, remote, channel, action);
  } else {
    Serial.printf("Received message on unknown topic: [%s]\n", topic);
  }
}

void mqttReconnect() {
  Serial.println("Attempting MQTT connection...");
  String clientId = "blindkit-" + deviceId;
  Serial.printf("Connecting with client id: %s\n", clientId.c_str());
  Serial.printf("Status topic: %s\n", statusTopic.c_str());

  if (client.connect(clientId.c_str(), statusTopic.c_str(), 1, true, "offline")) {
    Serial.println("Connected!");
    // Update device status (with retain)
    client.publish(statusTopic.c_str(), "online", true);
    // Subscribe to send_code topic
    client.subscribe(sendCodeTopic.c_str(), 1);
  } else {
    Serial.print("failed to connect, rc=");
    Serial.println(client.state());
  }
}

int lastConnectionAttempt;

void setup() {
  Serial.begin(115200);
  setupWifi();
  setupMqtt();
  setupOta();

  lastConnectionAttempt = millis();
  mqttReconnect();
}

void loop() {
  int now = millis();

  if (!client.loop() && now - lastConnectionAttempt > MQTT_CONNECT_RETRY_MS) {
    lastConnectionAttempt = now;
    mqttReconnect();
  }

  ArduinoOTA.handle();
}
