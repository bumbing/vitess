#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# the only keyspace currently sharded, and require monitoring is `patio`
KEYSPACE="patio"
# need one of the shards to get `ddl` from
SAMPLE_SHARD="-80"
# suppress alert
SUPPRESS_ALERT=false

if [[ $# -gt 0 && "$1" == "--suppress" ]]; then
    echo "[drift] --suppress option enabled"
    SUPPRESS_ALERT=true
fi

# `pinschema` arguments used to generate the `vschema`
if [[ ${TELETRAAN_ENV_TYPE:-test} == "prod" ]]; then
    # this applies to both `patio` in `prod` & `shadow`
    KEYSPACE_ARGS=(
        -include-cols -cols-authoritative -create-sequences
        -create-primary-vindexes -create-secondary-vindexes
        -default-scatter-cache-capacity 100000
        -table-scatter-cache-capacity "campaigns:200000,product_groups:1000000"
    )
else
    KEYSPACE_ARGS=(
        -create-sequences -include-cols -cols-authoritative
        -create-primary-vindexes -create-secondary-vindexes
        -default-scatter-cache-capacity 100000
    )
fi

if [[ -n "$TELETRAAN_LOOKUP_VINDEX_WHITELIST" ]]; then
    KEYSPACE_ARGS+=("-lookup-vindex-whitelist=$TELETRAAN_LOOKUP_VINDEX_WHITELIST")
fi

VTCTL_CMD="/vt/bin/vtctlclient -server localhost:15991"
PINSCHEMA_CMD="/vt/bin/pinschema"
ALERT_CMD="python -m vitess_utils.alerts"

if ! $VTCTL_CMD ValidateSchemaKeyspace "${KEYSPACE}"; then
    echo "[drift] Bypassing vschema drift check, keyspace: ${KEYSPACE} under change"
    exit 0
fi

echo "[drift] Finding tablets from ${KEYSPACE}/${SAMPLE_SHARD} to pull schemas from..."
KEYSPACE_MASTER=$($VTCTL_CMD ListAllTablets | grep " ${KEYSPACE} ${SAMPLE_SHARD} " | grep " master " | cut -d' ' -f 1)

echo "[drift] Snapshotting current ${KEYSPACE} schema..."
KEYSPACE_SCHEMA_FILE=$(mktemp -t schema-"${KEYSPACE}".sql.XXXX)
KEYSPACE_SCHEMA_CONTENT=$($VTCTL_CMD GetSchema "$KEYSPACE_MASTER" | jq -r '.table_definitions[].schema + ";"')
echo "$KEYSPACE_SCHEMA_CONTENT" > "$KEYSPACE_SCHEMA_FILE"

echo "[drift] Generating ${KEYSPACE} vschema ..."
KEYSPACE_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${KEYSPACE_ARGS[@]}" "$KEYSPACE_SCHEMA_FILE")
KEYSPACE_VSCHEMA_FILE=$(mktemp -t vschema-"${KEYSPACE}".json.XXXX)
echo "$KEYSPACE_VSCHEMA" > "$KEYSPACE_VSCHEMA_FILE"

echo "[drift] Diffing ${KEYSPACE} vschema..."
KEYSPACE_VSCHEMA_OLD=$($VTCTL_CMD GetVSchema ${KEYSPACE})
DIFF=$(diff --strip-trailing-cr -u <(echo "$KEYSPACE_VSCHEMA_OLD") <(echo "$KEYSPACE_VSCHEMA") || test "$?" -eq 1)
KEYSPACE_VSCHEMA_DIFF_FILE=$(mktemp -t vschema-"${KEYSPACE}".diff.XXXX)
if [ "$DIFF" ]; then
    echo "$DIFF" > "$KEYSPACE_VSCHEMA_DIFF_FILE"
    echo "[drift] Reporting diff for ${KEYSPACE} with bot suppressed=${SUPPRESS_ALERT}"
    if [[ "$SUPPRESS_ALERT" != true ]]; then
        $ALERT_CMD --message "vschema drift detected in ${KEYSPACE}" --attach "$KEYSPACE_VSCHEMA_DIFF_FILE"
    fi
fi

echo "[drift] Cleaning up temp files..."
rm -f "$KEYSPACE_SCHEMA_FILE"
rm -f "$KEYSPACE_VSCHEMA_FILE"
rm -f "$KEYSPACE_VSCHEMA_DIFF_FILE"
