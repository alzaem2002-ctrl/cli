# Running the GitHub SDK Tutorial on iPad Air

This guide provides step-by-step instructions for running the Jupyter Notebook tutorial (`manage-prompts-with-sdk-tutorial.ipynb`) on your iPad Air using iOS Python environments.

## Overview

The tutorial demonstrates programmatic interaction with GitHub using the PyGithub SDK. It's fully compatible with iPad Air when using appropriate Python environments.

## Required Apps for iPad Air

Since iPad doesn't support native command-line Python installation, you'll need to use one of these apps:

### Option 1: Juno (Recommended)
- **App**: [Juno - Jupyter for iOS](https://apps.apple.com/app/juno-jupyter-ide/id1462586500)
- **Features**: Native Jupyter notebook support with pip package installation
- **Cost**: Paid app (~$15)
- **Best for**: Complete offline functionality

### Option 2: Juno Connect
- **App**: [Juno Connect](https://apps.apple.com/app/juno-connect/id1315744137)
- **Features**: Connect to remote Jupyter servers
- **Cost**: Paid app (~$10)
- **Best for**: Connecting to cloud-based Jupyter instances or local servers

### Option 3: Carnets
- **App**: [Carnets - Jupyter](https://apps.apple.com/app/carnets-jupyter/id1450994949)
- **Features**: Free Jupyter notebook app with scipy support
- **Cost**: Free
- **Best for**: Budget-friendly option

### Option 4: Pythonista
- **App**: [Pythonista 3](https://apps.apple.com/app/pythonista-3/id1085978097)
- **Features**: Full Python IDE for iOS
- **Cost**: Paid app (~$10)
- **Note**: Requires manual notebook conversion to .py scripts

## Setup Instructions

### Using Juno (Recommended Method)

1. **Install Juno from the App Store**
   - Search for "Juno - Jupyter for iOS"
   - Purchase and install the app

2. **Transfer the Tutorial Files**
   
   **Method A: Via Files App**
   - Download all tutorial files to your iPad:
     - `manage-prompts-with-sdk-tutorial.ipynb`
     - `example_usage.py`
     - `tutorial-requirements.txt`
     - `TUTORIAL_README.md`
   - Open Files app on iPad
   - Navigate to iCloud Drive or local storage
   - Create a folder named "GitHub-Tutorial"
   - Move all downloaded files to this folder
   - Open Juno app
   - Navigate to the "GitHub-Tutorial" folder

   **Method B: Via GitHub**
   - Open Juno app
   - Use built-in file browser to clone or download from GitHub
   - Navigate to: `https://github.com/alzaem2002-ctrl/cli`
   - Download the tutorial files directly

3. **Install Required Packages**
   
   Juno has a built-in package manager. To install dependencies:
   
   - Open Juno app
   - Tap on "Settings" (gear icon)
   - Select "Package Manager"
   - Install the following packages:
     - `PyGithub` (version 2.1.1 or higher)
     - `requests` (version 2.31.0 or higher)
   
   Note: Jupyter and notebook support are built into Juno, so you don't need to install them separately.

4. **Set Up GitHub Token**
   
   Since iOS doesn't use traditional environment variables, you'll need to modify the notebook:
   
   - Open `manage-prompts-with-sdk-tutorial.ipynb` in Juno
   - In the first code cell, replace:
     ```python
     token = os.environ.get('GITHUB_TOKEN', 'your_token_here')
     ```
   
   With:
     ```python
     token = 'your_personal_access_token_here'
     ```
   
   **Security Note**: For better security, consider using Juno's secure storage or a password manager app.

5. **Run the Verification Script**
   
   Before running the full notebook:
   - Open `example_usage.py` in Juno
   - Modify the token handling as described above
   - Run the script to verify your setup

6. **Run the Tutorial**
   
   - Open `manage-prompts-with-sdk-tutorial.ipynb`
   - Execute cells one by one by tapping the "Play" button
   - Follow the instructions in each cell

### Using Juno Connect (Remote Server)

1. **Set Up a Remote Jupyter Server**
   
   **Option A: Use Google Colab (Free)**
   - Go to [colab.research.google.com](https://colab.research.google.com)
   - Upload the notebook file
   - Install PyGithub: `!pip install PyGithub requests`
   - Run the tutorial

   **Option B: Use Your Own Server**
   - Install Jupyter on a computer or cloud instance
   - Start Jupyter with: `jupyter notebook --ip=0.0.0.0`
   - Note the URL and token

2. **Connect from iPad**
   - Open Juno Connect
   - Tap "Add Server"
   - Enter server URL and token
   - Navigate to your tutorial files
   - Run the notebook

### Using Carnets (Free Option)

1. **Install Carnets from the App Store**
   - Search for "Carnets - Jupyter"
   - Install the free app

2. **Transfer Files**
   - Use the same file transfer methods as described for Juno
   - Open Carnets app
   - Navigate to your files using the built-in browser

3. **Install Packages**
   - Carnets comes with many packages pre-installed
   - To install PyGithub, create a new notebook and run:
     ```python
     %pip install PyGithub requests
     ```

4. **Configure Token**
   - Same as Juno instructions above
   - Modify the notebook to include your token directly

5. **Run Tutorial**
   - Open and execute the notebook cells

## iPad-Specific Tips

### Touch Interface
- **Running Cells**: Tap the play button or use the toolbar
- **Editing Code**: Tap any code cell to bring up the keyboard
- **Selecting Code**: Long-press to select text
- **Undo**: Use three-finger swipe left (iOS gesture)

### Keyboard Shortcuts (with external keyboard)
- `Shift + Enter`: Run cell and move to next
- `Ctrl + Enter`: Run cell and stay
- `Tab`: Auto-complete (in supported apps)
- `Esc`: Exit edit mode

### File Management
- Use the Files app for easy file organization
- Keep tutorial files in iCloud Drive for backup
- Export completed notebooks to PDF or HTML for sharing

### Token Security
Since you need to embed tokens in the code on iPad:
- Create a token with minimal required permissions
- Use separate tokens for testing vs. production
- Store tokens in a secure notes app or password manager
- Don't share notebooks containing tokens
- Consider using GitHub's fine-grained personal access tokens

### Limitations on iPad
1. **Environment Variables**: iOS doesn't support traditional environment variables
   - **Solution**: Modify code to use direct token assignment
2. **Command Line**: No native terminal access
   - **Solution**: Use in-app package managers
3. **File System**: Limited compared to desktop
   - **Solution**: Use Files app and iCloud Drive
4. **Background Processing**: Apps may suspend
   - **Solution**: Keep app active during long operations

## Troubleshooting

### "Package Not Found" Error
- Ensure you're using the correct package manager in your app
- Try restarting the app and reinstalling
- Check if the package name is correct (case-sensitive)

### "Authentication Failed" Error
- Verify your GitHub token is valid
- Check token permissions include `repo`, `read:org`, `read:user`
- Ensure there are no extra spaces in the token string

### "Import Error" for PyGithub
- Confirm PyGithub is installed: `%pip list | grep -i github`
- Reinstall if needed: `%pip install --upgrade PyGithub`

### Notebook Won't Load
- Check file wasn't corrupted during transfer
- Verify it's a valid .ipynb file
- Try opening in another app or re-downloading

### Connection Issues (Juno Connect)
- Verify server is running and accessible
- Check firewall settings on server
- Ensure you're on the same network (or using proper port forwarding)

## Alternative: Using a Desktop Environment

If iPad apps don't meet your needs, consider:

1. **Use iPad as Remote Terminal**
   - Install Termius or Blink Shell
   - SSH to a computer or cloud instance
   - Run Jupyter remotely and access via browser

2. **GitHub Codespaces**
   - Open repository in GitHub Codespaces
   - Access via Safari on iPad
   - Full desktop-like experience in browser

3. **Cloud Jupyter Services**
   - Google Colab (free)
   - Kaggle Kernels (free)
   - Azure Notebooks
   - Binder (launch from GitHub)

## Testing Your Setup

Run this code snippet to verify everything works:

```python
# Test 1: Import libraries
try:
    from github import Github
    import requests
    print("✓ All libraries imported successfully")
except ImportError as e:
    print(f"✗ Import error: {e}")

# Test 2: Check PyGithub version
from github import __version__ as gh_version
print(f"✓ PyGithub version: {gh_version}")

# Test 3: Test GitHub connection (requires token)
try:
    token = 'your_token_here'  # Replace with your token
    g = Github(token)
    user = g.get_user()
    print(f"✓ Successfully authenticated as: {user.login}")
except Exception as e:
    print(f"✗ Authentication failed: {e}")
```

## Next Steps

Once your environment is set up:

1. Read through `TUTORIAL_README.md` for an overview
2. Run `example_usage.py` to verify your setup
3. Open and work through `manage-prompts-with-sdk-tutorial.ipynb`
4. Experiment with the code examples
5. Try modifying examples to work with your own repositories

## Support and Resources

- **Tutorial Issues**: Open an issue in this repository
- **PyGithub Docs**: https://pygithub.readthedocs.io/
- **GitHub API Docs**: https://docs.github.com/en/rest
- **Juno Support**: https://juno.sh/support/
- **Carnets Support**: https://github.com/holzschu/Carnets

## Version Compatibility

This tutorial is tested with:
- **Python**: 3.8+ (3.12.3 recommended)
- **PyGithub**: 2.1.1+ (2.8.1 recommended)
- **Jupyter Notebook**: 7.0.0+
- **iOS**: 14.0+ (latest recommended)

---

**Note**: This tutorial was adapted specifically for iPad Air users. For desktop/laptop users, please refer to `TUTORIAL_README.md` for standard setup instructions.
