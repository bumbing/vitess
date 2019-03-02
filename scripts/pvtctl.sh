#!/bin/bash
set -eu

ENV_NAME="$1"
shift

if [[ "$ENV_NAME" == "" ]]; then
   echo "Usage: pvtcl.sh [vitess env] [args to vtctl...]"
   exit 1
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
    # Proxy through vtctld from a mac
    if [[ "$ENV_NAME" == "prod" ]]; then
        HOST_TYPE="vtctld-prod"
    elif [[ "$ENV_NAME" == "shadow" ]]; then
        HOST_TYPE="vtctld-shadow"
    elif [[ "$ENV_NAME" == "latest" ]]; then
        HOST_TYPE="vtctld-latest"
    elif [[ "$ENV_NAME" == "test" ]]; then
        HOST_TYPE="vtctld-test"
    else
        echo "Looks like an invalid env name for a mac. Valid names tend to look like prod, latest, test, etc."
        echo "Usage: pvtcl.sh [vitess env] [args to vtctl...]"
        exit 1
    fi

    HOST_NAME=$(fh -h $HOST_TYPE | head -n 1)
    if [[ "$HOST_NAME" == "" ]]; then
        echo "fh lookup of hostname failed"
        exit 1
    fi

    # Escape all the args to make sure they get passed through ok through ssh to the docker command
    escaped_args=()
    for arg in "$@"; do
        escaped_args=( "${escaped_args[@]:-}" "$(printf '%q' "$arg")" )
    done

    ssh -t "$HOST_NAME" sudo docker exec -ti vtctld /vt/bin/vtctlclient -action_timeout 5m -server localhost:15991 "${escaped_args[@]}"
    exit
fi

ENV_CONFIG_FILE="/var/config/config.services.vitess_environments_config"

ENV_EXISTS=$(jq -r ".${ENV_NAME} | select (.!=null)"  $ENV_CONFIG_FILE)

if [[ "$ENV_EXISTS" == "" ]]; then
   echo "Looks like an invalid env name. Valid names tend to look like prod, latest, test, etc."
   echo "Usage: pvtcl.sh [vitess env] [args to vtctl...]"
   exit 1
fi

# devapps have jq 1.3 which doesn't yet support "join", so xargs is used to join the host list
ZK_SERVERS=$(jq -r ".${ENV_NAME}.zk.hosts[]" $ENV_CONFIG_FILE | xargs  | sed -e 's/ /,/g')
ZK_ROOT=$(jq -r ".${ENV_NAME}.zk.global_root" $ENV_CONFIG_FILE)

# TODO(dweitzman): Ideally we'd point at vtctld here, but we don't have fixed hostnames for vtctld at this time.
/vt/bin/vtctl -topo_implementation zk2 -topo_global_server_address "$ZK_SERVERS" -topo_global_root "$ZK_ROOT" "$@"
