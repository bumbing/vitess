#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# check with each vitess `master`, see if topology is the same as mysql's
MASTERS="$(/vt/bin/vtctlclient -server localhost:15991 ListAllTablets | grep master | cut -d ' ' -f 5 | cut -d '.' -f 1)"
for TABLET in $MASTERS; do
    python -m vitess_utils.vitess --action validateshardtabletmaster --host "$TABLET"
done
