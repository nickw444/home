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

#define SONOFF_BUTTON    0
#define SONOFF_LED      13
#define EEPROM_SALT     1263

#define DHTPIN 14
#define DHTTYPE DHT21

// How often to transmit a reading in millis
#define READING_EVERY 1000 * 30

// TODO: Implement willmsg.
static SimpleMQTT mqtt(SONOFF_LED, EEPROM_SALT);
static Button button(SONOFF_BUTTON, false, true, 20);
static DHT dht(DHTPIN, DHTTYPE);
static Timer t;

void publishReading();
void republish(char * payload, unsigned int length);

void setup() {
  Serial.begin(115200);
  dht.begin();

  mqtt.subscribeTo("republish", republish);
  mqtt.beginConfig();

  t.every(READING_EVERY, publishReading);
}

void loop() {
  mqtt.tick();
  t.update();

  button.read();
  if (button.pressedFor(10000)) {
    Serial.println("Reset Settings");
    mqtt.reset();
  }
}

void republish(char * payload, unsigned int length) {
  publishReading();
}

void publishReading() {
  // Sensor readings may also be up to 2 seconds 'old' (its a very slow sensor)
  char buff[7];
  float h = dht.readHumidity();
  float t = dht.readTemperature();

  if (!isnan(h)) {
    dtostrf(h, -6, 2, buff);
    Serial.printf("Humidity: %s\n",  buff);
    mqtt.publish("humidity", buff);
  }

  if (!isnan(t)) {
    dtostrf(t, -6, 2, buff);
    Serial.printf("Temperature: %s\n",  buff);
    mqtt.publish("temperature", buff);
  }
}
