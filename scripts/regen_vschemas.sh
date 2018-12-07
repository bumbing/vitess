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

go run vitess.io/vitess/go/cmd/pinschema -create-primary-vindexes -create-secondary-vindexes -default-scatter-cache-capacity 100000 -create-sequences -table-scatter-cache-capacity campaigns:200000 ~/code/optimus/pepsi/server/src/test/resources/patio_db_dump/patio.sql > patio.json
go run vitess.io/vitess/go/cmd/pinschema -sequence-table-ddls ~/code/optimus/pepsi/server/src/test/resources/patio_db_dump/patio.sql > patio_ddl.sql
go run vitess.io/vitess/go/cmd/pinschema ~/code/optimus/pepsi/server/src/test/resources/patio_db_dump/patio_general.sql > patiogeneral.json

