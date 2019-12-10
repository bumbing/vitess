#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# this script will validate the completeness of `CREATE TABLE` ddls
# it's meant for integration between DBA's schema change jenkins build and vitess vschema jenkins validation
# NOTE it takes exactly 2 arguments: $1 for `PATIOGENERAL_SCHEMA_FILE` $2 for `PATIO_SCHEMA_FILE`, both should suffix with `*.sql`
PATIOGENERAL_SCHEMA_FILE="${1:?must provide PATIOGENERAL_SCHEMA_FILE}"
PATIO_SCHEMA_FILE="${2:?must provide PATIO_SCHEMA_FILE}"

PATIO_ARGS=(-include-cols 
    -create-sequences 
    -create-primary-vindexes 
    -create-secondary-vindexes
    -default-scatter-cache-capacity 100000 
    -validate-keyspace patio 
    -validate-shards 2
)
PATIOGENERAL_ARGS=(-include-cols)
PINSCHEMA_CMD="/vt/bin/pinschema"

# create a combined `.sql` from both `patiogeneral` and `patio`
PATIO_SCHEMA_VALIDATION_FILE=$(mktemp -t patio-schema-validation.sql.XXXX)
cat "$PATIOGENERAL_SCHEMA_FILE" > "$PATIO_SCHEMA_VALIDATION_FILE"
printf "\n" >> "$PATIO_SCHEMA_VALIDATION_FILE"
cat "$PATIO_SCHEMA_FILE" >> "$PATIO_SCHEMA_VALIDATION_FILE"

# create vschema validation source file which combines both patiogeneral and patio
PATIO_VSCHEMA_VALIDATION_FILE=$(mktemp -t patio-vschema-validation.json.XXXX)
PATIOGENERAL_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${PATIOGENERAL_ARGS[@]}" "$PATIOGENERAL_SCHEMA_FILE")
PATIO_VSCHEMA=$($PINSCHEMA_CMD create-vschema "${PATIO_ARGS[@]}" "$PATIO_SCHEMA_FILE")
cat << EOF > "$PATIO_VSCHEMA_VALIDATION_FILE"
{
  "patio":$PATIO_VSCHEMA,
  "patiogeneral":$PATIOGENERAL_VSCHEMA
}
EOF

$PINSCHEMA_CMD validate-vschema "${PATIO_ARGS[@]}" -validate-vschema-file "$PATIO_VSCHEMA_VALIDATION_FILE" "$PATIO_SCHEMA_VALIDATION_FILE" 2>&1
