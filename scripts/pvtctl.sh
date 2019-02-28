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
    # Use a different local port number for each host type to make sure
    # we don't get confused an send a command intented to latest when
    # the proxy is pointing to prod.
    if [[ "$ENV_NAME" == "prod" ]]; then
        HOST_TYPE="vtctld-prod"
        TUNNEL_PORT=16127
    elif [[ "$ENV_NAME" == "shadow" ]]; then
        HOST_TYPE="vtctld-shadow"
        TUNNEL_PORT=16128
    elif [[ "$ENV_NAME" == "latest" ]]; then
        HOST_TYPE="vtctld-latest"
        TUNNEL_PORT=16129
    elif [[ "$ENV_NAME" == "test" ]]; then
        HOST_TYPE="vtctld-test"
        TUNNEL_PORT=16130
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
    
    # Start an ssh tunnel specific to this host, or reuse one that already exists.
    # Each host type users a separate local port to make sure there's no risk of
    # getting confused about which type of vtctl you're sending commands to.
    if ! ssh -q -O check -S ~/.ssh/controlmaster-pvtctl-"$HOST_NAME" "$HOST_NAME" > /dev/null 2>&1; then
       ssh -q -M -L $TUNNEL_PORT:localhost:15991 -f -o ExitOnForwardFailure=yes -S ~/.ssh/controlmaster-pvtctl-"$HOST_NAME" "$HOST_NAME" sleep 120 > /dev/null 2>&1
    fi

    # NOTE(dweitzman): This "go run" command will only work if your GOPATH
    # is all set up or alternative if you create a go.mod file using
    # ~/code/dw/scripts/vt_fix_gomod.sh
    #
    # Longer term, we should probably put a homebrew package at
    # https://phabricator.pinadmin.com/diffusion/BREW/browse/master/
    #
    # We could actually even use an off-the-shelf vtctlclient here if upstream
    # (open source) vitess published a homebrew package.
    go run ./go/cmd/vtctlclient -server localhost:$TUNNEL_PORT "$@"
    # Close the port forwarding.
    ssh -q -O exit -S ~/.ssh/controlmaster-pvtctl-"$HOST_NAME" "$HOST_NAME"
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
