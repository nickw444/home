#include <Transmitter.h>
#include <Scheduler.h>
#include <Arduino.h>

class RAEXRemoteCode {
public:
  uint8_t channel;
  uint16_t remote;
  uint8_t action;
  RAEXRemoteCode(uint8_t channel, uint16_t remote, uint8_t action);
  static RAEXRemoteCode deserialise(String serialised);
  uint8_t calculateChecksum();
};

class RAEXRemote {
private:
  Scheduler* scheduler;
  Transmitter* transmitter;
  ManchesterTransmitter* manchester;
  AGCTransmitter* agc;

  void writeHeader();
  void writeData(RAEXRemoteCode *code);
public:
  RAEXRemote(Scheduler* scheduler, Transmitter* transmitter);
  void transmitCode(RAEXRemoteCode* code);
};
