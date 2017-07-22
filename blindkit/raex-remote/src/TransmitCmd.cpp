#include "TransmitCmd.h"

TransmitCmd::TransmitCmd(RemoteType remoteType, String payload) {
  this->payload = payload;
  this->remoteType = remoteType;
}

TransmitCmd TransmitCmd::deserialise(String payload) {

  return TransmitCmd(RemoteType::RAEX, "");
}

RemoteType TransmitCmd::getRemoteType() {
  return this->remoteType;
}

String TransmitCmd::getPayload() {
  return this->payload;
}
