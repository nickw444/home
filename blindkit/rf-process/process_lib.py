"""
A collection of library utilities for working with and processing
captured WAV files.
"""

import struct
from collections import namedtuple, deque
from typing import Iterator, List, Tuple
from wave import Wave_read
from enum import Enum
import sys


Sample = namedtuple('Sample', ['val', 'pos'])

class FrameReader(object):
    def __init__(self, wav_file: Wave_read, no_channels: int,
                 channel: int) -> None:
        """
        FrameReader abstracts frame extraction from a WAV file. It allows
        channel agnostic data retreival from a Wave_read object.
        """
        self.wav_file = wav_file
        self.no_channels = no_channels
        self.channel = channel

    def get_samples(self) -> Iterator[Sample]:
        for i in range(self.wav_file.getnframes()):
            yield Sample(self._get_sample(self.wav_file.readframes(1)), i)

    def _get_sample(self, frame: bytes) -> int:
        data = struct.unpack("<{}h".format(self.no_channels), frame)
        return data[self.channel]


class Phase(Enum):
    """
    A classification of the phase of a waveform at a particular sample
    point.
    """
    POS = 'POS'
    NEG = 'NEG'
    ZERO = 'ZERO'


"""
SampleGroup is a collection of same-phase samples.
"""
SampleGroup = namedtuple('SampleGroup', ['phase', 'nsamples', 'start'])


class WaveformClassifier(object):
    def __init__(self, threshold: int) -> None:
        """
        WaveformClassifier abstracts the classification of a raw sample value
        to a Phase based on a particular threshold.
        """
        self.threshold = threshold

    def classify_sample(self, sample: int) -> Phase:
        if sample < -self.threshold:
            return Phase.NEG
        elif sample > self.threshold:
            return Phase.POS

        return Phase.ZERO


class PhaseGrouper(object):

    def __init__(self, frame_reader: FrameReader,
                 classifier: WaveformClassifier, sensitivity: int=3) -> None:
        """
        PhaseGrouper attempts to reconcile same-phased samples into
        discrete groups, identifying their length and their start
        positions.
        """
        self.frame_reader = frame_reader
        self.classifier = classifier
        self.sensitivity = sensitivity

    def get_grouped_samples(self) -> Iterator[SampleGroup]:
        """
        Groups phase changes. Yields phase groupings.
        Sensitivity controls how many samples of a different phase are required
        for a phase change.

        (Phase, nSamples)
        """
        curr_phase_start = 0
        curr_phase = None
        curr_phase_length = 0

        next_phase_start = 0
        next_phase_length = 0
        next_phase = None

        for (sample, idx) in self.frame_reader.get_samples():
            phase = self.classifier.classify_sample(sample)

            if phase != curr_phase:
                if next_phase != phase:
                    # Next phase is different to the existing one. Clear it.
                    next_phase_length = 0
                    next_phase_start = idx

                next_phase = phase
                next_phase_length += 1

            if next_phase_length > self.sensitivity:
                # Next phase has hit the sensitivty required. Phase change is
                # taking place. Yield the old phase.
                yield SampleGroup(curr_phase, curr_phase_length,
                                  curr_phase_start)

                curr_phase_length = next_phase_length
                curr_phase = next_phase
                curr_phase_start = next_phase_start

                next_phase = None
                next_phase_length = 0

            else:
                # No phase change hit yet. Increment current phase.
                curr_phase_length += 1

        yield SampleGroup(curr_phase, curr_phase_length, curr_phase_start)


