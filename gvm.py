#!/usr/bin/env python3
import os
import sys
import json
import re
import urllib.request
import zipfile
import shutil
import time
from pathlib import Path

# Add script directory to sys.path to allow importing godot.py
script_dir = Path(__file__).parent.resolve()
if str(script_dir) not in sys.path:
    sys.path.insert(0, str(script_dir))
import godot

# Global godot installation directory (next to this script)
godot_dir = script_dir / 'godot-versions'


def get_installed_versions(godot_dir):
    """Scan the godot-versions directory for installed Godot executables,
    extract versions, and group them.
    """
    try:
        binaries = [f for f in os.listdir(godot_dir) if f.lower().endswith('.exe')]
    except Exception as e:
        print(f"Error: Failed to read directory '{godot_dir}': {e}", file=sys.stderr)
        return []
        
    versions = {}
    for b in binaries:
        v_info = godot.parse_version_from_filename(b)
        if v_info != (0, 0, 0, 0, 0):
            # Extract version string from filename
            match = re.search(r'[Gg]odot_v([0-9]+\.[0-9]+(?:\.[0-9]+)?(?:-[a-zA-Z0-9\.]+)?)(?:_|\.)', b)
            if match:
                v_str = match.group(1)
                if v_str not in versions:
                    versions[v_str] = []
                versions[v_str].append(b)
                
    return versions

def cmd_list():
    if not godot_dir.is_dir():
        print(f"Error: Godot installation directory '{godot_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)
        
    installed = get_installed_versions(godot_dir)
    if not installed:
        print(f"No Godot versions installed in {godot_dir}.")
        return
        
    # Get active version from .gvm configuration
    config_path = godot.find_gvm_config()
    active_version = None
    if config_path:
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                active_version = data.get('version') or data.get('godot')
                if active_version:
                    active_version = active_version.lstrip('v').lower()
        except Exception:
            pass
            
    print(f"Installed versions in {godot_dir}:")
    # Sort versions descending
    sorted_versions = sorted(installed.keys(), key=lambda x: godot.parse_version_from_filename(f"Godot_v{x}_win64.exe"), reverse=True)
    for v in sorted_versions:
        is_active = False
        if active_version and (active_version == v.lower() or active_version == v.lower().replace('-stable', '')):
            is_active = True
            
        active_label = " (active)" if is_active else ""
        prefix = "  * " if is_active else "    "
        
        # List if console or GUI or both are present
        files = installed[v]
        has_console = any('_console.exe' in f.lower() for f in files)
        has_gui = any('_console.exe' not in f.lower() for f in files)
        types = []
        if has_gui: types.append("GUI")
        if has_console: types.append("Console")
        types_str = f" [{', '.join(types)}]" if types else ""
        
        print(f"{prefix}{v}{types_str}{active_label}")

def get_repo_for_tag(tag):
    tag_lower = tag.lower()
    if any(pre in tag_lower for pre in ['rc', 'beta', 'dev', 'alpha']):
        return "godotengine/godot-builds"
    return "godotengine/godot"

def cmd_releases(show_all=False):
    repo = "godotengine/godot-builds" if show_all else "godotengine/godot"
    url = f"https://api.github.com/repos/{repo}/releases?per_page=100"
    label = "all" if show_all else "stable"
    print(f"Fetching {label} releases from GitHub ({repo})...")
    try:
        req = urllib.request.Request(
            url, 
            headers={
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) GVM/1.0',
                'Accept': 'application/vnd.github.v3+json'
            }
        )
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode('utf-8'))
            
        releases = []
        for release in data:
            tag = release.get('tag_name', '')
            clean_tag = tag.lstrip('v')
            # Filter for stable releases unless show_all is specified
            if show_all or 'stable' in tag.lower():
                releases.append(clean_tag)
                
        print(f"\nAvailable {label} releases:")
        # Display releases
        for r in releases:
            print(f"  - {r}")
            
    except Exception as e:
        print(f"Error fetching releases: {e}", file=sys.stderr)
        sys.exit(1)

def download_with_progress(url, dest_path):
    req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        with urllib.request.urlopen(req) as response:
            total_size = int(response.info().get('Content-Length', 0))
            block_size = 1024 * 64  # 64 KB blocks
            downloaded = 0
            
            print(f"Downloading to temp file...")
            start_time = time.time()
            with open(dest_path, 'wb') as f:
                while True:
                    buffer = response.read(block_size)
                    if not buffer:
                        break
                    f.write(buffer)
                    downloaded += len(buffer)
                    
                    if total_size > 0:
                        percent = downloaded / total_size
                        bar_length = 30
                        filled_length = int(round(bar_length * percent))
                        bar = '█' * filled_length + '-' * (bar_length - filled_length)
                        elapsed = time.time() - start_time
                        speed = (downloaded / (1024 * 1024)) / elapsed if elapsed > 0 else 0
                        sys.stdout.write(f"\r|{bar}| {percent:.1%} ({downloaded / (1024*1024):.1f}/{total_size / (1024*1024):.1f} MB) - {speed:.1f} MB/s")
                        sys.stdout.flush()
                    else:
                        sys.stdout.write(f"\rDownloaded {downloaded / (1024*1024):.1f} MB")
                        sys.stdout.flush()
            print("\nDownload complete.")
            return True
    except Exception as e:
        print(f"\nError downloading file: {e}", file=sys.stderr)
        if dest_path.exists():
            os.remove(dest_path)
        return False

