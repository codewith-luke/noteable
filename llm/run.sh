#!/usr/bin/env bash
set -e

# Get the directory of the script
script_dir=$(dirname "$0")

# Change the working directory to the script directory
cd "$script_dir"

source "./venv/bin/activate"
python3 script.py -i "$1"