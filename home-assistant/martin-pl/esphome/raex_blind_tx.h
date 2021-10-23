#include "esphome.h"

// Goals
//  - Eventual consistency; retry over a longer period of time.
//  - Short-term retries for instant control where needed.

#define TX_PIN 3

#define TRANSMIT_BLOCKS 6
#define TRANSMIT_BLOCKS_DELAY_MS (10 * 60 * 1000) // 6 attempts, one every 10 mins, in order to reach eventual consistency

#define TRANSMIT_RETRIES 10
#define TRANSMIT_RETRY_DELAY_MS 2000

#define CLOCK_WIDTH 360
#define LOCKOUT_DELAY_MS 200

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
    RaexMessage(
      int execute_time,
      uint8_t blocks_remain,
      uint8_t retries_remain,
      uint16_t remote_id,
      uint8_t channel_id,
      raex_action_t action_id
    );

    int execute_time;
    uint8_t blocks_remain;
    uint8_t retries_remain;
    uint16_t remote_id;
    uint8_t channel_id;
    raex_action_t action_id;
};

RaexMessage::RaexMessage(
  int execute_time,
  uint8_t blocks_remain,
  uint8_t retries_remain,
  uint16_t remote_id,
  uint8_t channel_id,
  raex_action_t action_id
):
  execute_time(execute_time),
  blocks_remain(blocks_remain),
  retries_remain(retries_remain),
  remote_id(remote_id),
  channel_id(channel_id),
  action_id(action_id) {};


class RaexBlindTransmitComponent : public Component, public CustomAPIDevice {
  private:
    std::map<uint32_t, RaexMessage*> pending_messages;
    int lockout_until = 0;

  public:

    void setup() override {
      // This will be called once to set up the component
      // think of it as the setup() call in Arduino
      pinMode(3, OUTPUT);

      register_service(&RaexBlindTransmitComponent::transmit, "transmit",
                      {"remote_id", "channel_id", "action"});
      register_service(&RaexBlindTransmitComponent::transmit_custom, "transmit_custom",
                      {"remote_id", "channel_id", "action", "blocks", "retries"});
    }

    void loop() override {
      if (!pending_messages.empty()) {
        int now = millis();
        if (lockout_until > now) {
          // To avoid trasmissions which are sent nearby to improve successful transmission.
          return;
        }

        // Check for pending messages
        for (auto it = pending_messages.begin(); it != pending_messages.end(); it++) {
          auto key = it->first;
          auto msg = it->second;

          if (msg->execute_time <= now) {
            ESP_LOGD(TAG, "Executing send for [%d,%d,%d] [retries: %d, blocks: %d]",
              msg->remote_id, msg->channel_id, msg->action_id, msg->retries_remain, msg->blocks_remain);

            txPrepare(TX_PIN, 200, CLOCK_WIDTH);
            txRaexSend(TX_PIN, msg->remote_id, msg->channel_id, msg->action_id, CLOCK_WIDTH);
            lockout_until = millis() + LOCKOUT_DELAY_MS;

            if (msg->retries_remain > 0) {
              // Schedule next retry
              msg->retries_remain--;
              msg->execute_time = now + TRANSMIT_RETRY_DELAY_MS;
              return;
            }

            if (msg->blocks_remain > 0) {
              // Schedule next block
              msg->blocks_remain--;
              msg->retries_remain = TRANSMIT_RETRIES - 1;
              msg->execute_time = now + TRANSMIT_BLOCKS_DELAY_MS;
              return;
            }


            // No retries or blocks remain, remove from pending messages
            ESP_LOGD(TAG, "No retries or blocks remain for [%d,%d,%d], removing",
              msg->remote_id, msg->channel_id, msg->action_id);
            pending_messages.erase(it);
            delete msg;
            return;
          }
        }
      }
    }

    void transmit(int remote_id, int channel_id, std::string action) {
      transmit_custom(remote_id, channel_id, action, TRANSMIT_BLOCKS, TRANSMIT_RETRIES, CLOCK_WIDTH);
    }

