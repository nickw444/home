#include <ESP8266WiFi.h>
#include <ESP8266mDNS.h>
#include <WiFiManager.h>
#include <PubSubClient.h>
#include <Ticker.h>
#include <EEPROM.h>
#include <SimpleMQTT.h>
#include <Timer.h>


#define SONOFF_BUTTON    0
#define SONOFF_LED      13
#define EEPROM_SALT     1263
#define SONOFF_RELAY    12


static SimpleMQTT mqtt(SONOFF_BUTTON, SONOFF_LED, EEPROM_SALT);

void setup() {
  Serial.begin(115200);
  mqtt.subscribeTo("republish", republish);
  mqtt.subscribeTo("relay/set", setRelay);

  // TODO: onbuttonpress

  mqtt.beginConfig();

  t.every(READING_EVERY, publishReading);
}

void loop() {
  mqtt.tick();
}

void publishReading() {
  client.publish(topicRelayState.c_str(), currentState == RELAY_STATE_ON ? "1" : "0");
}

void republish(char * payload, unsigned int length) {
  publishReading();
}

void setRelay(char * payload, unsigned int length) {
  // TODO
}

void ledTick();
void setState(enum relayState s);
void setState(enum relayState s, bool notify);
void onEnterConfigMode (WiFiManager *wifi);
void onSaveConfig();
void toggle();
void reboot();
void reset();
String getPlainMac(void);
void makeTopicStrings();
void notifyState();

void mqttReconnect();
void mqttCallback(char* topic, byte* payload, unsigned int length);

void loop() {
  if (!client.connected()) {
    mqttReconnect();
  }

  client.loop();
  button.read();

  if (button.pressedFor(10000)) {
    Serial.println("Reset Settings");
    reset();
  } else if (button.wasReleased()) {
    Serial.println("Toggle Relay");
    toggle();
  }
}

void setState(enum relayState s, bool notify) {
  Serial.printf("Relay State Is %s\n", s == RELAY_STATE_ON ? "On" : "Off");
  currentState = s;
  digitalWrite(SONOFF_RELAY, s);
  digitalWrite(SONOFF_LED, (s + 1) % 2); // led is active low

  if (notify) {
    notifyState();
  }
}

void setState(enum relayState s) {
  setState(s, true);
}


void turnOn() {
  setState(RELAY_STATE_ON);
}

void turnOff() {
  setState(RELAY_STATE_OFF);
}

void toggle() {
  setState(currentState == RELAY_STATE_ON ? RELAY_STATE_OFF : RELAY_STATE_ON);
}

void reboot() {
  ESP.reset();
  delay(2000);
}

void reset() {

  WMSettings defaults;
  settings = defaults;
  EEPROM.begin(512);
  EEPROM.put(0, settings);
  EEPROM.end();

  WiFi.disconnect();
  delay(1000);
  ESP.reset();
  delay(2000);
}


void onEnterConfigMode (WiFiManager *wifi) {
  Serial.println("Entered config mode");
  Serial.println(WiFi.softAPIP());
  //if you used auto generated SSID, print it
  Serial.println(wifi->getConfigPortalSSID());
  //entered config mode, make led toggle faster
  ticker.attach(0.2, ledTick);
}

void onSaveConfig() {
  Serial.println("Should save config");
  shouldSaveConfig = true;
}

void mqttReconnect() {
  Serial.println("Attempting MQTT connection...");
  // Create a random client ID
  String clientId = "esp-";
  clientId += getPlainMac();

  // Attempt to connect. We will setup a will topic publish so that when
  // the device disconnects, it will set it's state to off.
  if (client.connect(clientId.c_str(), settings.mqttUser, settings.mqttPassword, topicRelayState.c_str(), 0, false, "0")) {
    Serial.println("Connected to MQTT");

    client.subscribe(topicReboot.c_str());
    client.subscribe(topicRelaySet.c_str());
    client.subscribe(topicRepublish.c_str());
    client.subscribe(topicReset.c_str());
    Serial.println("Subscribed to topics");
    notifyState();
    Serial.println("Notified of current state");

  } else {
    Serial.print("failed, rc=");
    Serial.print(client.state());
    Serial.println(" try again in 5 seconds");
    // Wait 5 seconds before retrying
    delay(5000);
  }
}

void mqttCallback(char* topic, byte* payload, unsigned int length) {
  Serial.printf("Message arrived [%s]\n", topic);

  if(strcmp(topic, topicReboot.c_str()) == 0) {
    Serial.println("Reboot was requested.");
    reboot();
  } else if (strcmp(topic, topicRelaySet.c_str()) == 0) {
    if (payload[0] == '1') {
      Serial.println("Turning on.");
      turnOn();
    } else if (payload[0] == '0') {
      Serial.println("Turning off.");
      turnOff();
    } else {
      Serial.println("Invalid payload provided.");
    }
  } else if (strcmp(topic, topicRepublish.c_str()) == 0) {
    Serial.println("Republish was requested.");
    notifyState();
  } else if (strcmp(topic, topicReset.c_str()) == 0) {
    Serial.println("Reset was requested.");
    reset();
  }
}


String getPlainMac(void) {
  byte mac[6];
  WiFi.macAddress(mac);
  String sMac = "";
  for (int i = 0; i < 6; ++i) {
    sMac += String(mac[i], HEX);
  }
  return sMac;
}

void makeTopicStrings() {
  String macAddress = getPlainMac();
  topicRelayState = "device/" + macAddress + "/relay";
  topicRelaySet = "device/" + macAddress + "/relay/set";
  topicReboot = "device/" + macAddress + "/reboot";
  topicRepublish = "device/" + macAddress + "/republish";
  topicReset = "device/" + macAddress + "/reset";
}

void notifyState() {
  client.publish(topicRelayState.c_str(), currentState == RELAY_STATE_ON ? "1" : "0");
}
