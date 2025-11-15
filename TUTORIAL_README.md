# GitHub SDK Tutorial

This directory contains a Jupyter notebook tutorial for working with the GitHub API using Python.

## ✓ Environment Ready

All required packages are installed and the tutorial is ready to run in this environment!

## File: manage-prompts-with-sdk-tutorial.ipynb

A comprehensive tutorial that demonstrates how to programmatically interact with GitHub repositories, issues, pull requests, and more using the PyGithub library.

### Prerequisites

- Python 3.7 or higher
- A GitHub personal access token with appropriate permissions
- Jupyter Notebook or JupyterLab

### Installation

**Already installed in this environment!** ✓

If you need to install in a different environment:

```bash
pip install -r tutorial-requirements.txt
```

Or install them individually:

```bash
pip install PyGithub requests jupyter notebook
```

### Quick Test

Run the example script to verify everything is working:

```bash
python3 example_usage.py
```

### Running the Tutorial

1. Set your GitHub token as an environment variable:
   ```bash
   export GITHUB_TOKEN=your_personal_access_token
   ```

2. Start Jupyter Notebook:
   ```bash
   jupyter notebook
   ```

3. Open `manage-prompts-with-sdk-tutorial.ipynb` in the Jupyter interface

4. Follow the tutorial step by step, running each cell

### What You'll Learn

- Authenticating with the GitHub API
- Accessing repository information
- Managing issues programmatically
- Working with pull requests
- Adding and reading comments
- Searching GitHub
- Managing labels

### Safety Notes

The tutorial includes commented-out code for operations that modify data (creating issues, adding comments, etc.). Uncomment these sections only if you:
- Have write access to the repository
- Understand the implications of the operations
- Are working in a test environment or are intentionally creating/modifying data

### Resources

- [PyGithub Documentation](https://pygithub.readthedocs.io/)
- [GitHub REST API Documentation](https://docs.github.com/en/rest)
- [GitHub CLI Documentation](https://cli.github.com/manual/)

## Contributing

If you find issues or want to improve the tutorial, please submit a pull request or open an issue in the repository.