def extract_zip_and_clean(zip_path, extract_dir):
    temp_extract = Path(extract_dir) / "_temp_extract"
    if temp_extract.exists():
        shutil.rmtree(temp_extract)
    temp_extract.mkdir(parents=True, exist_ok=True)
    
    print(f"Extracting package...")
    try:
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            zip_ref.extractall(temp_extract)
            
        # Check if there's a single directory inside temp_extract
        items = os.listdir(temp_extract)
        if len(items) == 1 and os.path.isdir(temp_extract / items[0]):
            source_dir = temp_extract / items[0]
        else:
            source_dir = temp_extract
            
        # Move all items from source_dir to extract_dir
        print(f"Installing files to '{extract_dir}'...")
        for item in os.listdir(source_dir):
            src = source_dir / item
            dest = Path(extract_dir) / item
            if dest.exists():
                if dest.is_dir():
                    shutil.rmtree(dest)
                else:
                    os.remove(dest)
            shutil.move(str(src), str(dest))
            
        print("Installation complete.")
        return True
    except Exception as e:
        print(f"Error during extraction: {e}", file=sys.stderr)
        return False
    finally:
        if temp_extract.exists():
            shutil.rmtree(temp_extract)

def cmd_install(version):
    if not godot_dir.is_dir():
        try:
            os.makedirs(godot_dir, exist_ok=True)
        except Exception as e:
            print(f"Error: Cannot create directory '{godot_dir}': {e}", file=sys.stderr)
            sys.exit(1)
            
    # Normalize version input (e.g., "4.6.3" -> "4.6.3-stable")
    tag = version.lstrip('v')
    if re.match(r'^\d+\.\d+(?:\.\d+)*$', tag):
        tag = f"{tag}-stable"
        
    # Check if already installed
    installed = get_installed_versions(godot_dir)
    # Match both standard and tag format
    if tag in installed or tag.replace('-stable', '') in installed:
        print(f"Version '{tag}' is already installed.")
        return
        
    print(f"Searching for release '{tag}' on GitHub...")
    # Fetch release info to get exact asset URLs from appropriate repository
    repo = get_repo_for_tag(tag)
    url = f"https://api.github.com/repos/{repo}/releases/tags/{tag}"
    
    try:
        req = urllib.request.Request(
            url, 
            headers={
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) GVM/1.0',
                'Accept': 'application/vnd.github.v3+json'
            }
        )
        with urllib.request.urlopen(req) as response:
            release_data = json.loads(response.read().decode('utf-8'))
    except Exception as e:
        print(f"Error: Release '{tag}' not found or GitHub API error: {e}", file=sys.stderr)
        print("Please verify the version number. Use 'gvm releases -a' to list all available versions.", file=sys.stderr)
        sys.exit(1)
        
    assets = release_data.get('assets', [])
    download_url = None
    asset_name = None
    
    # Search for Windows 64-bit non-mono zip (ends with win64.exe.zip or win64.zip)
    for asset in assets:
        name = asset.get('name', '')
        # We look for a 64-bit windows executable zip: e.g., Godot_v4.6.3-stable_win64.exe.zip
        # Make sure we don't accidentally match mono or debug symbols unless that's what's wanted
        if '_win64.exe.zip' in name.lower() or ('_win64.zip' in name.lower() and 'mono' not in name.lower() and 'debug' not in name.lower()):
            download_url = asset.get('browser_download_url')
            asset_name = name
            break
            
    if not download_url:
        print(f"Error: Could not find a suitable Windows 64-bit build (.win64.exe.zip) in release '{tag}'.", file=sys.stderr)
        print("Available assets for this release:", file=sys.stderr)
        for asset in assets:
            print(f"  - {asset.get('name')}", file=sys.stderr)
        sys.exit(1)
        
    # Download zip to temp directory
    temp_dir = Path(os.environ.get('TEMP', '.'))
    temp_zip_path = temp_dir / f"gvm_download_{tag}.zip"
    
    print(f"Found asset: {asset_name}")
    if not download_with_progress(download_url, temp_zip_path):
        sys.exit(1)
        
    # Extract and install
    success = extract_zip_and_clean(temp_zip_path, godot_dir)
    
    # Clean up zip
    if temp_zip_path.exists():
        try:
            os.remove(temp_zip_path)
        except Exception:
            pass
            
    if not success:
        sys.exit(1)
        
    print(f"Successfully installed Godot {tag}!")

