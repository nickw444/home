# rf-process

A collection of tooling for converting 433MHz signal captures in WAV files to binary data.

- `process_waveform.py`: A tool to open a WAV file and search for possible RF transmissions and output digitised data as JSON.
- `tabulate.py`: Read JSON files captured by the previous tool, apply decoding, then tabulate it as CSV, ideal for visual analysis in a spreadsheet.
- [`captures/`](captures/) contains some codes that were captured from the remotes via `process_waveform.py`.
