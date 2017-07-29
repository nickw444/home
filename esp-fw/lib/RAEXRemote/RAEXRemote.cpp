#include "RAEXRemote.h"

RAEXRemoteCode::RAEXRemoteCode(uint8_t channel, uint16_t remote, uint8_t action) {
  this->channel = channel;
  this->remote = remote;
  this->action = action;
}

RAEXRemoteCode RAEXRemoteCode::deserialise(String serialised) {
  unsigned int channel;
  unsigned int remote;
  unsigned int action;
  sscanf(serialised.c_str(), "%u,%u,%u", &channel, &remote, &action);
  return RAEXRemoteCode(channel, remote, action);
}

uint8_t RAEXRemoteCode::calculateChecksum() {
  return channel + (remote & 0xFF) + (remote >> 8) + action + 3;
}


RAEXRemote::RAEXRemote(Scheduler* scheduler, Transmitter* transmitter) {
  this->scheduler = scheduler;
  this->transmitter = transmitter;
  this->manchester = new ManchesterTransmitter(transmitter, scheduler, 330 * 2);
  this->agc = new AGCTransmitter(transmitter, scheduler, 330);
}

void RAEXRemote::transmitCode(RAEXRemoteCode* code) {
  agc->writeAGC(300);

  for (size_t i = 0; i < 4; i++) {
    writeHeader();
    writeData(code);
  }
}

void RAEXRemote::writeHeader() {
  for (size_t i = 0; i < 20; i++) {
    transmitter->writeLow();
    scheduler->delayUs(330 * 2);
    transmitter->writeHigh();
    scheduler->delayUs(330 * 2);
  }

  // Finish off pre-header
  transmitter->writeLow();
  scheduler->delayUs(330 * 2);

  // Transmit long part
  transmitter->writeHigh();
  scheduler->delayUs(330 * 2 * 4);
  transmitter->writeLow();
  scheduler->delayUs(330 * 2 * 4);
}

void RAEXRemote::writeData(RAEXRemoteCode *code) {
  uint8_t checksum = code->calculateChecksum();

  // Write fixed first bit.
  manchester->writeBit(0);
  // Write code data.
  manchester->writeByteBigEndian(code->channel);
  manchester->writeByteBigEndian(code->remote & 0xFF);
  manchester->writeByteBigEndian(code->remote >> 8);
  manchester->writeByteBigEndian(code->action);
  manchester->writeByteBigEndian(checksum);

  // Write fixed last bits.
  manchester->writeBit(0);
  manchester->writeBit(0);
}
