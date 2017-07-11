#include <Arduino.h>
#include <Button.h>

Button button = Button(7, true, true, 25);


const char * preamble = "1010101010101010101010101010101010101010101010101010101010101010101010101010101010101010";
const char * header = "110011001100110011001100110000000011111111";
const char * code = "0011001100110011110000111100001111000011110000110011110011000011001111001100001100111100001111001100001100111100110011001100110011000011001111001100001111000011110011";

#define BIT_US 330
#define TX_PIN 9

void onPress(Button& b);
void do_transmit();

void setup() {
    Serial.begin(115200);
    pinMode(TX_PIN, OUTPUT);
    // pinMode(7, INPUT_PULLUP);
}


void do_transmit() {
  Serial.println("Transmitting");

  size_t preamble_size = strlen(preamble);
  size_t header_size = strlen(header);
  size_t code_size = strlen(code);

  for (size_t i = 0; i < 20; i++) {
    for (size_t i = 0; i < preamble_size; i++) {
      // Serial.print(preamble[i]);
      digitalWrite(TX_PIN, preamble[i] == '0' ? HIGH : LOW);
      delayMicroseconds(BIT_US);
    }
  }

  for (size_t i = 0; i < 1; i++) {
    for (size_t i = 0; i < header_size; i++) {
      // Serial.print(header[i]);
      digitalWrite(TX_PIN, header[i] == '0' ? HIGH : LOW);
      delayMicroseconds(BIT_US);
    }

    for (size_t i = 0; i < code_size; i++) {
      digitalWrite(TX_PIN, code[i] == '0' ? HIGH : LOW);
      delayMicroseconds(BIT_US);
    }
  }
  Serial.println("Done");
}

void loop() {
  button.read();
  if (button.isPressed()) {
    do_transmit();
  } else {
    // Serial.println("Not pressed");
  }
}
