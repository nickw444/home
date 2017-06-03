#ifndef SIMPLEMQTT_H_
#define SIMPLEMQTT_H_

#include <Ticker.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <WiFiManager.h>
#include <EEPROM.h>
#include <Arduino.h>

#define TOPIC_REBOOT  "reboot"
#define TOPIC_RESET   "reset"

#define SIMPLEMQTT_CALLBACK_SIGNATURE std::function<void(char *, unsigned int)>
#define ON_CONNECT_SIGNATURE std::function<void(void)>

class Subscription {
  public:
    SIMPLEMQTT_CALLBACK_SIGNATURE cb;
    String topic;
    Subscription *next;
};

struct WMSettings {
  uint16_t eepromSalt = 0x00;
  char mqttAddress[30] = "";
  char mqttUser[17] = "";
  char mqttPassword[17] = "";
  int mqttPort = 8883;
};

class SimpleMQTT {
  public:
    SimpleMQTT(uint8_t ledPin, uint16_t eeprom_salt);
    SimpleMQTT(uint8_t ledPin, uint16_t eeprom_salt, String willTopic, const char * willMsg);
    void beginConfig();
    void tick();

    void onConnect(ON_CONNECT_SIGNATURE callback);

    void subscribeTo(String topic, SIMPLEMQTT_CALLBACK_SIGNATURE callback);
    void publish(String topic, const char * data);

    void reboot();
    void reset();

    static String getPlainMac(void);
    String hostname;
    String macAddress;

  private:
    Ticker ticker;

    WiFiClientSecure espClient;
    PubSubClient* client;

    uint8_t ledPin;
    uint16_t eepromSalt;
    String willTopic;
    const char * willMsg;

    bool shouldSaveConfig = false;

    Subscription * subscriptions;
    struct WMSettings settings;

    ON_CONNECT_SIGNATURE onConnectCallback;

    void mqttReconnect();

    void mqttCallback(char * topic, byte * payload, unsigned int length);
    static void _mqttCallback(char * topic, byte * payload, unsigned int length);

    void tickLED();
    static void _tickLED();

    void onEnterConfigMode(WiFiManager *wifi);
    static void _onEnterConfigMode(WiFiManager *wifi);

    void onSaveConfig();
    static void _onSaveConfig();

    void init(uint8_t ledPin, uint16_t eepromSalt);
    String makeTopicString(String topic);
};


#endif /* SIMPLEMQTT_H_ */
