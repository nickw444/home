#!/bin/sh

REMOTE_TMP_DIR="/tmp";
REMOTE_BIN_DIR="/sprinkle/bin";
REMOTE_HOST="sprinkle";
REMOTE_USER="sprinkle";

# Build an ARM binary.
build() {
    out_binary_name=$1;
    input_path=$2;
    GOOS=linux GOARCH=arm GOARM=5 go build -o $out_binary_name $input_path;
}

# Upload a ARM build binary
upload() {
    binary_name=$1;
    upload_path="$REMOTE_TMP_DIR/$binary_name"
    scp $binary_name sprinkle@sprinkle:$upload_path;
}

# Replace the existing running binary with one that has just been uploaded.
replace() {
    binary_name="$1";
    proc_name="$2";

    ssh "$REMOTE_USER@$REMOTE_HOST" << EOF
        set -eux
        supervisorctl stop $proc_name;
        mv "$REMOTE_TMP_DIR/$binary_name" "$REMOTE_BIN_DIR/$binary_name";
        supervisorctl start $proc_name;
EOF
}

_usage() {
  cat << EOF
Usage $(basename $0):
  -n    Name of the package. Default to basename of current working dir
  -s    Source dir of the package. Default to current working dir
EOF
  exit 1
}

set -eu
name=$(basename $PWD)
srcdir=$(pwd)

while getopts 'n:s:h' OPTION ;
  do
    case "$OPTION" in
      n)
        name="$OPTARG"
        ;;
      s)
        srcdir="$OPTARG"
        ;;
      h)
        _usage
        ;;
      ?)
        _usage
        ;;
    esac
done

set -eux
binary_name="$name-arm"
build "$binary_name" "$srcdir";
upload "$binary_name";
replace "$binary_name" "$name";
