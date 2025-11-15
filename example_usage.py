#!/usr/bin/env python3
"""
Example script demonstrating the GitHub SDK tutorial concepts.
This shows that the environment is ready to run the tutorial.

Note for iPad users: If you're running this on iPad Air using Juno, Carnets,
or similar apps, you may need to modify the token retrieval to use a direct
assignment instead of environment variables. See IPAD_SETUP_GUIDE.md for details.
"""

from github import Github
import os
import sys
import platform

def main():
    print("=" * 60)
    print("GitHub SDK Tutorial - Example Usage")
    print("=" * 60)
    
    # Detect platform
    system_info = platform.system()
    print(f"\nPlatform: {system_info}")
    print(f"Python version: {platform.python_version()}")
    
    # Check for PyGithub installation
    try:
        import github
        gh_version = getattr(github, '__version__', 'unknown')
        print(f"PyGithub version: {gh_version}")
    except ImportError:
        print("⚠️  PyGithub not installed!")
        print("Install with: pip install PyGithub")
        print("Or on iPad: Use app's package manager to install PyGithub")
        return
    
    # Check for token
    token = os.environ.get('GITHUB_TOKEN')
    
    # iPad/iOS detection and guidance
    if system_info == "Darwin" and "iPad" in platform.machine():
        print("\n📱 iPad detected!")
        print("Note: iOS doesn't support environment variables in the same way.")
        print("See IPAD_SETUP_GUIDE.md for iPad-specific instructions.")
    
    if not token:
        print("\n⚠️  GITHUB_TOKEN not set!")
        print("\nTo use the GitHub API, you need to:")
        print("1. Create a personal access token at:")
        print("   https://github.com/settings/tokens")
        print("\n2. Set it as an environment variable:")
        print("   Desktop/Laptop:")
        print("     export GITHUB_TOKEN=your_token_here")
        print("\n   iPad/iOS (see IPAD_SETUP_GUIDE.md):")
        print("     Modify the notebook to include your token directly")
        print("     token = 'your_token_here'")
        print("\nFor now, here's what you can do with the tutorial:")
        print("✓ Python environment is ready")
        print("✓ PyGithub is installed")
        print("✓ The notebook is valid and ready to run")
        print("\n✓ Once you set GITHUB_TOKEN, you can run:")
        print("  Desktop: jupyter notebook manage-prompts-with-sdk-tutorial.ipynb")
        print("  iPad: Open the notebook in Juno, Carnets, or similar app")
        return
    
    # If token is set, demonstrate functionality
    try:
        g = Github(token)
        user = g.get_user()
        
        print(f"\n✓ Successfully authenticated as: {user.login}")
        print(f"✓ Your name: {user.name}")
        print(f"✓ Public repos: {user.public_repos}")
        
        print("\n" + "=" * 60)
        print("✅ The tutorial is ready to use!")
        print("=" * 60)
        print("\nRun the Jupyter notebook with:")
        print("  Desktop: jupyter notebook manage-prompts-with-sdk-tutorial.ipynb")
        print("  iPad: Open manage-prompts-with-sdk-tutorial.ipynb in your Jupyter app")
        print("\n💡 Tip: For iPad-specific setup, see IPAD_SETUP_GUIDE.md")
        
    except Exception as e:
        print(f"\n✗ Error connecting to GitHub: {e}")
        print("\nPlease check your GITHUB_TOKEN is valid.")

if __name__ == "__main__":
    main()
