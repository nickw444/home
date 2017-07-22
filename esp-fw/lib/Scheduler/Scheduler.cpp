#include <Arduino.h>
#include "Scheduler.h"

Scheduler::Scheduler() {

}

void Scheduler::delayUs(unsigned int us) {
  delayMicroseconds(us);
}
