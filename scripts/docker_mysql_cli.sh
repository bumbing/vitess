#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if [[ "${1:-}" == "" ]]; then
   echo "Usage: docker_mysql_cli.sh [vitess env] [knox role]"
   exit 1
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
  # Silence shellcheck because we want the devapp's $HOME, not laptop's.
  # shellcheck disable=SC2088
  ssh -t devapp '~/code/vitess/scripts/docker_mysql_cli.sh' "$@"
  exit 0
fi

env_name="${1}"
role="${2:-scriptro}"

env_config_file="/var/config/config.services.vitess_environments_config"
env_exists=$(jq -r ".${env_name} | select (.!=null)"  $env_config_file)

if [[ "$env_exists" == "" ]]; then
   echo "Looks like an invalid env name. Valid names tend to look like prod, latest, test, etc."
   echo "Usage: docker_mysql_cli.sh [vitess env] [knox role]"
   exit 1
fi

host=$(jq -r ".${env_name}.gates.master.host" $env_config_file)

# TODO(dweitzman): Set username to role and omit password. Soon passwords will no longer be
# needed as long as TLS is in use.
# username=$role
# password=
username=$(knox get "mysql:rbac:$role:credentials" | cut -d@ -f1)
password=$(knox get "mysql:rbac:$role:credentials" | cut -d\| -f2)
prompt="hostname='$host', port=3306) \d \u($role)> "

cmd="mysql -c -A -h $host -P 3306 --user=$username --password=$password \
   --prompt=\"$prompt\" --ssl-mode=REQUIRED \
   --ssl-ca /var/lib/normandie/fuse/ca/generic \
   --ssl-cert /var/lib/normandie/fuse/cert/generic \
   --ssl-key /var/lib/normandie/fuse/key/generic"
echo "\$ ${cmd/password=$password/password=REDACTED}"
# When testing with 127.0.0.1, you'll need to set "--network host"
docker run -v /var/lib/normandie/fuse/:/var/lib/normandie/fuse/:ro -it mysql:8 sh -c "exec $cmd"
