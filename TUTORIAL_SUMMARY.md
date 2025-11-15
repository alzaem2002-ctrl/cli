# GitHub SDK Tutorial - Quick Start Guide

This document provides a quick overview of the GitHub SDK tutorial and how to get started on different platforms.

## 📚 What's Included

This repository now includes a complete tutorial for learning how to programmatically interact with GitHub using Python and the PyGithub SDK.

### Tutorial Files

| File | Purpose |
|------|---------|
| `manage-prompts-with-sdk-tutorial.ipynb` | Interactive Jupyter notebook with hands-on examples |
| `example_usage.py` | Quick test script to verify your environment |
| `check_compatibility.py` | Comprehensive compatibility checker |
| `tutorial-requirements.txt` | Python package dependencies |
| `TUTORIAL_README.md` | Main tutorial documentation |
| `IPAD_SETUP_GUIDE.md` | Complete guide for iPad Air users |

## 🚀 Quick Start

### For Desktop/Laptop Users

1. **Install dependencies:**
   ```bash
   pip install -r tutorial-requirements.txt
   ```

2. **Set your GitHub token:**
   ```bash
   export GITHUB_TOKEN=your_personal_access_token
   ```

3. **Test your setup:**
   ```bash
   python3 example_usage.py
   ```

4. **Run the tutorial:**
   ```bash
   jupyter notebook manage-prompts-with-sdk-tutorial.ipynb
   ```

5. **Full instructions:** See [TUTORIAL_README.md](TUTORIAL_README.md)

### For iPad Air Users

