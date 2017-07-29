#include <Arduino.h>
#include <Scheduler.h>
#include <Transmitter.h>
#include <RAEXRemote.h>

#define TX_PIN 9
#define BUTTON_PIN 7
#define LED_PIN 13

static Transmitter transmitter = Transmitter(TX_PIN);
static Scheduler scheduler = Scheduler();
static RAEXRemote raexRemote = RAEXRemote(&scheduler, &transmitter);

static bool hasChannel = false;
static bool hasRemote = false;
static bool isWaiting = false;
static uint8_t channel;
static uint16_t remote;
static uint8_t lastAction = 127;

void setup() {
  Serial.begin(115200);
  pinMode(TX_PIN, OUTPUT);
}

static char buffer[1024];
static int pos = 0;
char * readSerialLine() {
  if (Serial.available() > 0){
    int byte = Serial.read();
    if (byte >= 0) {
      if (byte == '\n') {
        buffer[pos] = 0;
        pos = 0;
        return buffer;
      } else if (byte == '\r') {
        // Ignore CR.
      } else {
        buffer[pos++] = (char)byte;
        buffer[pos] = 0;
      }
    }
  }
  return NULL;
}

void loop() {
  if (!hasChannel) {
    if (!isWaiting) Serial.print("Channel: ");
    isWaiting = true;
    char * line = readSerialLine();
    if (line != NULL) {
      isWaiting = false;
      hasChannel = true;
      channel = atoi(line);
    }
  } else if (!hasRemote) {
    if (!isWaiting) Serial.print("Remote: ");
    isWaiting = true;
    char * line = readSerialLine();
    if (line != NULL) {
      isWaiting = false;
      hasRemote = true;
      remote = atoi(line);
    }
  } else {
    if (!isWaiting) {
      Serial.print("Channel: ");
      Serial.print(channel);
      Serial.print(", Remote: ");
      Serial.println(remote);
      Serial.print("Action: ");
    }
    isWaiting = true;

    char * line = readSerialLine();
    if (line != NULL) {
      isWaiting = false;
      if (strcmp(line, "reset") == 0) {
        hasChannel = false;
        hasRemote = false;
      } else {
        uint8_t action;
        switch (line[0]) {
          case 'P':
          case 'p':
          action = 127;
          break;
          case 'U':
          case 'u':
          action = 254;
          break;
          case 'D':
          case 'd':
          action = 252;
          break;
          case 'S':
          case 's':
          action = 253;
          break;
          default:
          action = lastAction;
          break;
        }

        lastAction = action;
        RAEXRemoteCode raexRemoteCode = RAEXRemoteCode(channel, remote, action);
        raexRemote.transmitCode(&raexRemoteCode);
      }
    }
  }
}
