#!/bin/bash

TARGET_DIR=$1
NUM_FILES=$2
FILE_TO_COPY=$3


# Check if the directory exists
if [ -d "$TARGET_DIR" ]; then
    echo "Directory $TARGET_DIR exists. Changing to it."
else
    echo "Directory $TARGET_DIR does not exist. Creating it."
    mkdir -p "$TARGET_DIR"
fi

# Change to the target directory
cd "$TARGET_DIR" || { echo "Failed to change directory to $TARGET_DIR"; exit 1; }

for i in $(seq 1 "$NUM_FILES"); do
  dd if=/dev/urandom bs=5000 count=1 of=file"$i";
done

# All arguments as an array
args=("$@")

for (( i=3; i<${#args[@]}; i++ )); do
  file="${args[i]}"
  cp "$FILE_TO_COPY" "$file"
  echo "$FILE_TO_COPY copied to $file"
done
