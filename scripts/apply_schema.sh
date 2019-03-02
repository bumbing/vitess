#!/bin/bash

# Applies a schema change on ads latest and then fixes the vschema.
# Usage:
#     ./scripts/vt_apply_schema.sh test -sql "create table ..." patiogeneral
#     ./scripts/vt_apply_schema.sh test -sql-file create_table.sql patio

set -eu

VTENV="$1"
shift

echo Operating on environment "$VTENV". For ads-latest, run "$0" latest

if [[ "$OSTYPE" == "darwin"* ]]; then
  ./scripts/pvtctl.sh "$VTENV" ApplySchema "$@"
  ./scripts/fix_vschema.sh "$VTENV"
else
  /vt/scripts/pvtctl.sh "$VTENV" ApplySchema "$@"
  /vt/scripts/vt_fix_vschema.sh "$VTENV"
fi
