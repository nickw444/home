"""
A tool to turn processed data from process_waveform.py
JSON output into tabulated CSV data, ideal for visual
analysis using Microsoft Excel.
"""

import json
import csv
import sys
import re
from collections import namedtuple
from enum import Enum
from typing import List, Iterator
import os


data = json.load(open(sys.argv[1]))
outfile = open(sys.argv[2], 'w')

class CaptureIdentifier(object):
    """
    Inspects raw capture data and identifies what action is being
    performed by the remote.
    """

    class Kind():
        UP = 'UP'
        DOWN = 'DOWN'
        STOP = 'STOP'
        PAIR = 'PAIR'

    # Bits to Kind mapping
    KIND_MAP = {
        '0101': Kind.DOWN,
        '0110': Kind.UP,
        '1001': Kind.STOP,
        '1010': Kind.PAIR
    }

    # It appears there is a checksum encoded into the payload for
    # each action. We should check this to using this map.
    KIND_CHECK_MAP = {
        Kind.UP: '10',
        Kind.STOP: '10',
        Kind.DOWN: '10',
        Kind.PAIR: '01',
    }

    Identification = namedtuple('Identification', ['kind'])

    def identify_bits(self, data: str) -> Identification:
        kind = CaptureIdentifier.KIND_MAP[data[50:54]]
        if kind is not None:
            check_bits = CaptureIdentifier.KIND_CHECK_MAP[kind]
            if check_bits != data[64:66]:
                kind = None

        return CaptureIdentifier.Identification(kind)


class RawBitProcessor(object):
    """
    An identity bit processor.
    """
    def process_bits(self, data: str) -> str:
        return data


class ManchesterBitProcessor(object):
    """
    A bit processor that decodes a stream of raw data using
    the manchester encoding scheme.
    """
    def process_bits(self, data: str) -> str:
        out = ''
        for pair in self.split_pairs(data):
            if pair[0] == '1' and pair[1] == '0':
                out += '1'
            elif pair[0] == '0' and pair[1] == '1':
                out += '0'
            else:
                raise Exception("Not valid manchester encoded data")

        return out

    def split_pairs(self, data: str) -> Iterator[str]:
        accum = ''
        for char in data:
            accum += char
            if len(accum) == 2:
                yield accum
                accum = ''


class PWMBitProcessor(object):
    """
    A bit processor that decodes a stream of raw data using
    a PWM encoding scheme.

    A sequence of length 1 emits a 0, a sequence of length
    2 emits a 1
    """
    def process_bits(self, data: str) -> str:
        run_length = 0
        curr = None
        out = ''
        for char in data:
            if char != curr:
                # Phase change occured.
                if run_length == 1:
                    out += '0'
                elif run_length == 2:
                    out += '1'
                elif run_length == 0:
                    pass
                else:
                    raise Exception("Not valid PWM data.")

                curr = char
                run_length = 0

            run_length += 1

        return out


identifier = CaptureIdentifier()
processor = RawBitProcessor()
# processor = PWMBitProcessor()
# processor = ManchesterBitProcessor()


tabulated = []  # type: List[dict]

filename_re = re.compile(r'([A-Z][0-9])_(CH)?([0-9])(_.*)?\.wav$')

for capture_file in data:
    filename = os.path.basename(capture_file['filename'])
    res = filename_re.match(filename)
    remote = res.group(1)
    channel = res.group(3)

    for capture in capture_file['captures']:
        data = capture['data']
        if data is None:
            continue

        ident = identifier.identify_bits(data)

        tabulated.append({
            'channel': channel,
            'remote': remote,
            'action': ident.kind,
            'data': processor.process_bits(data)
        })



writer = csv.writer(outfile)
writer.writerow(['Remote', 'Channel', 'Action', 'Data'])
for row in tabulated:
    # print(','.join(row['data']))
    writer.writerow([
        row['remote'],
        row['channel'],
        row['action'],
        ','.join(row['data'])
    ])

