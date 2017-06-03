#include "SimpleMQTT.h"

// We need to keep reference to a global in order to pass C function pointers
// around the place. TODO: Work out how to do this without a global + static
// methods.
static SimpleMQTT *g_SimpleMQTTInstance;

SimpleMQTT::SimpleMQTT(uint8_t ledPin, uint16_t eepromSalt) {
  init(ledPin, eepromSalt);
}

SimpleMQTT::SimpleMQTT(uint8_t ledPin, uint16_t eepromSalt, String willTopic, const char * willMsg) {
  init(ledPin, eepromSalt);

  this->willTopic = willTopic;
  this->willMsg = willMsg;
}

void SimpleMQTT::init(uint8_t ledPin,  uint16_t eepromSalt) {
  macAddress = getPlainMac();
  hostname = "esp-" + macAddress;
  this->client = new PubSubClient(espClient);
  this->ledPin = ledPin;
  this->eepromSalt = eepromSalt;

  g_SimpleMQTTInstance = this;
}

String SimpleMQTT::makeTopicString(String topic) {
  return "esp/" + macAddress + "/" + topic;
}

void SimpleMQTT::beginConfig() {
  pinMode(ledPin, OUTPUT);

  WiFiManager wifiManager;
  wifiManager.setAPCallback(SimpleMQTT::_onEnterConfigMode);
  wifiManager.setConfigPortalTimeout(180); //Reboot if it's not configured.

  // Handle Config Params
  EEPROM.begin(512);
  EEPROM.get(0, settings);
  EEPROM.end();

  if (settings.eepromSalt != eepromSalt) {
    Serial.println("Invalid settings in EEPROM, trying with defaults");
    WMSettings defaults;
    settings = defaults;
    settings.eepromSalt = eepromSalt;
    WiFi.disconnect();
  }

  WiFiManagerParameter mqttServerAddress("mqtt-server-address", "MQTT Server Address", settings.mqttAddress, 30);
  WiFiManagerParameter mqttServerPort("mqtt-server-port", "MQTT Server Port", String(settings.mqttPort).c_str(), 6);
  WiFiManagerParameter mqttUsername("mqtt-username", "MQTT User", settings.mqttUser, 16);
  WiFiManagerParameter mqttPassword("mqtt-password", "MQTT Password", settings.mqttPassword, 16);
  wifiManager.addParameter(&mqttServerAddress);
  wifiManager.addParameter(&mqttServerPort);
  wifiManager.addParameter(&mqttUsername);
  wifiManager.addParameter(&mqttPassword);

  //set config save notify callback
  wifiManager.setSaveConfigCallback(SimpleMQTT::_onSaveConfig);

  if (!wifiManager.autoConnect(hostname.c_str())) {
    Serial.println("failed to connect and hit timeout");
    reboot();
  }

  if (shouldSaveConfig) {
    Serial.println("Saving config");

    strcpy(settings.mqttAddress, mqttServerAddress.getValue());
    strcpy(settings.mqttUser, mqttUsername.getValue());
    strcpy(settings.mqttPassword, mqttPassword.getValue());
    settings.mqttPort = atoi(mqttServerPort.getValue());

    EEPROM.begin(512);
    EEPROM.put(0, settings);
    EEPROM.end();
  }

  ticker.detach(); // Stop Blinking LED
  Serial.println("Device is started...");
  Serial.printf("settings.mqttAddress: '%s'\n", settings.mqttAddress);
  Serial.printf("settings.mqttPort: '%d'\n", settings.mqttPort);
  Serial.printf("settings.mqttUser: '%s'\n", settings.mqttUser);
  Serial.printf("settings.mqttPassword: '%s'\n", settings.mqttPassword);


  subscribeTo(TOPIC_REBOOT, std::bind(&SimpleMQTT::reboot, this));
  subscribeTo(TOPIC_RESET, std::bind(&SimpleMQTT::reset, this));

  client->setServer(settings.mqttAddress, settings.mqttPort);
  client->setCallback(SimpleMQTT::_mqttCallback);
}

void SimpleMQTT::tick() {
  if (!client->connected()) {
    mqttReconnect();
  }
  client->loop();
}

