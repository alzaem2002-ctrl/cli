#!/bin/bash
#
# Quick start script for ComfyUI AnimateDiff Pro
#

echo "Starting ComfyUI AnimateDiff Pro..."
echo ""

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "Error: Virtual environment not found!"
    echo "Please run ./setup.sh first"
    exit 1
fi

# Activate virtual environment
source venv/bin/activate

# Check if ComfyUI is installed
if [ ! -d "ComfyUI" ]; then
    echo "Error: ComfyUI not found!"
    echo "Please run ./setup.sh first"
    exit 1
fi

# Run the main script
python run_animatediff.py
