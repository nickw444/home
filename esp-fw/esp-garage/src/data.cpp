#include "data.h"

const char * serializeDoorState(DOOR_STATE state) {
  switch (state) {
    case DOOR_OPEN:
      return "OPEN";
    case DOOR_CLOSED:
      return "CLOSED";
    case DOOR_UNKNOWN:
      return "UNKNOWN";
  }
}

const char * serializeRelayState(RELAY_STATE state) {
  switch (state) {
    case RELAY_OPEN:
      return "OPEN";
    case RELAY_CLOSED:
      return "CLOSED";
    case RELAY_UNKNOWN:
      return "UNKNOWN";
  }
}
