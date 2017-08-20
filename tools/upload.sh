#!/bin/sh

REMOTE_TMP_DIR="/tmp";
REMOTE_BIN_DIR="/sprinkle/bin";
REMOTE_HOST="sprinkle";
REMOTE_USER="sprinkle";

# Build an ARM binary.
build() {
    binary_name=$1;
    GOOS=linux GOARCH=arm GOARM=5 go build -o $binary_name;
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


main() {
    set -eux
    binary_name="$1-arm";
    proc_name="$1";

    build "$binary_name"; 
    upload "$binary_name";
    replace "$binary_name" "$proc_name";
}


if [ "$#" -gt 1 ]; then
    echo "Usage $0 <name>";
    exit 1;
fi;

package=${1:-$(basename $PWD)}
echo "Performing upload and replace for $package"
main $package;
