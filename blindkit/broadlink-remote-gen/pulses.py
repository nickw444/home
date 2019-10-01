import argparse
import base64
import struct
from enum import Enum
from typing import NamedTuple, List


class BlindAction(Enum):
    UP = 254
    DOWN = 252
    STOP = 253
    PAIR = 127


def main():
    parser = argparse.ArgumentParser(description='Generate Broadlink data for RAEX blinds')
    parser.add_argument('--remote', required=True, type=int)
    parser.add_argument('--channel', required=True, type=int)
    parser.add_argument('--ncodes', type=int, default=1)
    parser.add_argument('--broadlink-repeats', type=int, default=7, help='Number of time for the broadlink to repeat the transmission')
    parser.add_argument('--payload-repeats', type=int, default=4, help='Number of repetitions of the remote payload within the broadlink payload')
    args = parser.parse_args()

    encoder = BroadlinkEncoder(repeats=args.broadlink_repeats)

    for x in range(args.ncodes):
        channel = args.channel + x
        print("+++ Remote {}, Channel {} +++".format(args.remote, channel))
        for action_name, action in [
            ('UP', BlindAction.UP),
            ('DOWN', BlindAction.DOWN),
            ('STOP', BlindAction.STOP),
            ('PAIR', BlindAction.PAIR)
        ]:
            remote_code = RemoteCode(
                channel=channel,
                remote=args.remote,
                action=action
            )

            data = []
            data += build_preamble()
            for _ in range(args.payload_repeats):
                data += remote_code.get_phase_durations()
                # Drop a little bit of padding between payloads within a transmission
                data += [PhaseDuration(not data[-1].phase, 5000)]

            payload = encoder.encode(data)
            # print("".join("{:02x}".format(c) for c in payload))
            print('{}: {}'.format(action_name, base64.b64encode(payload)))
        print("")


def build_preamble():
    data = []
    for x in range(200):
        data.append(PhaseDuration(1, 330))
        data.append(PhaseDuration(0, 330))
    return data


class PhaseDuration(NamedTuple):
    phase: int  # either 1 or 0
    duration: int  # Duration in us


class RemoteCode(object):
    CLOCK_WIDTH = 660

    def __init__(self, channel: int, remote: int, action: BlindAction):
        self.channel = channel
        self.remote = remote
        self.action = action

    def get_phase_durations(self):
        return self._build_header() + self._build_payload()

    def _build_header(self):
        data = []
        for x in range(20):
            data.append(PhaseDuration(0, RemoteCode.CLOCK_WIDTH))
            data.append(PhaseDuration(1, RemoteCode.CLOCK_WIDTH))

        # Last transmission to LOW
        data.append(PhaseDuration(0, RemoteCode.CLOCK_WIDTH))

        # Transmit long part
        data.append(PhaseDuration(1, RemoteCode.CLOCK_WIDTH * 4))
        data.append(PhaseDuration(0, RemoteCode.CLOCK_WIDTH * 4))
        return data

    def _build_payload(self):
        data = self._encode()
        return self._get_manchester_phases(data)

    def _encode(self):
        data = b''
        data += struct.pack('B', self.channel)
        data += struct.pack('B', self.remote & 0xFF)
        data += struct.pack('B', self.remote >> 8)
        data += struct.pack('B', self.action.value)
        data += struct.pack('B', self._calculate_checksum())
        return data

    def _get_manchester_phases(self, bytes):
        """
        Takes a bytes-like object and writes the bit phases in order
        Each byte has its least significant bit written first.
        """
        HIGH = [
            PhaseDuration(0, RemoteCode.CLOCK_WIDTH),
            PhaseDuration(1, RemoteCode.CLOCK_WIDTH)
        ]
        LOW = [
            PhaseDuration(1, RemoteCode.CLOCK_WIDTH),
            PhaseDuration(0, RemoteCode.CLOCK_WIDTH)
        ]

        # Obtain the phase positions for each clock cycle
        phases = []

        # Write fixed first bit
        phases += LOW

        for byte in bytes:
            bits = bin(byte)[2:].rjust(8, '0')[::-1]
            for bit in bits:
                phases += HIGH if bit == '1' else LOW

        return phases

    def _calculate_checksum(self):
        return (self.channel + (self.remote & 0xFF) + \
                (self.remote >> 8) + (self.action.value & 0xFF) + 3) & 0xFF


class BroadlinkEncoder(object):
    def __init__(self, repeats: int):
        self._repeats = repeats

    def encode(self, phases: List[PhaseDuration]):
        data = self.encode_phases(phases)

        payload = b''
        # It seems there's no need to send the header when dealing with
        # home-assistant
        # payload = '\x02\x00\x00\x00' #
        payload += b'\xb2'  # RF
        payload += struct.pack('B', self._repeats)  # Repeat Count
        payload += struct.pack('<H', len(data))  # Data length
        payload += data

        return payload

    def encode_phases(self, phases: List[PhaseDuration]) -> bytes:
        phases = self.optimize_phases(phases)
        data = b''
        if phases[0].phase == 0:
            # Insert a dummy to invert the subsequent sequence
            data += b'\x00'

        data += b''.join([
            self.pulse_length(phase.duration) for phase in phases
        ])

        return data

    def optimize_phases(self, phases: List[PhaseDuration]) -> List[PhaseDuration]:
        """Optimizes a collection of phase duration to join sequenced durations"""
        optimized = []
        for phase in phases:
            prev = optimized[-1] if len(optimized) else None
            if prev is not None and prev[0] == phase.phase:
                prev[1] += phase.duration
            else:
                optimized.append([phase.phase, phase.duration])

        return [
            PhaseDuration(phase, duration) for (phase, duration) in optimized
        ]

    def pulse_length(self, duration: int):
        """Convert a duration in us to a pulse byte"""
        return struct.pack('B', int(round(duration * 269 / 8192)))


if __name__ == '__main__':
    main()
