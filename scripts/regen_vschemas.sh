# TODO(dweitzman): Move this to the vitess-utils repo
# Long term, we'll want to integrate this into the schema change process such that we create vindexes
# prior to the db cols and update the authoritative column list in vschema after updating the db.
#
# This is a convenient script for regenerating the patio and patiogeneral vschemas, which can be applied using vtctl ApplyVSchema
# Assumptions:
# - You're in the vitess repo root directory
# - You've run "go mod init"
# - You have optimus checked out at ~/code/optimus
# - You've recently run ./pepsi/utils/dump_patio_tables.sh in the optimus directory to get the latest schema
#
# Then to apply the changes you could use commands like these:
# $ vtctl <env-specific args> ApplyVSchema -vschema_file patio.json patio
# $ vtctl <env-specific args> ApplyVSchema -vschema_file patiogeneral.json patiogeneral

set -ex

# PATIO_DDLS=~/code/optimus/pepsi/server/src/test/resources/patio_db_dump/patio.sql
# GENERAL_DDLS=~/code/optimus/pepsi/server/src/test/resources/patio_db_dump/patio_general.sql

PATIO_DDLS=$(find ~/code/mysql_change_management/sharded/coladb/schemas/ -name "*.sql")
GENERAL_DDLS=$(find ~/code/mysql_change_management/sharded/patiogeneraldb/schemas/ -name "*.sql")

OUTPUT_DIR=genschemas/
INCLUDE_COLS_ARGS="-include-cols -cols-authoritative"
VINDEX_ARGS="-create-primary-vindexes \
-create-secondary-vindexes \
-default-scatter-cache-capacity 100000 \
-table-scatter-cache-capacity campaigns:200000"
PATIO_SEQUENCE_ARGS="-create-sequences"

mkdir -p $OUTPUT_DIR
go run vitess.io/vitess/go/cmd/pinschema create-vschema $INCLUDE_COLS_ARGS $VINDEX_ARGS $PATIO_SEQUENCE_ARGS $PATIO_DDLS > $OUTPUT_DIR/patio.json
go run vitess.io/vitess/go/cmd/pinschema create-vschema $INCLUDE_COLS_ARGS $GENERAL_DDLS > $OUTPUT_DIR/patiogeneral.json
go run vitess.io/vitess/go/cmd/pinschema create-seq $PATIO_DDLS > $OUTPUT_DIR/create_seq.sql
go run vitess.io/vitess/go/cmd/pinschema remove-autoinc $PATIO_DDLS > $OUTPUT_DIR/remove_autoinc.sql