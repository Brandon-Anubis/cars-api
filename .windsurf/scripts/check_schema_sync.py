#!/usr/bin/env python3
"""Post-write hook: detect drift between SQL schema and Go domain model."""
import sys
import json
import re
import os

SQL_FILE = "migrations/001_create_cars.sql"
DOMAIN_FILE = "internal/domain/car.go"

# Map Spanner SQL types to expected Go types
SPANNER_TYPE_MAP = {
    "STRING": ["string", "spanner.NullString"],
    "INT64": ["int64", "int", "spanner.NullInt64"],
    "FLOAT64": ["float64", "spanner.NullFloat64"],
    "TIMESTAMP": ["time.Time"],
    "BOOL": ["bool", "spanner.NullBool"],
}


def extract_sql_columns(sql_content):
    """Extract column names from CREATE TABLE statement."""
    columns = []
    for line in sql_content.split("\n"):
        line = line.strip().rstrip(",")
        match = re.match(r"^(\w+)\s+(STRING|INT64|FLOAT64|TIMESTAMP|BOOL)", line)
        if match:
            columns.append(match.group(1))
    return columns


def extract_struct_fields(go_content):
    """Extract field names from Go struct with spanner tags."""
    fields = []
    tag_pattern = re.compile(r'spanner:"(\w+)"')
    for line in go_content.split("\n"):
        match = tag_pattern.search(line)
        if match:
            fields.append(match.group(1))
    return fields


def main():
    data = json.load(sys.stdin)
    file_path = data.get("tool_info", {}).get("file_path", "")

    # Only run for schema or domain model changes
    if SQL_FILE not in file_path and DOMAIN_FILE not in file_path:
        return

    # Check both files exist
    if not os.path.exists(SQL_FILE) or not os.path.exists(DOMAIN_FILE):
        return

    with open(SQL_FILE, "r") as f:
        sql_columns = extract_sql_columns(f.read())

    with open(DOMAIN_FILE, "r") as f:
        struct_fields = extract_struct_fields(f.read())

    if not sql_columns or not struct_fields:
        return

    # Compare
    sql_set = set(sql_columns)
    struct_set = set(struct_fields)

    in_sql_not_struct = sql_set - struct_set
    in_struct_not_sql = struct_set - sql_set

    if in_sql_not_struct or in_struct_not_sql:
        print(f"\n⚠️ Schema/Model drift detected:")
        if in_sql_not_struct:
            print(f"  Columns in SQL but missing spanner tag in Go struct: {in_sql_not_struct}")
        if in_struct_not_sql:
            print(f"  Spanner tags in Go struct but missing from SQL: {in_struct_not_sql}")
        print(f"\nPlease sync {SQL_FILE} and {DOMAIN_FILE}.\n")


if __name__ == "__main__":
    main()