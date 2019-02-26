# This script runs periodically on the vtctld instance to do sanity checks
# that involve calling vtctl commands.

set -e

echo Logging to /var/log/vtctld/validate_cron.log
exec > /var/log/vtctld/validate_cron.log
exec 2>&1

set -x

VTCTL_CMD="/vt/bin/vtctlclient -server localhost:15991 -action_timeout 10s"

# opentsdb listens on port 18126 at Pinterest.
# It has an HTTP API, but we can also just sent it "put" commands.
REPORT_CMD="nc -w 5 localhost 18126"

# Uncomment when testing to avoid sending real stats:
# REPORT_CMD="cat"

# Check that the topology is consistent with itself and with tablet state.
# Not included yet: -ping-tablets option to verify that tablets are
# reachable.
OUTCOME=failed
if $VTCTL_CMD Validate; then
    OUTCOME=success
fi
echo "put vitess.validate_all $(date +%s) 1 outcome=$OUTCOME" | $REPORT_CMD

for keyspace in $($VTCTL_CMD GetKeyspaces); do

  # Verify that all shhards within a keyspace have the same schema (tables, columns)
  OUTCOME=failed
  if $VTCTL_CMD ValidateSchemaKeyspace $keyspace; then
      OUTCOME=success
  fi
  echo "put vitess.validate_schema $(date +%s) 1 outcome=$OUTCOME keyspace=$keyspace" | $REPORT_CMD

  # Validate that all tablets agree one what users exist and what permissions they have.
  OUTCOME=failed
  if $VTCTL_CMD ValidatePermissionsKeyspace $keyspace; then
      OUTCOME=success
  fi
  echo "put vitess.validate_permissions $(date +%s) 1 outcome=$OUTCOME keyspace=$keyspace" | $REPORT_CMD

  # Validate that all tablets are running the same vttablet version.
  # Technically we already have stats on which tablet versions are running,
  # but there's no harm in having this also.
  OUTCOME=failed
  if $VTCTL_CMD ValidateVersionKeyspace $keyspace; then
      OUTCOME=success
  fi
  echo "put vitess.validate_versions $(date +%s) 1 outcome=$OUTCOME keyspace=$keyspace" | $REPORT_CMD
done
