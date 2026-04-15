#!/usr/bin/env python3
"""Pre-run-command hook: block dangerous terminal commands."""
import sys
import json

BLOCKED_PATTERNS = [
    "rm -rf /",
    "rm -rf ~",
    "rm -rf .",
    "drop database",
    "DROP DATABASE",
    "git push --force",
    "git reset --hard",
    "chmod -R 777",
    ":(){ :|:& };:",
]

WARN_PATTERNS = [
    "rm -rf",
    "git clean -fd",
    "docker system prune",
]


def main():
    data = json.load(sys.stdin)
    command = data.get("tool_info", {}).get("command_line", "")

    # Check for blocked patterns
    for pattern in BLOCKED_PATTERNS:
        if pattern in command:
            print(
                f"\n\u274c BLOCKED: Cascade attempted to run a dangerous command:"
                f"\n  {command}"
                f"\n\nPattern matched: '{pattern}'"
                f"\nThis command has been prevented from executing.",
                file=sys.stderr,
            )
            sys.exit(2)  # Exit code 2 = block the action

    # Check for warning patterns (don't block, just inform)
    for pattern in WARN_PATTERNS:
        if pattern in command:
            print(
                f"\n⚠️ WARNING: Potentially destructive command:"
                f"\n  {command}"
                f"\nProceeding, but please verify this is intended."
            )
            return


if __name__ == "__main__":
    main()