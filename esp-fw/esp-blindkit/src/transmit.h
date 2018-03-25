#ifndef TRANSMIT_H_
#define TRANSMIT_H_

#include <Arduino.h>

enum _raex_action {
  TX_RAEX_ACTION_UP = 254,
  TX_RAEX_ACTION_DOWN = 252,
  TX_RAEX_ACTION_STOP = 253,
  TX_RAEX_ACTION_PAIR = 127,
};

typedef enum _raex_action raex_action_t;

// Write AGC Header. Primes the transmitter for sending.
void txPrepare(int txPin, int numRounds);

// Write data for a RAEX blind
void txRaexSend(int txPin, uint16_t remote, uint8_t channel, raex_action_t action);

#endif /* TRANSMIT_H_ */
