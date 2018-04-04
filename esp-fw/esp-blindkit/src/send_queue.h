#ifndef SEND_QUEUE_H_
#define SEND_QUEUE_H_

#include <Arduino.h>
#include "transmit.h"

void send_pending_raex_tx(int tx_pin, int now);
void request_raex_tx(uint16_t remote, uint8_t channel, raex_action_t action);
void print_pending_raex_messages();

#endif /* SEND_QUEUE_H_ */
