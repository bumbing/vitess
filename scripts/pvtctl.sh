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
    
    # NOTE(dweitzman): The use of "sleep" here is an awkward short-term hack until we have a better
    # alternative. ssh port tunneling is used here because we don't have x509 certs we can use
    # to securely authenticate pinployees so that we could open up the vtctld grpc port at
    # the network level. One option would be to open up the grpc port only to vitess admins
    # on prod machines and maybe more widely on dev machines.
    #
    # The complication with ssh port forwarding in a bash script is forwarding a port temporarily
    # and then killing ssh. If we start a job in the background, there's some complexity in tracking
    # the job's pid and trying to make sure it gets killed when the bash script exits. Using
    # "sleep 5" here ensures that the port forwarding will stop after 5 seconds, which is reliable
    # but has two disadvantages:
    # - Since we're using a fixed port ID, you can't run pvtctl.sh commands from your laptop in
    #   too quickly. You need to wait 5 seconds after one command starts before starting another.
    # - Long-running vtctl commands could get canceled after 5 seconds. This could be risky for
    #   something like migrating served types.
    #
    # Another alternative would be to "ssh <host> docker exec vtctld ...", but I can't figure out
    # how to make that work so that command line arguments don't get messed up in transit.
    # If you want to try to fix it, make sure that you can run this command:
    # $ pvtctl.sh latest VReplicationExec tablet-1234 "select * from _vt.vreplication"
    ssh -f -o ExitOnForwardFailure=yes -L 16127:localhost:15991 "$HOST_NAME" sleep 5
    # NOTE(dweitzman): This "go run" command will only work if your GOPATH
    # is all set up or alternative if you create a go.mod file using
    # ~/code/dw/scripts/vt_fix_gomod.sh
    #
    # Longer term, we should probably put a homebrew package at
    # https://phabricator.pinadmin.com/diffusion/BREW/browse/master/
    #
    # We could actually even use an off-the-shelf vtctlclient here if upstream
    # (open source) vitess published a homebrew package.
    go run ./go/cmd/vtctlclient -server localhost:16127 "$@"
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
