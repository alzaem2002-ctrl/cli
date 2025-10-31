#!/usr/bin/env python3
"""
ComfyUI AnimateDiff Pro - Video Generation Script
Generates professional videos from images using AnimateDiff
"""

import json
import subprocess
import time
import sys
import os


def create_workflow():
    """Create the AnimateDiff workflow configuration"""
    workflow = {
        "1": {
            "inputs": {"ckpt_name": "Realistic_Vision_V6.0_B1_fp16.safetensors"},
            "class_type": "CheckpointLoaderSimple"
        },
        "2": {
            "inputs": {
                "text": "stunning cinematic landscape, mountains, sunrise, 8k, masterpiece, highly detailed",
                "clip": ["1", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "3": {
            "inputs": {
                "text": "low quality, worst quality, blurry, watermark",
                "clip": ["1", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "4": {
            "inputs": {"motion_model_name": "v3_sd15_mm.safetensors"},
            "class_type": "AnimateDiffLoaderV3"
        },
        "5": {
            "inputs": {
                "context_length": 16,
                "context_stride": 1,
                "context_overlap": 4,
                "closed_loop": False,
                "motion_scale": 1.0,
                "beta_schedule": "sqrt_linear",
                "motion_model": ["4", 0]
            },
            "class_type": "ADE_AnimateDiffOptions"
        },
        "6": {
            "inputs": {
                "seed": 42,
                "steps": 20,
                "cfg": 7.5,
                "sampler_name": "euler",
                "scheduler": "normal",
                "denoise": 1.0,
                "model": ["1", 0],
                "positive": ["2", 0],
                "negative": ["3", 0],
                "latent_image": ["5", 0]
            },
            "class_type": "KSampler"
        },
        "7": {
            "inputs": {
                "samples": ["6", 0],
                "vae": ["1", 2]
            },
            "class_type": "VAEDecode"
        },
        "8": {
            "inputs": {
                "images": ["7", 0],
                "frame_rate": 8,
                "format": "video/h264-mp4",
                "crf": 21,
                "filename_prefix": "animatediff"
            },
            "class_type": "VHS_VideoCombine"
        }
    }
    return workflow


def start_server():
    """Start the ComfyUI server"""
    print("🚀 Starting ComfyUI server...")
    
    # Check if ComfyUI directory exists
    if not os.path.exists("ComfyUI"):
        print("❌ Error: ComfyUI directory not found!")
        print("Please run setup.sh first to install ComfyUI")
        sys.exit(1)
    
    # Start the server
    subprocess.Popen([sys.executable, "ComfyUI/main.py"])
    print("⏳ Waiting for server to start...")
    time.sleep(10)


def save_workflow():
    """Save the workflow to a JSON file"""
    workflow = create_workflow()
    with open('workflow.json', 'w') as f:
        json.dump(workflow, f, indent=2)
    print("✓ Workflow saved to workflow.json")


def check_environment():
    """Check if the environment is properly set up"""
    print("🔍 Checking environment...")
    
    # Check Python version
    py_version = sys.version_info
    if py_version.major < 3 or (py_version.major == 3 and py_version.minor < 10):
        print(f"⚠️  Warning: Python {py_version.major}.{py_version.minor} detected. Python 3.10+ recommended.")
    else:
        print(f"✓ Python {py_version.major}.{py_version.minor} detected")
    
    # Check if ComfyUI exists
    if os.path.exists("ComfyUI"):
        print("✓ ComfyUI directory found")
    else:
        print("❌ ComfyUI not found. Run setup.sh first.")
        return False
    
    # Check for required models
    models_ok = True
    
    checkpoint_path = "ComfyUI/models/checkpoints/Realistic_Vision_V6.0_B1_fp16.safetensors"
    if os.path.exists(checkpoint_path):
        print("✓ Checkpoint model found")
    else:
        print("⚠️  Checkpoint model not found")
        models_ok = False
    
    motion_model_path = "ComfyUI/models/animatediff_models/v3_sd15_mm.safetensors"
    if os.path.exists(motion_model_path):
        print("✓ Motion model found")
    else:
        print("⚠️  Motion model not found")
        models_ok = False
    
    if not models_ok:
        print("\n⚠️  Some models are missing. Download them using setup.sh")
    
    return True


if __name__ == "__main__":
    print("=" * 60)
    print("ComfyUI AnimateDiff Pro - Video Generation System")
    print("=" * 60)
    print()
    
    if not check_environment():
        sys.exit(1)
    
    save_workflow()
    start_server()
    
    print()
    print("=" * 60)
    print("✓ ComfyUI is running at http://127.0.0.1:8188")
    print("=" * 60)
    print()
    print("📝 To use the workflow:")
    print("   1. Open http://127.0.0.1:8188 in your browser")
    print("   2. Load the workflow.json file")
    print("   3. Generate your video!")
    print()
    print("Press Ctrl+C to stop the server")
