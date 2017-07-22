#ifndef TRANSMITTER_H_
#define TRANSMITTER_H_

#include <Arduino.h>
#include <Scheduler.h>

class Transmitter {
private:
  int txPin;
public:
  Transmitter(int txPin);
  void writeHigh();
  void writeLow();
};

class ManchesterTransmitter {
private:
  unsigned int clockWidth;
  Transmitter* transmitter;
  Scheduler* scheduler;
public:
  ManchesterTransmitter(Transmitter *transmitter, Scheduler* scheduler, unsigned int clockWidth);
  void writeBit(bool bit);
  void writeByteBigEndian(uint8_t byte);
};

class AGCTransmitter {
private:
  unsigned int clockWidth;
  Transmitter* transmitter;
  Scheduler* scheduler;
public:
  AGCTransmitter(Transmitter *transmitter, Scheduler* scheduler, unsigned int clockWidth);
  void writeAGC(unsigned int numRounds);
};

#endif
