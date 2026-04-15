#!/usr/bin/env python3
"""Post-write hook: auto-format Go files after Cascade edits them."""
import sys
import json
import subprocess
import shutil


def main():
    data = json.load(sys.stdin)
    file_path = data.get("tool_info", {}).get("file_path", "")

    if not file_path.endswith(".go"):
        return

    # Run gofmt
    if shutil.which("gofmt"):
        subprocess.run(["gofmt", "-w", file_path], capture_output=True)

    # Run goimports if available
    if shutil.which("goimports"):
        subprocess.run(["goimports", "-w", file_path], capture_output=True)


if __name__ == "__main__":
    main()