1. **Install a Jupyter app:**
   - [Juno](https://apps.apple.com/app/juno-jupyter-ide/id1462586500) (Recommended)
   - [Carnets](https://apps.apple.com/app/carnets-jupyter/id1450994949) (Free)
   - [Juno Connect](https://apps.apple.com/app/juno-connect/id1315744137) (For remote servers)

2. **Transfer tutorial files:**
   - Download files to iPad via Files app
   - Or use the app's built-in file browser

3. **Install packages:**
   - Use the app's package manager to install PyGithub and requests

4. **Configure authentication:**
   - Open the notebook in your app
   - Replace `'your_token_here'` with your actual GitHub token

5. **Full instructions:** See [IPAD_SETUP_GUIDE.md](IPAD_SETUP_GUIDE.md)

## 📱 iPad Air Compatibility

This tutorial is fully adapted for iPad Air with the following considerations:

### What Works on iPad

✅ **Full Jupyter Notebook Support**
- All code examples work identically on iPad
- Touch-friendly interface in supported apps
- Keyboard shortcuts with external keyboard

✅ **Package Installation**
- PyGithub, requests, and other dependencies
- Via app-specific package managers
- No need for command-line access

✅ **File Management**
- Integration with Files app
- iCloud Drive support
- Easy file sharing

### iPad-Specific Considerations

⚠️ **Environment Variables**
- iOS doesn't support traditional environment variables
- **Solution:** Modify the notebook to include your token directly
- Security guidance provided in the iPad Setup Guide

⚠️ **Command Line**
- No native terminal access
- **Solution:** Use app-based package managers
- Alternative: Use SSH apps to connect to remote servers

⚠️ **Background Processing**
- Apps may suspend when not active
- **Solution:** Keep app in foreground during execution
- Short-running operations complete before suspension

## 📖 What You'll Learn

The tutorial covers:

1. **Authentication** - Connecting to GitHub API with tokens
2. **Repository Access** - Reading repository information
3. **Issue Management** - Listing, searching, and creating issues
4. **Pull Request Operations** - Working with PRs programmatically
5. **Comments** - Reading and adding comments
6. **Search Queries** - Finding issues and PRs across GitHub
7. **Label Management** - Working with repository labels

## 🔒 Security Best Practices

### Token Security

**Desktop:**
- Use environment variables: `export GITHUB_TOKEN=token`
- Never commit tokens to version control
- Use `.env` files with `.gitignore`

**iPad:**
- Store tokens in password manager apps
- Don't share notebooks with embedded tokens
- Create tokens with minimal required permissions
- Use fine-grained personal access tokens

### Safe Operations

All write operations in the tutorial are **commented out by default**:
- Creating issues
- Adding comments
- Modifying labels

Uncomment only when:
- You have write access to the repository
- You understand the operation's impact
- You're in a test environment or working intentionally

## 🧪 Testing Your Setup

Run the compatibility checker to verify everything is ready:

```bash
python3 check_compatibility.py
```

This will check:
- ✅ Python version (3.7+ required, 3.12.3 recommended)
- ✅ Required packages (PyGithub, requests)
- ✅ Optional packages (jupyter, notebook)
- ✅ Tutorial file integrity
- ✅ Platform-specific guidance

## 🎯 Learning Path

### Beginner Level
1. Run `example_usage.py` to test your setup
2. Follow the notebook step-by-step
3. Execute each cell and read the output
4. Experiment with read-only operations

### Intermediate Level
1. Modify code examples for your own repositories
2. Try different search queries
3. Explore additional PyGithub features
4. Build simple automation scripts

### Advanced Level
1. Uncomment and test write operations (safely)
2. Create custom automation workflows
3. Integrate with other GitHub features
4. Build production tools

## 📚 Additional Resources

### Official Documentation
- [PyGithub Documentation](https://pygithub.readthedocs.io/)
- [GitHub REST API](https://docs.github.com/en/rest)
- [GitHub CLI](https://cli.github.com/manual/)

### iOS Python Apps
- [Juno Support](https://juno.sh/support/)
- [Carnets GitHub](https://github.com/holzschu/Carnets)
- [Pythonista Forum](https://forum.omz-software.com/category/5/pythonista)

### Getting Help
- **Tutorial Issues**: Open an issue in this repository
- **PyGithub Issues**: [PyGithub GitHub](https://github.com/PyGithub/PyGithub)
- **iPad App Issues**: Contact app support directly

## 🎓 Tutorial Structure

The notebook is organized into clear sections:

1. **Introduction** - Overview and prerequisites
2. **Setup** - Authentication and initialization
3. **Repositories** - Working with repository data
4. **Issues** - Managing issues programmatically
5. **Pull Requests** - PR operations
6. **Comments** - Reading and writing comments
7. **Search** - Finding issues and PRs
8. **Labels** - Label management
9. **Summary** - Recap and next steps

Each section includes:
- Clear explanations
- Working code examples
- Expected output samples
- Safety notes where applicable

## 💡 Tips for Success

### Desktop Users
1. Use virtual environments: `python -m venv venv`
2. Keep Jupyter updated: `pip install --upgrade jupyter`
3. Use Git for version control of your modifications
4. Backup your work regularly

### iPad Users
1. Use an external keyboard for efficiency
2. Keep apps updated for latest features
3. Store files in iCloud for automatic backup
4. Learn touch gestures for your app
5. Consider using iPad in landscape mode for more screen space

## 🐛 Troubleshooting

### Common Issues

**"Module 'github' not found"**
- Desktop: `pip install PyGithub`
- iPad: Install via app's package manager

**"Authentication failed"**
- Check token is valid and not expired
- Verify token has required permissions
- Ensure no extra spaces in token string

**"Notebook won't open"**
- Verify file wasn't corrupted during transfer
- Check it's a valid .ipynb file
- Try re-downloading

**iPad app crashes or freezes**
- Restart the app
- Check for app updates
- Free up iPad storage space
- Try a different app

For more troubleshooting help, see:
- Desktop: [TUTORIAL_README.md](TUTORIAL_README.md)
- iPad: [IPAD_SETUP_GUIDE.md](IPAD_SETUP_GUIDE.md)

## 🤝 Contributing

Found an issue or want to improve the tutorial?
1. Open an issue describing the problem or enhancement
2. Fork the repository
3. Make your changes
4. Submit a pull request

## 📄 License

This tutorial follows the same license as the main repository (MIT License).

## 🙏 Acknowledgments

- Based on tutorial from PR #13
- Adapted for iPad Air compatibility
- Uses the excellent PyGithub library
- Inspired by the GitHub CLI project

---

**Ready to get started?** Choose your platform and follow the appropriate guide:
- 💻 Desktop: [TUTORIAL_README.md](TUTORIAL_README.md)
- 📱 iPad: [IPAD_SETUP_GUIDE.md](IPAD_SETUP_GUIDE.md)
- 🧪 Test First: Run `python3 check_compatibility.py`
