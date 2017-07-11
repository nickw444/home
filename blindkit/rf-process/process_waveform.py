import wave
import sys
from process_lib import (FrameReader, WaveformClassifier, PhaseGrouper,
                         PreambleDetector, HeaderDetector, RawDecoder, Capture)
from enum import Enum
from typing import List, Iterator # noqa
import json

import logging
log = logging.getLogger(__name__)


"""
Params
R Captures:
frame_reader = FrameReader(f, no_channels=f.getnchannels(), channel=0)
classifier = WaveformClassifier(threshold=1000)
phase_grouper = PhaseGrouper(frame_reader, classifier, sensitivity=3)
preamble_detector = PreambleDetector(width=30, sensitivity=10, nrepeats=10)
header_detector = HeaderDetector(width=119, sensitivity=20)
decoder = RawDecoder(bit_width=30, sensitivity=10)
"""

class Stage(Enum):
    PREAMBLE = 'PREAMBLE'
    HEADER = 'HEADER'
    PAYLOAD = 'PAYLOAD'

def process_file(filename: str) -> Iterator[Capture]:
    print(f'Opening File: {filename}')

    f = wave.open(filename)
    print("No Frames: ", f.getnframes())
    print("No Channels: ", f.getnchannels())
    print("Sample Rate: ", f.getframerate())
    print("Sample Size: ", f.getsampwidth())
    print("Compression: ", f.getcompname())

    frame_reader = FrameReader(f, no_channels=f.getnchannels(), channel=0)
    classifier = WaveformClassifier(threshold=1000)
    phase_grouper = PhaseGrouper(frame_reader, classifier, sensitivity=3)
    preamble_detector = PreambleDetector(width=30, sensitivity=10, nrepeats=10)
    header_detector = HeaderDetector(width=119, sensitivity=20)
    decoder = RawDecoder(bit_width=30, sensitivity=10)

    detection_stage = Stage.PREAMBLE
    curr_capture = None
    next_capture = Capture()

    for group in phase_grouper.get_grouped_samples():
        if detection_stage == Stage.PREAMBLE:
            preamble_found, found_at = preamble_detector.recv_group(group)
            if preamble_found:
                log.debug('Found preamble @ {}'.format(found_at))
                next_capture.preamble_pos = found_at
                preamble_detector.acknowledge()
                detection_stage = Stage.HEADER

        elif detection_stage == Stage.HEADER:
            header_found, found_at = header_detector.recv_group(group)
            if header_found:
                log.debug('Found header @ {}'.format(found_at))
                next_capture.header_pos = found_at
                header_detector.acknowledge()

                if curr_capture is not None:
                    log.debug('End of curr capture due to new header @ {}'.format(found_at))
                    yield curr_capture

                curr_capture = next_capture
                next_capture = Capture()

                # Header found - go back to searching for preamble.
                detection_stage = Stage.PREAMBLE
                continue

        if curr_capture is not None:
            # Now that we've found a preamble, we should begin looking for the
            # actual data. Attempt to find the data.
            try:
                decoder.recv_group(group, curr_capture)
            except Exception as e:
                log.debug('End of curr capture due to decoder exception @ {}'.format(group.start))
                yield curr_capture
                curr_capture = None


def main(outfile: str, files: List[str]):
    output = []
    for filename in files:
        captures = [] # type: List[dict]
        recording = {
            'filename': filename,
            'captures': captures
        }
        last = None
        for capture in process_file(filename):
            if capture.data:
                if len(capture.data) < 84 or capture.data[:84] == last:
                    continue

                last = capture.data[:84]
            captures.append(serialize_capture(capture))

        output.append(recording)

    json.dump(output, open(outfile, 'w'), indent=2)

def serialize_capture(capture: Capture) -> dict:
    return {
        'data': capture.data,
        'header_pos': capture.header_pos,
        'preamble_pos': capture.preamble_pos
    }


if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    main(sys.argv[1], sys.argv[2:])
