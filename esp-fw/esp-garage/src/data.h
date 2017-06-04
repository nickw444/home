
enum DOOR_STATE {
  DOOR_OPEN,
  DOOR_CLOSED,
  DOOR_UNKNOWN,
};

enum RELAY_STATE {
  RELAY_OPEN,
  RELAY_CLOSED,
  RELAY_UNKNOWN,
};

const char * serializeDoorState(DOOR_STATE state);
const char * serializeRelayState(RELAY_STATE state);
