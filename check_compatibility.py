#!/usr/bin/env python3
"""
Compatibility checker for the GitHub SDK tutorial.
This script verifies that all requirements are met for running the tutorial
on various platforms including iPad Air.
"""

import sys
import platform
import json

def check_python_version():
    """Check if Python version meets requirements."""
    version = sys.version_info
    required = (3, 7)
    
    if version >= required:
        print(f"✓ Python {version.major}.{version.minor}.{version.micro} (requirement: {required[0]}.{required[1]}+)")
        return True
    else:
        print(f"✗ Python {version.major}.{version.minor}.{version.micro} (requirement: {required[0]}.{required[1]}+)")
        return False

def check_package(package_name, import_name=None):
    """Check if a package is installed."""
    if import_name is None:
        import_name = package_name
    
    try:
        mod = __import__(import_name)
        version = getattr(mod, '__version__', 'unknown')
        print(f"✓ {package_name} installed (version: {version})")
        return True
    except ImportError:
        print(f"✗ {package_name} not installed")
        return False

def check_notebook_file():
    """Check if the notebook file exists and is valid."""
    try:
        with open('manage-prompts-with-sdk-tutorial.ipynb', 'r') as f:
            notebook = json.load(f)
        
        cell_count = len(notebook.get('cells', []))
        print(f"✓ Notebook file is valid ({cell_count} cells)")
        return True
    except FileNotFoundError:
        print("✗ Notebook file not found")
        return False
    except json.JSONDecodeError:
        print("✗ Notebook file is corrupted (invalid JSON)")
        return False

def check_platform_compatibility():
    """Provide platform-specific guidance."""
    system = platform.system()
    machine = platform.machine()
    
    print(f"\nPlatform: {system} ({machine})")
    
    if system == "Darwin" and "iPad" in machine:
        print("\n📱 iPad Detected!")
        print("   → See IPAD_SETUP_GUIDE.md for detailed setup instructions")
        print("   → Recommended apps: Juno, Juno Connect, or Carnets")
        print("   → Note: You'll need to modify token handling for iOS")
    elif system == "Darwin":
        print("\n🍎 macOS Detected")
        print("   → Standard setup should work fine")
        print("   → Install packages with: pip3 install -r tutorial-requirements.txt")
    elif system == "Linux":
        print("\n🐧 Linux Detected")
        print("   → Standard setup should work fine")
        print("   → Install packages with: pip3 install -r tutorial-requirements.txt")
    elif system == "Windows":
        print("\n🪟 Windows Detected")
        print("   → Standard setup should work fine")
        print("   → Install packages with: pip install -r tutorial-requirements.txt")
    else:
        print(f"\n❓ Unknown platform: {system}")
        print("   → The tutorial should still work if Python and pip are available")

def main():
    print("=" * 70)
    print("GitHub SDK Tutorial - Compatibility Check")
    print("=" * 70)
    print()
    
    all_checks = []
    
    # Check Python version
    all_checks.append(check_python_version())
    print()
    
    # Check required packages
    print("Checking required packages:")
    all_checks.append(check_package("PyGithub", "github"))
    all_checks.append(check_package("requests"))
    print()
    
    # Check optional packages (Jupyter)
    print("Checking optional packages (for desktop use):")
    jupyter_installed = check_package("jupyter")
    notebook_installed = check_package("notebook")
    
    if not jupyter_installed or not notebook_installed:
        print("\n⚠️  Jupyter not installed (optional for desktop, required for iPad apps)")
        print("   Install with: pip3 install jupyter notebook")
    print()
    
    # Check notebook file
    print("Checking tutorial files:")
    all_checks.append(check_notebook_file())
    
    try:
        with open('TUTORIAL_README.md', 'r') as f:
            print("✓ TUTORIAL_README.md found")
    except FileNotFoundError:
        print("⚠️  TUTORIAL_README.md not found")
    
    try:
        with open('IPAD_SETUP_GUIDE.md', 'r') as f:
            print("✓ IPAD_SETUP_GUIDE.md found")
    except FileNotFoundError:
        print("⚠️  IPAD_SETUP_GUIDE.md not found (needed for iPad users)")
    
    try:
        with open('tutorial-requirements.txt', 'r') as f:
            print("✓ tutorial-requirements.txt found")
    except FileNotFoundError:
        print("⚠️  tutorial-requirements.txt not found")
    
    print()
    
    # Platform-specific guidance
    check_platform_compatibility()
    
    print()
    print("=" * 70)
    
    if all(all_checks):
        print("✅ All critical requirements met! You're ready to run the tutorial.")
        print()
        print("Next steps:")
        print("1. Set your GITHUB_TOKEN environment variable")
        print("2. Run: python3 example_usage.py")
        print("3. Open the notebook: manage-prompts-with-sdk-tutorial.ipynb")
        print()
        print("For iPad users: See IPAD_SETUP_GUIDE.md for detailed instructions")
    else:
        print("⚠️  Some requirements are missing. Please install them first.")
        print()
        print("Quick fix:")
        print("  pip3 install -r tutorial-requirements.txt")
        print()
        print("For iPad users: Use your app's package manager to install packages")
    
    print("=" * 70)

if __name__ == "__main__":
    main()
