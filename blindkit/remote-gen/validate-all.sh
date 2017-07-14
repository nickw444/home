#!/bin/sh
# A script to run remote-gen across each code in a file and
# ensure that all codes are consistent with the self-verification
# built into remote-gen.
#
# This script is primarily used to ensure that the logic in remote-gen
# for generating checksum's and other code-checks is working.

if [ $# -ne 1 ]; then
    echo "Usage: $0 <codes.in.txt>"
    exit 1
fi

if [ ! -f "$1" ]; then
    echo "File $1 does not exist"
    exit 1
fi

while read code;
do
    ./remote-gen info "$code" --validate > /dev/null
    if [ $? -ne 0 ]; then
        echo "Failure whilst processing $code";
        exit 1
    fi
done < $1

echo "All codes successfully validated."
