#!/bin/bash 

# Steps to fix topo:
# 0) Make sure there are no planned reparents or reshardings in progress and until the operation ends
# 1) Run this script to generate a list of command to copy the data from one place to another
# 2) vtctl UpdateCellInfo [-root <new root>] <cell>
# 3) Restart gates carefully, making sure they come up happy
# 4) Restart a tablet, making sure it comes up happy. Then the rest.
#    $ sudo initctl stop vttablet-3306 && sudo initctl start vttablet-3306
# 5) Restart vtctld
# 6) Clear out tablet info and SrvVSchema from the global keyspace
#
# If the operation needs to be retried, zk rm -r /zk/path can be used
# to clear out the new path and start again. But USE CAUTION IF DELETING!

zk_raw_cmd="${ZK_CMD:-go run ./go/cmd/zk}"
server="${ZK_SERVER:-vitess-infra-zookeeper-dev-001:2181}"
zkcmd="$zk_raw_cmd -server $server"
old_path="${ZK_OLD_PATH:-/vitess/m10n/dev/global}"
new_path="${ZK_NEW_PATH:-/vitess/m10n/dev/test}"

tablets=$($zkcmd ls $old_path/tablets/)
keyspaces=$($zkcmd ls $old_path/keyspaces/)

for keyspace in $keyspaces; do
  echo "$zkcmd" cp "$old_path/keyspaces/$keyspace/SrvKeyspace" "$new_path/keyspaces/$keyspace/SrvKeyspace"

  shards=$($zkcmd ls $old_path/keyspaces/$keyspace/shards)
  for shard in $shards; do
      echo "$zkcmd" cp "$old_path/keyspaces/$keyspace/shards/$shard/ShardReplication" "$new_path/keyspaces/$keyspace/shards/$shard/ShardReplication"
  done
done

for tablet in $tablets; do
  echo "$zkcmd" cp "$old_path/tablets/$tablet/Tablet" "$new_path/tablets/$tablet/Tablet"
done

echo "$zkcmd" cp "$old_path/SrvVSchema" "$new_path/SrvVSchema"

echo
echo To clean up data after the migration:

for keyspace in $keyspaces; do
  echo "$zkcmd" rm "$old_path/keyspaces/$keyspace/SrvKeyspace"

  shards=$($zkcmd ls $old_path/keyspaces/$keyspace/shards)
  for shard in $shards; do
      echo "$zkcmd" rm "$old_path/keyspaces/$keyspace/shards/$shard/ShardReplication"
  done
done

for tablet in $tablets; do
  echo "$zkcmd" rm -r "$old_path/tablets"
done

echo "$zkcmd" rm "$old_path/SrvVSchema"
