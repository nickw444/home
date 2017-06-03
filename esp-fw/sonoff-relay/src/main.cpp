#include <ESP8266WiFi.h>
#include <ESP8266mDNS.h>
#include <WiFiManager.h>
#include <PubSubClient.h>
#include <Ticker.h>
#include <EEPROM.h>
#include <SimpleMQTT.h>
#include <Timer.h>
#include <Button.h>


#define SONOFF_BUTTON    0
#define SONOFF_LED      13
#define EEPROM_SALT     6543
#define SONOFF_RELAY    12

// Re-transmit the state every minute
#define READING_EVERY   1000 * 60

// TODO: implement Willmsg.
static SimpleMQTT mqtt(SONOFF_LED, EEPROM_SALT);
static Button button(SONOFF_BUTTON, false, true, 20);
static Timer t;


void publishReading();
void republish(char * payload, unsigned int length);
void setRelay(char * payload, unsigned int length);
void toggle();
int getState();
void setState(int state);
void onConnect();

void setup() {
  Serial.begin(115200);
  pinMode(SONOFF_RELAY, OUTPUT);
  setState(HIGH);


  mqtt.subscribeTo("republish", republish);
  mqtt.subscribeTo("relay/set", setRelay);
  mqtt.onConnect(onConnect);

  mqtt.beginConfig();

  setState(HIGH);

  t.every(READING_EVERY, publishReading);
}

void loop() {
  mqtt.tick();
  t.update();

  button.read();
  if (button.pressedFor(10000)) {
    Serial.println("Reset Settings");
    mqtt.reset();
  } else if (button.wasReleased()) {
    toggle();
  }
}

void publishReading() {
  int state = getState();
  if (state == HIGH) {
      mqtt.publish("relay", "1");
  } else {
      mqtt.publish("relay", "0");
  }
}

void republish(char * payload, unsigned int length) {
  publishReading();
}

void setRelay(char * payload, unsigned int length) {
  if (payload[0] == '1') {
    setState(HIGH);
  } else {
    setState(LOW);
  }
  publishReading();
}

void toggle() {
  setState(!getState());
  publishReading();
}

int getState() {
  return digitalRead(SONOFF_RELAY);
}

void setState(int state) {
  Serial.printf("Setting state to %d\n", state);
  digitalWrite(SONOFF_RELAY, state);
  digitalWrite(SONOFF_LED, !state); // led is active low
}

void onConnect() {
  publishReading();
}