void SimpleMQTT::subscribeTo(String topic, SIMPLEMQTT_CALLBACK_SIGNATURE callback) {
  Subscription *sub = new Subscription;
  sub->cb = callback;
  sub->topic = makeTopicString(topic);
  sub->next = subscriptions;
  subscriptions = sub;
}

void SimpleMQTT::onConnect(ON_CONNECT_SIGNATURE fn) {
  onConnectCallback = fn;
}


void SimpleMQTT::reboot() {
  ESP.reset();
  delay(2000);
}

void SimpleMQTT::reset() {
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

String SimpleMQTT::getPlainMac(void) {
  byte mac[6];
  WiFi.macAddress(mac);
  String sMac = "";
  for (int i = 0; i < 6; ++i) {
    sMac += String(mac[i], HEX);
  }
  return sMac;
}


void SimpleMQTT::publish(String topic, const char * data) {
  if (topic != NULL && data != NULL) {
      client->publish(makeTopicString(topic).c_str(), data);
  }
}

void SimpleMQTT::onEnterConfigMode(WiFiManager *wifi) {
  Serial.println("Entered config mode");
  Serial.println(WiFi.softAPIP());
  //if you used auto generated SSID, print it
  Serial.println(wifi->getConfigPortalSSID());
  //entered config mode, make led toggle faster
  g_SimpleMQTTInstance->ticker.attach(0.2, SimpleMQTT::_tickLED);
}

void SimpleMQTT::tickLED() {
  //toggle state
  int state = digitalRead(g_SimpleMQTTInstance->ledPin);  // get the current state of GPIO1 pin
  digitalWrite(g_SimpleMQTTInstance->ledPin, !state);     // set pin to the opposite state
}

void SimpleMQTT::onSaveConfig() {
  Serial.println("Should save config");
  g_SimpleMQTTInstance->shouldSaveConfig = true;
}

void SimpleMQTT::mqttReconnect() {

  while(!client->connected()) {
    Serial.println("Attempting MQTT connection...");
    // Attempt to connect. We will setup a will topic publish so that when
    // the device disconnects, it will set it's state to off.
    bool result;
    if (willTopic != NULL && willMsg != NULL) {
      result = client->connect(hostname.c_str(), settings.mqttUser, settings.mqttPassword,
                               willTopic.c_str(), 0, false, (char *)willMsg);
    } else {
      result = client->connect(hostname.c_str(), settings.mqttUser, settings.mqttPassword);
    }

    if (result) {
      Serial.println("Connected to MQTT");

      for(Subscription *curr = subscriptions; curr != NULL; curr = curr->next) {
        Serial.println("Subscribed to topic: " + curr->topic);
        client->subscribe(curr->topic.c_str());
      }
      Serial.println("Subscribed to topics");

      if (onConnectCallback != NULL) {
        Serial.println("Executing on-connect callback");
        onConnectCallback();
      }
    } else {
      Serial.print("failed, rc=");
      Serial.print(client->state());
      Serial.println(" try again in 1 seconds");
      // Wait 5 seconds before retrying
      delay(1000);
    }
  }
  Serial.println("Connected!");
}

void SimpleMQTT::mqttCallback(char *topic, byte *payload, unsigned int length) {
  Serial.printf("Message arrived [%s]\n", topic);

  for(Subscription *curr = subscriptions; curr != NULL; curr = curr->next) {
    if (curr->topic.compareTo(topic) == 0) {
      Serial.println("Matching subscribed topic string: " + curr->topic);
      curr->cb((char *)payload, length);
      return;
    }
  }
  Serial.println("Topic does not have a handler");
}

void SimpleMQTT::_tickLED() {
  g_SimpleMQTTInstance->tickLED();
}

void SimpleMQTT::_onEnterConfigMode(WiFiManager *wifi) {
  g_SimpleMQTTInstance->onEnterConfigMode(wifi);
}

void SimpleMQTT::_onSaveConfig() {
  g_SimpleMQTTInstance->onSaveConfig();
}

void SimpleMQTT::_mqttCallback(char *topic, byte *payload, unsigned int length) {
  g_SimpleMQTTInstance->mqttCallback(topic, payload, length);
}