class PreambleDetector(object):
    def __init__(self, width: int, sensitivity: int, nrepeats: int,
                 pattern: List[Phase]=[Phase.POS, Phase.NEG]) -> None:
        """
        Given a stream of SampleGroups, attempts to determine if a
        preamble has been identified.
        """
        self.width = width
        self.sensitivity = sensitivity
        self.nrepeats = nrepeats * len(pattern)
        self.orig_pattern = pattern
        self.pattern = deque(pattern)

        # Count up until preamble length
        self.preamble_start = 0
        self.curr_lock_no_repeats = 0

    def recv_group(self, group: SampleGroup) -> Tuple[bool, int]:
        if group.phase == self.pattern[0] and \
            group.nsamples > self.width - self.sensitivity and \
                group.nsamples < self.width + self.sensitivity:

            if self.preamble_start is None:
                self.preamble_start = group.start
            self.pattern.rotate(-1)
            self.curr_lock_no_repeats += 1
        else:
            self._reset()

        if self.curr_lock_no_repeats > self.nrepeats:
            return (True, self.preamble_start)
        else:
            return (False, None)

    def _reset(self):
        self.pattern = deque(self.orig_pattern)
        self.curr_lock_no_repeats = 0
        self.preamble_start = None

    def acknowledge(self):
        """
        Acknowledge that the preamble has been found. Reset preamble detector
        to look for more preamble.
        """
        self._reset()


class HeaderDetector(object):
    def __init__(self, width: int, sensitivity: int,
                 shape: List[Phase]=[Phase.NEG, Phase.POS]) -> None:

        """
        Given a stream of SampleGroups, attempts to determine if a
        header has been identified.
        """

        # TODO: Merge usages of this into PreambleDetector and rename to
        # something like PatternDetector.

        self.width = width
        self.sensitivity = sensitivity
        self.shape = deque(shape)
        self.orig_shape = shape
        self.header_start = None

    def recv_group(self, group) -> Tuple[bool, int]:
        if len(self.shape) == 0:
            return (True, self.header_start)

        if group.phase == self.shape[0] and \
            group.nsamples > self.width - self.sensitivity and \
                group.nsamples < self.width + self.sensitivity:

            if self.header_start is None:
                self.header_start = group.start

            self.shape.popleft()
        else:
            self._reset()

        if len(self.shape) == 0:
            return (True, self.header_start)
        else:
            return (False, None)

    def _reset(self):
        self.shape = deque(self.orig_shape)
        self.header_start = None

    def acknowledge(self):
        self._reset()


class Capture():
    def __init__(self):
        """
        A concrete representation of successfully identified RF data.
        """
        self.preamble_pos = None
        self.header_pos = None
        self.data = None

    def append(self, data: str):
        if self.data is None:
            self.data = ''

        self.data += data

    def __repr__(self):
        return '<Capture preamble@{} header@{} data={}>'.format(
            self.preamble_pos, self.header_pos, self.data)


class RawDecoder(object):
    def __init__(self, bit_width: int, sensitivity: int) -> None:
        """
        Implements the Decoder interface. Decoders intend to transform
        a stream of SampleGroups into binary data.

        This decoder, RawDecoder decodes NEG phase pulses as `0` and POS phase
        pulses as `1`. Pulses of length 1 emit `X`, pulses of length 2 emit
        `XX`, where X is the decoded phase.
        """
        if bit_width + sensitivity >= bit_width * 2 - sensitivity:
            raise Exception("Sensitivity is too large for the provided bit width.")

        self.sensitivity = sensitivity
        self.bit_width = bit_width

    def recv_group(self, group: SampleGroup, capture: Capture):
        width = self.get_width(group)
        polarity = self.get_polarity(group)
        capture.append(polarity * width)

    def get_width(self, group: SampleGroup) -> int:
        if group.nsamples > self.bit_width - self.sensitivity and \
                group.nsamples < self.bit_width + self.sensitivity:
            return 1
        elif group.nsamples > (self.bit_width * 2) - self.sensitivity and \
                group.nsamples < (self.bit_width * 2) + self.sensitivity:
            return 2

        raise Exception(
            'Found a bit that was too wide at {}'.format(group.start))

    def get_polarity(self, group: SampleGroup) -> str:
        if group.phase == Phase.POS:
            return '1'
        elif group.phase == Phase.NEG:
            return '0'

        raise Exception("Unknown phase during data decoding.")
