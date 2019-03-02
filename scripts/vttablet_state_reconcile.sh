#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# find all masters from all keyspaces
# sample from: hzhou@vtctld-latest-0a03599a:~$ sudo docker exec -ti vtctld /vt/bin/vtctlclient -server localhost:15991 ListAllTablets | grep master
# test-0000000202 patio -80 master coladevdb-2-2.ec2.pin220.com:15103 coladevdb-2-2.ec2.pin220.com:3306 []
# test-0000000301 patio 80- master coladevdb-3-1.ec2.pin220.com:15103 coladevdb-3-1.ec2.pin220.com:3306 []
# test-0000001024 patiogeneral 0 master devpatiogeneraldb-1-24.ec2.pin220.com:15103 devpatiogeneraldb-1-24.ec2.pin220.com:3306 []
# test-0000001046 patio 0 master coladevdb-1-46.ec2.pin220.com:15103 coladevdb-1-46.ec2.pin220.com:3306 []
MASTERS="$(/vt/bin/vtctlclient -server localhost:15991 ListAllTablets | grep master | cut -d ' ' -f 5 | cut -d '.' -f 1)"
YESTERDAY=$(date --date=yesterday "+%Y-%m-%d")
TODAY=$(date "+%Y-%m-%d")

rm -f /vt/logs/sync-$YESTERDAY-*.log
for TABLET in $MASTERS; do
    # do a state sync between vitess topology and configv3 meta
    python -m vitess_utils.vitess -a vttabletstatechange --vtctld localhost:15991 --host $TABLET 2>&1 >> /vt/logs/"sync-$TODAY-$TABLET.log"
done
