#!/usr/bin/env python3
"""Post-write hook: enforce Clean Architecture dependency rules."""
import sys
import json
import re

# Forbidden import rules per layer
RULES = {
    "internal/domain/": {
        "forbidden": [
            "internal/service",
            "internal/repository",
            "internal/handler",
            "internal/config",
            "cloud.google.com",
            "github.com/go-chi",
        ],
        "allowed_external": ["github.com/google/uuid"],
        "label": "domain",
    },
    "internal/service/": {
        "forbidden": [
            "internal/handler",
            "internal/repository/spanner",
            "internal/config",
            "cloud.google.com",
        ],
        "allowed_external": [],
        "label": "service",
    },
    "internal/handler/": {
        "forbidden": [
            "internal/repository/spanner",
            "cloud.google.com/go/spanner",
        ],
        "allowed_external": [],
        "label": "handler",
    },
}


def main():
    data = json.load(sys.stdin)
    file_path = data.get("tool_info", {}).get("file_path", "")

    if not file_path.endswith(".go") or "_test.go" in file_path:
        return

    # Determine which layer this file belongs to
    layer_rules = None
    for path_prefix, rules in RULES.items():
        if path_prefix in file_path:
            layer_rules = rules
            break

    if not layer_rules:
        return  # File is not in a monitored layer

    # Read the file and extract imports
    try:
        with open(file_path, "r") as f:
            content = f.read()
    except FileNotFoundError:
        return

    # Find all import strings
    imports = re.findall(r'"([^"]+)"', content)

    violations = []
    for imp in imports:
        for forbidden in layer_rules["forbidden"]:
            if forbidden in imp:
                violations.append(f"[ARCH] {layer_rules['label']} layer imports forbidden package: {imp}")

    if violations:
        print(f"\n⚠️ Architecture violations in {file_path}:")
        for v in violations:
            print(f"  {v}")
        print(f"\nThe {layer_rules['label']} layer must not depend on these packages.")
        print("See .windsurf/rules/project-architecture.md for dependency rules.\n")


if __name__ == "__main__":
    main()