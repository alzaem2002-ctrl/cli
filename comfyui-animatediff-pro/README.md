# ComfyUI AnimateDiff Pro

Professional video generation system from images using AnimateDiff and ComfyUI.

## 🌟 Features

- **High-Quality Video Generation**: Create stunning cinematic videos from text prompts
- **Full Motion Control**: Complete control over animation and motion parameters
- **No Length Restrictions**: Generate videos of any length
- **Automated Workflow**: Seamless integration with Cursor Agent
- **Professional Results**: Using Realistic Vision V6.0 for photorealistic output

## 📋 Requirements

- **Python**: 3.10 or higher
- **Disk Space**: At least 50GB free space
- **GPU**: NVIDIA GPU with CUDA support (recommended)
- **RAM**: 16GB minimum, 32GB recommended
- **Operating System**: Linux, macOS, or Windows with WSL

## 🚀 Quick Start

### 1. Clone or Download

```bash
# If this is a git repository
git clone <repository-url>
cd comfyui-animatediff-pro
```

### 2. Run Setup

```bash
chmod +x setup.sh
./setup.sh
```

This will:
- Check your environment
- Create a virtual environment
- Install ComfyUI
- Install AnimateDiff Evolved and required extensions
- Download necessary models (Motion Module, Checkpoint, VAE)

**Note**: Model downloads may take a while depending on your internet connection (several GB of data).

### 3. Start the System

```bash
./start.sh
```

Or manually:

```bash
source venv/bin/activate
python run_animatediff.py
```

### 4. Access ComfyUI

Open your browser and navigate to:
```
http://127.0.0.1:8188
```

### 5. Load Workflow

1. In the ComfyUI interface, click on "Load" or drag and drop `workflow.json`
2. The pre-configured AnimateDiff workflow will be loaded
3. Adjust parameters as needed
4. Click "Queue Prompt" to generate your video

## 📁 Project Structure

```
comfyui-animatediff-pro/
├── setup.sh                 # Installation script
├── start.sh                 # Quick start script
├── run_animatediff.py       # Main execution script
├── workflow.json            # Pre-configured workflow
├── README.md                # This file
├── venv/                    # Python virtual environment
├── ComfyUI/                 # ComfyUI installation
│   ├── models/
│   │   ├── checkpoints/     # SD checkpoints
│   │   ├── vae/             # VAE models
│   │   └── animatediff_models/  # AnimateDiff motion modules
│   ├── custom_nodes/
│   │   ├── ComfyUI-AnimateDiff-Evolved/
│   │   ├── ComfyUI-Advanced-ControlNet/
│   │   └── ComfyUI-VideoHelperSuite/
│   └── output/              # Generated videos
```

## 🎨 Customizing Your Videos

### Edit Workflow Parameters

The `run_animatediff.py` script generates a default workflow. You can customize:

**Prompt (Node 2)**:
```json
"text": "stunning cinematic landscape, mountains, sunrise, 8k, masterpiece, highly detailed"
```

**Negative Prompt (Node 3)**:
```json
"text": "low quality, worst quality, blurry, watermark"
```

**Sampling Settings (Node 6)**:
- `steps`: Number of sampling steps (higher = better quality, slower)
- `cfg`: Classifier-free guidance scale (7.5 is default)
- `seed`: Random seed for reproducibility

**Animation Settings (Node 5)**:
- `context_length`: Number of frames processed together
- `motion_scale`: Strength of motion (0.0 to 2.0)
- `closed_loop`: Enable for looping videos

**Video Output (Node 8)**:
- `frame_rate`: Frames per second
- `format`: Output format (video/h264-mp4)
- `crf`: Compression quality (lower = better quality, larger file)

## 🔧 Advanced Usage

### Manual Model Downloads

If automatic downloads fail, manually download models:

**Motion Module**:
```bash
wget -P ComfyUI/models/animatediff_models/ \
  https://huggingface.co/guoyww/animatediff/resolve/main/v3_sd15_mm.safetensors
```

**Checkpoint**:
```bash
wget -P ComfyUI/models/checkpoints/ \
  https://huggingface.co/SG161222/Realistic_Vision_V6.0_B1_noVAE/resolve/main/Realistic_Vision_V6.0_B1_fp16.safetensors
```

**VAE**:
```bash
wget -P ComfyUI/models/vae/ \
  https://huggingface.co/stabilityai/sd-vae-ft-mse-original/resolve/main/vae-ft-mse-840000-ema-pruned.safetensors
```

### Running in the Background

```bash
nohup python run_animatediff.py > comfyui.log 2>&1 &
```

### API Usage

Once running, you can use the API:

```bash
# Check system stats
curl http://127.0.0.1:8188/system_stats

# Submit workflow via API
curl -X POST http://127.0.0.1:8188/prompt \
  -H "Content-Type: application/json" \
  -d @workflow.json
```

## 🤖 Cursor Agent Integration

### Prerequisites

Enable in your Cursor settings:

1. **Settings > Features > Agent Mode** → Enable
2. **Settings > Beta > YOLO Mode** → Enable  
3. **Command Allowlist**: `git, pip, python, mkdir, wget, curl`
4. **Model**: Claude 3.5 Sonnet (latest) or newer
5. **Enable**: Full Codebase Context

### Using with Cursor Agent

1. Copy the entire content of this README
2. Open Cursor Composer (`Cmd+I` or `Ctrl+I`)
3. Select **Agent Mode**
4. Paste the instructions
5. Press Enter and let Cursor work

## 🐛 Troubleshooting

### CUDA Not Available

If PyTorch doesn't detect CUDA:
```bash
pip uninstall torch torchvision torchaudio
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
```

### Out of Memory Errors

Reduce these parameters in the workflow:
- Lower `context_length` from 16 to 8
- Reduce `steps` from 20 to 15
- Use a smaller checkpoint model

### Server Won't Start

Check if port 8188 is already in use:
```bash
lsof -i :8188
kill -9 <PID>  # If needed
```

### Missing Models

Verify models are present:
```bash
ls -lh ComfyUI/models/checkpoints/
ls -lh ComfyUI/models/animatediff_models/
ls -lh ComfyUI/models/vae/
```

## 📚 Resources

- [ComfyUI Documentation](https://github.com/comfyanonymous/ComfyUI)
- [AnimateDiff Evolved](https://github.com/Kosinkadink/ComfyUI-AnimateDiff-Evolved)
- [Realistic Vision Model](https://huggingface.co/SG161222/Realistic_Vision_V6.0_B1_noVAE)

## 📄 License

This project uses several open-source components. Please refer to their respective licenses:
- ComfyUI: GPL-3.0
- AnimateDiff: MIT
- Model licenses vary by model

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## ⚠️ Notes

- First run will take longer as models are loaded into memory
- GPU with at least 8GB VRAM recommended for optimal performance
- Generated videos are saved in `ComfyUI/output/`
- Check `workflow.json` for the complete node configuration

## 📞 Support

For issues or questions:
1. Check the troubleshooting section
2. Review ComfyUI and AnimateDiff documentation
3. Open an issue in this repository

---

**Happy Video Generation! 🎬**
