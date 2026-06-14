#!/usr/bin/env python3
import os
import sys
import json
import re
import subprocess
from pathlib import Path

def find_gvm_config():
    """Search for .gvm configuration file upwards from the current directory,
    then in the user's home directory.
    """
    try:
        curr = Path.cwd()
        # Search upwards from the current working directory to root
        for parent in [curr] + list(curr.parents):
            gvm = parent / '.gvm'
            if gvm.is_file():
                return gvm
    except Exception:
        pass
    
    # Fallback to user home directory
    try:
        home_gvm = Path.home() / '.gvm'
        if home_gvm.is_file():
            return home_gvm
    except Exception:
        pass
        
    return None

def parse_version_from_filename(filename):
    """Parses a Godot filename to extract version details for sorting.
    Handles formats like: Godot_v4.6.3-stable_win64.exe, Godot_v4.10.0-rc2_win64_console.exe
    Returns a tuple: (major, minor, patch, status_rank, status_num)
    """
    # Look for Godot_v[major].[minor].[patch]-[status] or Godot_v[major].[minor]-[status]
    match = re.search(r'[Gg]odot_v([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(?:-([a-zA-Z0-9\.]+))?', filename)
    if match:
        major = int(match.group(1))
        minor = int(match.group(2))
        patch = int(match.group(3)) if match.group(3) else 0
        status = match.group(4) if match.group(4) else 'stable'
        
        # Sort 'stable' highest, down to dev releases
        status_order = {'stable': 5, 'rc': 4, 'beta': 3, 'alpha': 2, 'dev': 1}
        # Extract letters for status ranking (e.g. "rc" from "rc2")
        status_type = ''.join(c for c in status if c.isalpha()).lower()
        # Extract trailing digits (e.g. 2 from "rc2")
        status_num_match = re.search(r'\d+', status)
        status_num = int(status_num_match.group(0)) if status_num_match else 0
        
        status_rank = status_order.get(status_type, 0)
        return (major, minor, patch, status_rank, status_num)
    
    return (0, 0, 0, 0, 0)

def main():
    godot_dir = Path(__file__).parent.resolve() / 'godot-versions'
    if not godot_dir.is_dir():
        print(f"Error: Godot installation directory '{godot_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)
        
    config_path = find_gvm_config()
    version = None
    
    if config_path:
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                # Support both "version" and "godot" keys
                version = data.get('version') or data.get('godot')
        except Exception as e:
            print(f"Warning: Failed to parse configuration file '{config_path}': {e}", file=sys.stderr)
            
    # List all executable files in the Godot directory
    try:
        binaries = [Path(godot_dir) / f for f in os.listdir(godot_dir) if f.lower().endswith('.exe')]
    except Exception as e:
        print(f"Error: Failed to list directory '{godot_dir}': {e}", file=sys.stderr)
        sys.exit(1)
        
    if not binaries:
        print(f"Error: No executables (.exe) found in '{godot_dir}'.", file=sys.stderr)
        sys.exit(1)
        
    selected_exe = None
    if version:
        # Normalize target version string (strip leading 'v' if present)
        norm_version = version.lstrip('v').lower()
        
        # Filter binaries matching the normalized version in their filename
        matches = [b for b in binaries if norm_version in b.name.lower()]
        if not matches:
            print(f"Error: No Godot binary matching version '{version}' found in '{godot_dir}'.", file=sys.stderr)
            print("Available binaries in directory:", file=sys.stderr)
            for b in sorted(binaries, key=lambda x: x.name):
                print(f"  - {b.name}", file=sys.stderr)
            sys.exit(1)
            
        # Prioritize console binary (usually prints stdout/stderr to shell)
        console_matches = [m for m in matches if '_console.exe' in m.name.lower()]
        if console_matches:
            # Sort matching console binaries just in case there are multiple, get latest
            selected_exe = sorted(console_matches, key=lambda x: parse_version_from_filename(x.name))[-1]
        else:
            selected_exe = sorted(matches, key=lambda x: parse_version_from_filename(x.name))[-1]
    else:
        # No version specified in .gvm; fallback to latest version
        # Separate into console and non-console binaries
        console_binaries = [b for b in binaries if '_console.exe' in b.name.lower()]
        if console_binaries:
            selected_exe = sorted(console_binaries, key=lambda x: parse_version_from_filename(x.name))[-1]
        else:
            selected_exe = sorted(binaries, key=lambda x: parse_version_from_filename(x.name))[-1]
            
    if not selected_exe or not selected_exe.is_file():
        print("Error: Could not find or access a suitable Godot executable.", file=sys.stderr)
        sys.exit(1)
        
    # Launch Godot with forwarded arguments
    cmd = [str(selected_exe)] + sys.argv[1:]
    try:
        # Run subprocess and exit with its return code
        sys.exit(subprocess.call(cmd))
    except KeyboardInterrupt:
        # Handle ctrl+c gracefully
        sys.exit(130)
    except Exception as e:
        print(f"Error: Failed to execute Godot binary: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