    void transmit_custom(int remote_id, int channel_id, std::string action, int blocks, int retries, int clock_width) {
      ESP_LOGD(TAG, "Enqueing: %d, %d, %s, ", remote_id, channel_id, action.c_str());
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

      uint32_t key = remote_id << 8 | channel_id;
      auto search = pending_messages.find(key);
      if (search != pending_messages.end()) {
        ESP_LOGD(TAG, "Reusing existing slot with key: %d", key);
        auto msg = search->second;
        msg->execute_time = now;
        msg->blocks_remain = blocks - 1;
        msg->retries_remain = retries - 1;
        msg->action_id = action_id;
      } else {
        ESP_LOGD(TAG, "Inserting new message with key: %d", key);
        auto msg = new RaexMessage(now, blocks - 1, retries - 1, remote_id, channel_id, action_id);
        pending_messages.insert({key, msg});
      }

      ESP_LOGD(TAG, "New queue:");
      for (auto it = pending_messages.begin(); it != pending_messages.end(); it++) {
        ESP_LOGD(TAG, " - %d: [%d, %d, %d, %d]", it->first, it->second->remote_id, it->second->channel_id, it->second->action_id, it->second->execute_time);
      }
    }

    static void manchesterWriteBit(int txPin, uint16_t clockWidth, bool bit) {
      digitalWrite(txPin, bit ? LOW : HIGH);
      delayMicroseconds(clockWidth);
      digitalWrite(txPin, bit ? HIGH : LOW);
      delayMicroseconds(clockWidth);
    }

    static void manchesterWriteByteBigEndian(int txPin, uint16_t clockWidth, uint8_t byte) {
      for (size_t i = 0; i < 8; i++) {
        bool bit = (bool) (byte & (1 << i));
        manchesterWriteBit(txPin, clockWidth, bit);
      }
    }

    static void txPrepare(int txPin, int numRounds, uint16_t clockWidth) {
      for (int i = 0; i < numRounds; i++) {
        digitalWrite(txPin, HIGH);
        delayMicroseconds(clockWidth);
        digitalWrite(txPin, LOW);
        delayMicroseconds(clockWidth);
      }
      digitalWrite(txPin, HIGH);
      delayMicroseconds(clockWidth);
    }

    static void txRaexWriteHeader(int txPin, uint16_t clockWidth) {
      for (size_t i = 0; i < 20; i++) {
        digitalWrite(txPin, LOW);
        delayMicroseconds(clockWidth * 2);
        digitalWrite(txPin, HIGH);
        delayMicroseconds(clockWidth * 2);
      }

      // Finish off pre-header
      digitalWrite(txPin, LOW);
      delayMicroseconds(clockWidth * 2);

      // Transmit long part
      digitalWrite(txPin, HIGH);
      delayMicroseconds(clockWidth * 2 * 4);
      digitalWrite(txPin, LOW);
      delayMicroseconds(clockWidth * 2 * 4);
    }

    static void rxRaexWriteData(int txPin, uint16_t remote, uint8_t channel, raex_action_t action, int checksum, uint16_t clockWidth) {
      // Write fixed first bit.
      manchesterWriteBit(txPin, clockWidth * 2, 0);
      // Write code data.
      manchesterWriteByteBigEndian(txPin, clockWidth * 2, channel);
      manchesterWriteByteBigEndian(txPin, clockWidth * 2, remote & 0xFF);
      manchesterWriteByteBigEndian(txPin, clockWidth * 2, remote >> 8);
      manchesterWriteByteBigEndian(txPin, clockWidth * 2, action);
      manchesterWriteByteBigEndian(txPin, clockWidth * 2, checksum);

      // Write fixed last bits.
      manchesterWriteBit(txPin, clockWidth * 2, 0);
      manchesterWriteBit(txPin, clockWidth * 2, 0);
    }

    static uint8_t txRaexCalculateChecksum(uint16_t remote, uint8_t channel, uint8_t action) {
      return channel + (remote & 0xFF) + (remote >> 8) + (action & 0xFF) + 3;
    }

    static void txRaexSend(int txPin, uint16_t remote, uint8_t channel, raex_action_t action, uint16_t clockWidth) {
      uint8_t checksum = txRaexCalculateChecksum(remote, channel, action);

      for (int i = 0; i < 4; i++) {
        txRaexWriteHeader(txPin, clockWidth);
        rxRaexWriteData(txPin, remote, channel, action, checksum, clockWidth);
      }
    }
};

