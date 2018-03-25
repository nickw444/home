#include "transmit.h"

#define CLOCK_WIDTH 330

static void manchesterWriteBit(int txPin, int clockWidth, bool bit) {
  digitalWrite(txPin, bit ? LOW : HIGH);
  delayMicroseconds(clockWidth);
  digitalWrite(txPin, bit ? HIGH : LOW);
  delayMicroseconds(clockWidth);
}

static void manchesterWriteByteBigEndian(int txPin, int clockWidth, uint8_t byte) {
  for (size_t i = 0; i < 8; i++) {
    bool bit = (bool)(byte & (1 << i));
    manchesterWriteBit(txPin, clockWidth, bit);
  }
}

void txPrepare(int txPin, int numRounds) {
  for (int i = 0; i < numRounds; i++) {
    digitalWrite(txPin, HIGH);
    delayMicroseconds(CLOCK_WIDTH);
    digitalWrite(txPin, LOW);
    delayMicroseconds(CLOCK_WIDTH);
  }
  digitalWrite(txPin, HIGH);
  delayMicroseconds(CLOCK_WIDTH);
}

static void txRaexWriteHeader(int txPin) {
  for (size_t i = 0; i < 20; i++) {
    digitalWrite(txPin, LOW);
    delayMicroseconds(CLOCK_WIDTH * 2);
    digitalWrite(txPin, HIGH);
    delayMicroseconds(CLOCK_WIDTH * 2);
  }

  // Finish off pre-header
  digitalWrite(txPin, LOW);
  delayMicroseconds(CLOCK_WIDTH * 2);

  // Transmit long part
  digitalWrite(txPin, HIGH);
  delayMicroseconds(CLOCK_WIDTH * 2 * 4);
  digitalWrite(txPin, LOW);
  delayMicroseconds(CLOCK_WIDTH * 2 * 4);
}

static void rxRaexWriteData(int txPin, uint16_t remote, uint8_t channel,
  raex_action_t action, int checksum) {
    // Write fixed first bit.
    manchesterWriteBit(txPin, CLOCK_WIDTH, 0);
    // Write code data.
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH, channel);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH, remote & 0xFF);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH, remote >> 8);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH, action);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH, checksum);

    // Write fixed last bits.
    manchesterWriteBit(txPin, CLOCK_WIDTH, 0);
    manchesterWriteBit(txPin, CLOCK_WIDTH, 0);
}

static uint8_t txRaexCalculateChecksum(uint16_t remote, uint8_t channel,
  uint8_t action) {

  return channel + (remote & 0xFF) + (remote >> 8) + (action & 0xFF) + 3;
}

void txRaexSend(int txPin, uint16_t remote, uint8_t channel, raex_action_t action) {
  uint8_t checksum = txRaexCalculateChecksum(remote, channel, action);

  txRaexWriteHeader(txPin);
  rxRaexWriteData(txPin, remote, channel, action, checksum);
}
