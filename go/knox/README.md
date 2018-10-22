# Using Pinterest knox auth with vitess

## Enabling it in vtgate (or other servers)

```
  -knox_supported_roles scriptrw,longqueryrw
  -grpc_auth_mode knox
  -mysql_auth_server_impl knox
```

## Connecting to vtgate with mysql protocol + knox

Your usual `mysql_cli.py` command should work:

```
mysql_cli.py -p read-only <vtgate host>
```

If you're running on a port other than 3306, mysql_cli.py does not support that yet. You can do this instead:

NOTE(dweitzman): THIS COMMAND IS WRONG. You need to pull the username from `knox get` also. It's not scriptro.

```
mysql -v -h <vtgate host> -P <port> --protocol=TCP -u scriptro  -p`knox get mysql:rbac:scriptro:credentials | xargs echo | cut -f 2 -d\|`
```

# Connecting to vtgate with grpc + knox

**Warning:** The command below will actually fail because KnoxAuthClientCreds
requests transport security and we're not set up with TSL / encryption yet.
You can test with encryption disabled by having RequireTransportSecurity()
return false, if you're just curious to see this in action on localhost.
This presumes that `-grpc_auth_mode knox` was set in vtgate. By default, vtgate
gives any grpc user the role `unsecure_grpc_client` which we haven't forbidden
from accessing or doing anything.

```
go run vitess.io/vitess/go/cmd/vtclient -server <vtgatehost:port> -grpc_auth_knox_role longqueryro -knox_supported_roles longqueryro "select sleep(1)"
```

*TODO: Remove the need for `-knox_supported_roles` on the client side*
