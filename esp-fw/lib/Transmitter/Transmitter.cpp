#include "Transmitter.h"

Transmitter::Transmitter(int txPin) {
  this->txPin = txPin;
}

void Transmitter::writeHigh() {
  digitalWrite(this->txPin, HIGH);
}

void Transmitter::writeLow() {
  digitalWrite(this->txPin, LOW);
}

ManchesterTransmitter::ManchesterTransmitter(Transmitter *transmitter, Scheduler *scheduler, unsigned int clockWidth) {
  this->transmitter = transmitter;
  this->scheduler = scheduler;
  this->clockWidth = clockWidth;
}

void ManchesterTransmitter::writeBit(bool bit) {
  bit ? transmitter->writeLow() : transmitter->writeHigh();
  scheduler->delayUs(this->clockWidth);
  bit ? transmitter->writeHigh() : transmitter->writeLow();
  scheduler->delayUs(this->clockWidth);
}

void ManchesterTransmitter::writeByteBigEndian(uint8_t byte) {
  // Serial.print("Byte: ");
  // Serial.println(byte);
  for (size_t i = 0; i < 8; i++) {
    bool bit = (bool)(byte & (1 << i));
    writeBit(bit);
    // Serial.print(bit);
  }
  // Serial.println();
}

AGCTransmitter::AGCTransmitter(Transmitter *transmitter, Scheduler *scheduler, unsigned int clockWidth) {
  this->transmitter = transmitter;
  this->scheduler = scheduler;
  this->clockWidth = clockWidth;
}

void AGCTransmitter::writeAGC(unsigned int numRounds) {
  for (size_t i = 0; i < numRounds; i++) {
    transmitter->writeHigh();
    scheduler->delayUs(clockWidth);
    transmitter->writeLow();
    scheduler->delayUs(clockWidth);
  }

  transmitter->writeHigh();
  scheduler->delayUs(clockWidth);
}
