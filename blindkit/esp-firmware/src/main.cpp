#include <Arduino.h>

struct remote_code {
    uint8_t channel;
    uint16_t remote;
    uint8_t action;
}

struct remote_code deserialize_msg(char * s) {
    // Deserialise a remote code message.
    struct remote_code;
    remote_code.channel = 0;
    remote_code.remote = 0;
    remote_code.action = 0;
    return remote_code;
}

void do_remote_transmit(struct remote_code code) {
    uint8_t banks[5];
    banks[0] = code.channel;
    banks[1] = code.remote & 0xFF;
    banks[2] = code.remote & (0xFF << 7);
    banks[3] = code.action;
    banks[4] = calculate_checksum(code);

    // Transmit these banks
    transmit_manchester(banks, 5);
}

uint8_t calculate_checksum(struct remote_code code) {
    return code.channel + code.remote + code.action + 3;
}

void transmit_manchester(uint8_t banks[], int nbanks) {
    for (int i = 0; i < nbanks; ++i) {
        uint8_t bank = bank[i];
        for (int j = 0; j < 8; ++j) {
            bool bit = (bool)(bank & (1 << 7 - i))
            transmit_bit(val);
        }
    }
}

void write_agc() {

}
void write_preamble() {

}
void write_header() {
}

void transmit_bit(bool val) {
    // Low to high transition for 1, High to low for 0.
    digitalWrite(TX_PIN, val ? LOW : HIGH);
    delayMicroseconds(BIT_US);
    digitalWrite(TX_PIN, val ? HIGH : LOW);
    delayMicroseconds(BIT_US);
}
