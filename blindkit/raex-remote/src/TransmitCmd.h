#include <Arduino.h>

enum class RemoteType {
  RAEX
};

class TransmitCmd {
private:
  RemoteType remoteType;
  String payload;
public:
  TransmitCmd(RemoteType remoteType, String payload);
  static TransmitCmd deserialise(String payload);
  RemoteType getRemoteType();
  String getPayload();
};
