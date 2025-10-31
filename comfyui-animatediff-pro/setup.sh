#!/bin/bash
#
# ComfyUI AnimateDiff Pro - Installation Script
# Automated setup for video generation system
#

set -e  # Exit on error

echo "=========================================="
echo "ComfyUI AnimateDiff Pro Setup"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Stage 1: Check environment
echo "Stage 1: Checking environment..."
echo ""

# Check Python
if ! command -v python3 &> /dev/null; then
    print_error "Python 3 is not installed"
    exit 1
fi
PYTHON_VERSION=$(python3 --version)
print_info "Python version: $PYTHON_VERSION"

# Check pip
if ! command -v pip3 &> /dev/null; then
    print_error "pip is not installed"
    exit 1
fi
print_info "pip: $(pip3 --version)"

# Check disk space
DISK_SPACE=$(df -h . | awk 'NR==2 {print $4}')
print_info "Available disk space: $DISK_SPACE"

# Check Git
if ! command -v git &> /dev/null; then
    print_error "Git is not installed"
    exit 1
fi
print_info "Git: $(git --version)"

echo ""

# Stage 2: Create virtual environment (optional)
echo "Stage 2: Setting up Python environment..."
echo ""

if [ ! -d "venv" ]; then
    print_info "Creating virtual environment..."
    python3 -m venv venv
else
    print_info "Virtual environment already exists"
fi

# Activate virtual environment
print_info "Activating virtual environment..."
source venv/bin/activate

# Upgrade pip
print_info "Upgrading pip..."
pip install --upgrade pip setuptools wheel

echo ""

# Stage 3: Install ComfyUI
echo "Stage 3: Installing ComfyUI..."
echo ""

if [ ! -d "ComfyUI" ]; then
    print_info "Cloning ComfyUI..."
    git clone https://github.com/comfyanonymous/ComfyUI.git
    
    print_info "Installing ComfyUI requirements..."
    cd ComfyUI
    pip install -r requirements.txt
    
    # Install PyTorch (CUDA version)
    print_info "Installing PyTorch with CUDA support..."
    pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
    
    # Verify PyTorch installation
    print_info "Verifying PyTorch installation..."
    python -c "import torch; print(f'PyTorch version: {torch.__version__}'); print(f'CUDA available: {torch.cuda.is_available()}')"
    
    cd ..
else
    print_info "ComfyUI already installed"
fi

echo ""

# Stage 4: Install AnimateDiff Evolved
echo "Stage 4: Installing AnimateDiff Evolved..."
echo ""

cd ComfyUI/custom_nodes

if [ ! -d "ComfyUI-AnimateDiff-Evolved" ]; then
    print_info "Cloning AnimateDiff Evolved..."
    git clone https://github.com/Kosinkadink/ComfyUI-AnimateDiff-Evolved.git
else
    print_info "AnimateDiff Evolved already installed"
fi

if [ ! -d "ComfyUI-Advanced-ControlNet" ]; then
    print_info "Cloning Advanced ControlNet..."
    git clone https://github.com/Kosinkadink/ComfyUI-Advanced-ControlNet.git
else
    print_info "Advanced ControlNet already installed"
fi

if [ ! -d "ComfyUI-VideoHelperSuite" ]; then
    print_info "Cloning VideoHelperSuite..."
    git clone https://github.com/Kosinkadink/ComfyUI-VideoHelperSuite.git
else
    print_info "VideoHelperSuite already installed"
fi

cd ../..

echo ""

# Stage 5: Create directory structure
echo "Stage 5: Creating directory structure..."
echo ""

print_info "Creating model directories..."
mkdir -p ComfyUI/models/checkpoints
mkdir -p ComfyUI/models/vae
mkdir -p ComfyUI/models/animatediff_models
mkdir -p ComfyUI/models/animatediff_motion_lora
mkdir -p ComfyUI/custom_nodes/ComfyUI-AnimateDiff-Evolved/models
mkdir -p ComfyUI/custom_nodes/ComfyUI-AnimateDiff-Evolved/motion_lora
mkdir -p ComfyUI/output

print_info "Directory structure created"

echo ""

# Stage 6: Download models
echo "Stage 6: Downloading models..."
echo ""
print_warning "This may take a while depending on your internet connection"

# Download Motion Module
if [ ! -f "ComfyUI/models/animatediff_models/v3_sd15_mm.safetensors" ]; then
    print_info "Downloading Motion Module (v3_sd15_mm)..."
    wget -P ComfyUI/models/animatediff_models/ \
        https://huggingface.co/guoyww/animatediff/resolve/main/v3_sd15_mm.safetensors || \
        print_warning "Failed to download motion module. You can download it manually."
else
    print_info "Motion module already exists"
fi

# Download Checkpoint
if [ ! -f "ComfyUI/models/checkpoints/Realistic_Vision_V6.0_B1_fp16.safetensors" ]; then
    print_info "Downloading Realistic Vision V6.0 checkpoint..."
    print_warning "This is a large file (~3GB), please be patient..."
    wget -P ComfyUI/models/checkpoints/ \
        https://huggingface.co/SG161222/Realistic_Vision_V6.0_B1_noVAE/resolve/main/Realistic_Vision_V6.0_B1_fp16.safetensors || \
        print_warning "Failed to download checkpoint. You can download it manually."
else
    print_info "Checkpoint already exists"
fi

# Download VAE
if [ ! -f "ComfyUI/models/vae/vae-ft-mse-840000-ema-pruned.safetensors" ]; then
    print_info "Downloading VAE model..."
    wget -P ComfyUI/models/vae/ \
        https://huggingface.co/stabilityai/sd-vae-ft-mse-original/resolve/main/vae-ft-mse-840000-ema-pruned.safetensors || \
        print_warning "Failed to download VAE. You can download it manually."
else
    print_info "VAE already exists"
fi

echo ""

# Stage 7: Verify installation
echo "Stage 7: Verifying installation..."
echo ""

print_info "Checking models..."
ls -lh ComfyUI/models/checkpoints/ 2>/dev/null || print_warning "No checkpoints found"
ls -lh ComfyUI/models/animatediff_models/ 2>/dev/null || print_warning "No motion models found"
ls -lh ComfyUI/models/vae/ 2>/dev/null || print_warning "No VAE models found"

echo ""
echo "=========================================="
echo "Installation Complete!"
echo "=========================================="
echo ""
print_info "To start the system, run:"
echo "    source venv/bin/activate"
echo "    python run_animatediff.py"
echo ""
print_info "Or simply:"
echo "    ./start.sh"
echo ""
