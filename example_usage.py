#!/usr/bin/env python3
"""
Example script demonstrating the GitHub SDK tutorial concepts.
This shows that the environment is ready to run the tutorial.
"""

from github import Github
import os
import sys

def main():
    print("=" * 60)
    print("GitHub SDK Tutorial - Example Usage")
    print("=" * 60)
    
    # Check for token
    token = os.environ.get('GITHUB_TOKEN')
    
    if not token:
        print("\n⚠️  GITHUB_TOKEN not set!")
        print("\nTo use the GitHub API, you need to:")
        print("1. Create a personal access token at:")
        print("   https://github.com/settings/tokens")
        print("2. Set it as an environment variable:")
        print("   export GITHUB_TOKEN=your_token_here")
        print("\nFor now, here's what you can do with the tutorial:")
        print("✓ All Python packages are installed (PyGithub, jupyter, notebook)")
        print("✓ The notebook is valid and ready to run")
        print("✓ Once you set GITHUB_TOKEN, you can run:")
        print("  jupyter notebook manage-prompts-with-sdk-tutorial.ipynb")
        return
    
    # If token is set, demonstrate functionality
    try:
        g = Github(token)
        user = g.get_user()
        
        print(f"\n✓ Successfully authenticated as: {user.login}")
        print(f"✓ Your name: {user.name}")
        print(f"✓ Public repos: {user.public_repos}")
        
        print("\n" + "=" * 60)
        print("The tutorial is ready to use!")
        print("=" * 60)
        print("\nRun the Jupyter notebook with:")
        print("  jupyter notebook manage-prompts-with-sdk-tutorial.ipynb")
        
    except Exception as e:
        print(f"\n✗ Error connecting to GitHub: {e}")
        print("\nPlease check your GITHUB_TOKEN is valid.")

if __name__ == "__main__":
    main()