def cmd_use(version):
    if not godot_dir.is_dir():
        print(f"Error: Godot installation directory '{godot_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)
        
    # Normalize version input
    tag = version.lstrip('v')
    if re.match(r'^\d+\.\d+(?:\.\d+)*$', tag):
        tag = f"{tag}-stable"
        
    installed = get_installed_versions(godot_dir)
    
    # Check if the version is installed
    matched_version = None
    for v in installed.keys():
        if v.lower() == tag.lower() or v.lower().replace('-stable', '') == tag.lower().replace('-stable', ''):
            matched_version = v
            break
            
    if not matched_version:
        print(f"Error: Version '{version}' is not installed locally in {godot_dir}.", file=sys.stderr)
        print("Please install it first using: gvm install <version>", file=sys.stderr)
        sys.exit(1)
        
    # Create/update .gvm file in the current directory
    config_path = Path.cwd() / '.gvm'
    config_data = {
        "version": matched_version
    }
    
    try:
        with open(config_path, 'w', encoding='utf-8') as f:
            json.dump(config_data, f, indent=2)
        print(f"Success: Active Godot version set to '{matched_version}' in this directory ({config_path}).")
    except Exception as e:
        print(f"Error: Failed to write configuration file '{config_path}': {e}", file=sys.stderr)
        sys.exit(1)

def cmd_uninstall(version, force=False):
    if not godot_dir.is_dir():
        print(f"Error: Godot installation directory '{godot_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)

    # Normalize version input
    tag = version.lstrip('v')
    if re.match(r'^\d+\.\d+(?:\.\d+)*$', tag):
        tag = f"{tag}-stable"

    installed = get_installed_versions(godot_dir)

    # Find matching version
    matched_version = None
    for v in installed.keys():
        if v.lower() == tag.lower() or v.lower().replace('-stable', '') == tag.lower().replace('-stable', ''):
            matched_version = v
            break

    if not matched_version:
        print(f"Error: Version '{version}' is not installed.", file=sys.stderr)
        sys.exit(1)

    files_to_delete = [godot_dir / f for f in installed[matched_version]]

    # Warn if uninstalling the currently active version
    config_path = godot.find_gvm_config()
    active_version = None
    if config_path:
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                active_version = (data.get('version') or data.get('godot', '')).lstrip('v').lower()
        except Exception:
            pass
    if active_version and active_version.replace('-stable', '') == matched_version.lower().replace('-stable', ''):
        print(f"Warning: '{matched_version}' is currently the active version in '{config_path}'.")

    # List files to be deleted
    print(f"The following files will be removed:")
    for f in files_to_delete:
        print(f"  - {f.name}")

    # Confirm unless -y / --yes
    if not force:
        try:
            answer = input(f"\nUninstall Godot {matched_version}? [y/N] ").strip().lower()
        except (EOFError, KeyboardInterrupt):
            print("\nAborted.")
            sys.exit(0)
        if answer not in ('y', 'yes'):
            print("Aborted.")
            sys.exit(0)

    # Delete files
    errors = []
    for f in files_to_delete:
        try:
            f.unlink()
        except Exception as e:
            errors.append(f"{f.name}: {e}")

    if errors:
        print("Some files could not be deleted:", file=sys.stderr)
        for err in errors:
            print(f"  - {err}", file=sys.stderr)
        sys.exit(1)

    print(f"Successfully uninstalled Godot {matched_version}!")

def print_help():
    help_text = f"""Godot Version Manager (GVM) - CLI

Usage:
  gvm list | ls                     Show locally installed Godot versions in {godot_dir}
  gvm releases [-a]                 List available releases from GitHub (use -a/--all for pre-releases)
  gvm install <version>             Download and install a specific Godot version (e.g. 4.6.3)
  gvm use <version>                 Set the active Godot version for the current directory
  gvm uninstall <version> [-y]      Uninstall a locally installed version (-y to skip confirmation)

Example:
  gvm install 4.6.3
  gvm use 4.6.3
  gvm uninstall 4.6.3
"""
    print(help_text)

def main():
    if len(sys.argv) < 2:
        print_help()
        sys.exit(0)
        
    command = sys.argv[1].lower()
    
    if command in ('--help', '-h', 'help'):
        print_help()
    elif command in ('list', 'ls'):
        cmd_list()
    elif command in ('releases', 'release'):
        show_all = False
        if len(sys.argv) > 2 and sys.argv[2].lower() in ('-a', '--all'):
            show_all = True
        cmd_releases(show_all)
    elif command == 'install':
        if len(sys.argv) < 3:
            print("Error: Please specify the version to install. E.g.: gvm install 4.6.3", file=sys.stderr)
            sys.exit(1)
        cmd_install(sys.argv[2])
    elif command == 'use':
        if len(sys.argv) < 3:
            print("Error: Please specify the version to use. E.g.: gvm use 4.6.3", file=sys.stderr)
            sys.exit(1)
        cmd_use(sys.argv[2])
    elif command in ('uninstall', 'remove'):
        if len(sys.argv) < 3:
            print("Error: Please specify the version to uninstall. E.g.: gvm uninstall 4.6.3", file=sys.stderr)
            sys.exit(1)
        force = len(sys.argv) > 3 and sys.argv[3].lower() in ('-y', '--yes')
        cmd_uninstall(sys.argv[2], force)
    else:
        print(f"Error: Unknown command '{command}'.", file=sys.stderr)
        print_help()
        sys.exit(1)

if __name__ == '__main__':
    main()
