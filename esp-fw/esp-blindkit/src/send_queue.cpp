#include "send_queue.h"

#define TRANSMIT_RETRIES 4
#define TRANSMIT_RETRY_DELAY_MS 1500

struct raex_message {
  int execute_time;
  uint16_t remote;
  uint8_t channel;
  raex_action_t action;

  struct raex_message *next;
};

struct raex_message *queue_head = NULL;


void send_pending_raex_tx(int tx_pin, int now) {
  if (queue_head != NULL && queue_head->execute_time <= now) {
    struct raex_message *curr = queue_head;
    Serial.printf("Executing send for [%d,%d,%d]\n", curr->remote, curr->channel, curr->action);
    txPrepare(tx_pin, 200);
    txRaexSend(tx_pin, curr->remote, curr->channel, curr->action);
    queue_head = curr->next;
    free(curr);
  }
}

void request_raex_tx(uint16_t remote, uint8_t channel, raex_action_t action) {
  int now = millis();

  // Purge any existing messages for this remote/channel.
  Serial.printf("Purging from queue\n");
  struct raex_message *curr = queue_head;
  struct raex_message *prev = NULL;
  while (curr != NULL) {
    if (curr->remote == remote && curr->channel == channel) {
      Serial.printf("Found message to purge: %p\n", curr);
      struct raex_message *next = curr->next;
      if (prev == NULL) {
        queue_head = next;
      } else {
        prev->next = next;
      }
      free(curr);
      curr = next;
    } else {
      prev = curr;
      curr = curr->next;
    }
  }

  // Insert new messages for this remote/channel.
  Serial.printf("Inserting new message\n");
  for (int i = 0; i < TRANSMIT_RETRIES; i++) {
    struct raex_message *msg = (struct raex_message *)malloc(sizeof(struct raex_message));
    msg->remote = remote;
    msg->channel = channel;
    msg->action = action;
    msg->execute_time = now + (i * TRANSMIT_RETRY_DELAY_MS);
    msg->next = NULL;

    Serial.printf("Created node %p\n", msg);

    if (queue_head == NULL) {
      Serial.printf("Inserting at queue head\n");
      queue_head = msg;
    } else {
      struct raex_message *curr = queue_head;
      struct raex_message *prev = NULL;
      bool inserted = false;

      while (curr != NULL) {
        if (msg->execute_time < curr->execute_time) {
          if (prev == NULL) {
            queue_head = msg;
          } else {
            prev->next = msg;
          }
          msg->next = curr;
          inserted = true;
          Serial.printf("Inserting in middle somewhere\n");
          break;
        }
        prev = curr;
        curr = curr->next;
      }
      // Must be the tail!
      if (!inserted) {
        Serial.printf("Inserting at tail\n");
        prev->next = msg;
      }
    }
  }
  Serial.printf("Done\n");
}

void print_pending_raex_messages() {
  struct raex_message *curr = queue_head;
  while (curr != NULL) {
    Serial.printf("[%d,%d,%d,%d] -> ", curr->remote, curr->channel, curr->action, curr->execute_time);
    curr = curr->next;
  }
  Serial.printf("X\n");
}
