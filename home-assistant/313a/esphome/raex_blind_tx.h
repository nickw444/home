#include "esphome.h"

#define TX_PIN 3
#define TRANSMIT_RETRIES 7
#define TRANSMIT_RETRY_DELAY_MS 1000
#define CLOCK_WIDTH 330

static const char *TAG = "raex_blind_tx";

enum _raex_action {
  TX_RAEX_ACTION_UP = 254,
  TX_RAEX_ACTION_DOWN = 252,
  TX_RAEX_ACTION_STOP = 253,
  TX_RAEX_ACTION_PAIR = 127,
};
typedef enum _raex_action raex_action_t;

class RaexMessage {
 public:
  RaexMessage(int execute_time, uint16_t remote_id, uint8_t channel_id, raex_action_t action_id);

  int execute_time;
  uint16_t remote_id;
  uint8_t channel_id;
  raex_action_t action_id;
};

RaexMessage::RaexMessage(int execute_time, uint16_t remote_id, uint8_t channel_id, raex_action_t action_id): execute_time(execute_time), remote_id(remote_id), channel_id(channel_id), action_id(action_id) {};


class RaexBlindTransmitComponent : public Component, public CustomAPIDevice {
 private:
  std::vector<RaexMessage> pending_messages;

 public:
  void setup() override {
    // This will be called once to set up the component
    // think of it as the setup() call in Arduino
    pinMode(3, OUTPUT);

    register_service(&RaexBlindTransmitComponent::transmit, "transmit",
                     {"remote_id", "channel_id", "action"});
  }

  void loop() override {
    if (!pending_messages.empty()) {
      int now = millis();
      auto msg = pending_messages.begin();
      if (msg->execute_time <= now) {
        ESP_LOGD(TAG, "Executing send for [%d,%d,%d]", msg->remote_id, msg->channel_id, msg->action_id);

        txPrepare(TX_PIN, 200);
        txRaexSend(TX_PIN, msg->remote_id, msg->channel_id, msg->action_id);

        pending_messages.erase(msg);
      }
    }
  }

  void transmit(int remote_id, int channel_id, std::string action) {
    ESP_LOGD(TAG, "Enqueing: %d, %d, %s", remote_id, channel_id, action.c_str());
    int now = millis();

    raex_action_t action_id;
    if (action.compare("OPEN") == 0) {
      action_id = TX_RAEX_ACTION_UP;
    } else if (action.compare("CLOSE") == 0) {
      action_id = TX_RAEX_ACTION_DOWN;
    } else if (action.compare("STOP") == 0) {
      action_id = TX_RAEX_ACTION_STOP;
    } else if (action.compare("PAIR") == 0) {
      action_id = TX_RAEX_ACTION_PAIR;
    } else {
      ESP_LOGE(TAG, "Malformed payload received. Unknown action [%s]", action.c_str());
      return;
    }


    // Purge any existing messages for this remote/channel.
    for (auto it = pending_messages.begin(); it != pending_messages.end(); ) {
        if (it->remote_id == remote_id && it->channel_id == channel_id) {
          ESP_LOGD(TAG, "Found message to purge at address: %d", it - pending_messages.begin());
          it = pending_messages.erase(it);
        } else {
          it++;
        }
    }

    // Insert new messages (in order) for this remote/channel.
    ESP_LOGD(TAG, "Inserting new messages");
    for (int i = 0; i < TRANSMIT_RETRIES; i++) {
      auto msg =
          RaexMessage(now + (i * TRANSMIT_RETRY_DELAY_MS), remote_id, channel_id, action_id);

      // Insert in order
      bool inserted = false;
      for (auto it = pending_messages.begin(); it != pending_messages.end(); it++) {
        if (msg.execute_time < it->execute_time) {
          ESP_LOGD(TAG, "Inserting before %d", it - pending_messages.begin());
          pending_messages.insert(it, msg);
          inserted = true;
          break;
        }
      }

      // If not inserted, it must be the new tail element
      if (!inserted) {
        ESP_LOGD(TAG, "Inserting at tail");
        pending_messages.push_back(msg);
      }
    }

    ESP_LOGD(TAG, "New queue:");
    for (auto it = pending_messages.begin(); it != pending_messages.end(); it++) {
      ESP_LOGD(TAG, " - [%d, %d, %d, %d]", it->remote_id, it->channel_id, it->action_id, it->execute_time);
    }
  }

  static void manchesterWriteBit(int txPin, int clockWidth, bool bit) {
    digitalWrite(txPin, bit ? LOW : HIGH);
    delayMicroseconds(clockWidth);
    digitalWrite(txPin, bit ? HIGH : LOW);
    delayMicroseconds(clockWidth);
  }

  static void manchesterWriteByteBigEndian(int txPin, int clockWidth, uint8_t byte) {
    for (size_t i = 0; i < 8; i++) {
      bool bit = (bool) (byte & (1 << i));
      manchesterWriteBit(txPin, clockWidth, bit);
    }
  }

  static void txPrepare(int txPin, int numRounds) {
    for (int i = 0; i < numRounds; i++) {
      digitalWrite(txPin, HIGH);
      delayMicroseconds(CLOCK_WIDTH);
      digitalWrite(txPin, LOW);
      delayMicroseconds(CLOCK_WIDTH);
    }
    digitalWrite(txPin, HIGH);
    delayMicroseconds(CLOCK_WIDTH);
  }

  static void txRaexWriteHeader(int txPin) {
    for (size_t i = 0; i < 20; i++) {
      digitalWrite(txPin, LOW);
      delayMicroseconds(CLOCK_WIDTH * 2);
      digitalWrite(txPin, HIGH);
      delayMicroseconds(CLOCK_WIDTH * 2);
    }

    // Finish off pre-header
    digitalWrite(txPin, LOW);
    delayMicroseconds(CLOCK_WIDTH * 2);

    // Transmit long part
    digitalWrite(txPin, HIGH);
    delayMicroseconds(CLOCK_WIDTH * 2 * 4);
    digitalWrite(txPin, LOW);
    delayMicroseconds(CLOCK_WIDTH * 2 * 4);
  }

  static void rxRaexWriteData(int txPin, uint16_t remote, uint8_t channel, raex_action_t action, int checksum) {
    // Write fixed first bit.
    manchesterWriteBit(txPin, CLOCK_WIDTH * 2, 0);
    // Write code data.
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH * 2, channel);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH * 2, remote & 0xFF);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH * 2, remote >> 8);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH * 2, action);
    manchesterWriteByteBigEndian(txPin, CLOCK_WIDTH * 2, checksum);

    // Write fixed last bits.
    manchesterWriteBit(txPin, CLOCK_WIDTH * 2, 0);
    manchesterWriteBit(txPin, CLOCK_WIDTH * 2, 0);
  }

  static uint8_t txRaexCalculateChecksum(uint16_t remote, uint8_t channel, uint8_t action) {
    return channel + (remote & 0xFF) + (remote >> 8) + (action & 0xFF) + 3;
  }

  static void txRaexSend(int txPin, uint16_t remote, uint8_t channel, raex_action_t action) {
    uint8_t checksum = txRaexCalculateChecksum(remote, channel, action);

    for (int i = 0; i < 4; i++) {
      txRaexWriteHeader(txPin);
      rxRaexWriteData(txPin, remote, channel, action, checksum);
    }
  }
};

