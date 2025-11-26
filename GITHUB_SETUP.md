# GitHub Repository Setup Guide

This guide will help you push your sticker project to GitHub.

## Step 1: Install Git (if not already installed)

### Option A: Install Git for Windows
1. Download Git from: https://git-scm.com/download/win
2. Run the installer and follow the prompts
3. Restart your terminal/PowerShell after installation

### Option B: Use GitHub Desktop (Easier GUI option)
1. Download GitHub Desktop from: https://desktop.github.com/
2. Install and sign in with your GitHub account
3. Use the GUI to create and push repositories

## Step 2: Initialize Git Repository

Open PowerShell in your project directory and run:

```powershell
# Initialize git repository
git init

# Configure your git identity (if not already set)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Mini Sticker Engine project"
```

## Step 3: Create GitHub Repository

1. Go to https://github.com/new
2. Repository name: `sticker-project` (or any name you prefer)
3. Description: "Mini Sticker Engine - A simplified sticker-style loyalty campaign system"
4. Choose **Public** or **Private**
5. **DO NOT** initialize with README, .gitignore, or license (we already have these)
6. Click "Create repository"

## Step 4: Connect and Push to GitHub

After creating the repository, GitHub will show you commands. Use these:

```powershell
# Add the remote repository (replace YOUR_USERNAME with your GitHub username)
git remote add origin https://github.com/YOUR_USERNAME/sticker-project.git

# Rename branch to main (if needed)
git branch -M main

# Push to GitHub
git push -u origin main
```

If you're using SSH instead of HTTPS:
```powershell
git remote add origin git@github.com:YOUR_USERNAME/sticker-project.git
git branch -M main
git push -u origin main
```

## Step 5: Verify

1. Go to your GitHub repository page
2. Verify all files are uploaded
3. Check that `TECH_NOTES.md` is visible
4. Copy the repository URL to share

## Troubleshooting

### Authentication Issues
If you get authentication errors:
- For HTTPS: Use a Personal Access Token instead of password
  - Go to GitHub Settings > Developer settings > Personal access tokens
  - Generate a new token with `repo` permissions
  - Use the token as your password when pushing
- For SSH: Set up SSH keys
  - Follow: https://docs.github.com/en/authentication/connecting-to-github-with-ssh

### Git Not Found
- Make sure Git is installed and added to PATH
- Restart your terminal after installation
- Try using the full path: `C:\Program Files\Git\bin\git.exe`

## Quick Reference

Your repository URL will be:
```
https://github.com/YOUR_USERNAME/sticker-project
```

Make sure to include this URL when submitting your solution!

