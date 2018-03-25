echo cgzip cmd/vtclient mysql mysql/endtoend vt/logutil vt/mysqlctl vt/topo/consultopo vt/topo/etcd2topo vt/topo/zk2topo vt/vtqueryserver vt/vttablet/endtoend vt/zkctl | xargs -n 1 | xargs -I{} echo //go/{}:go_default_test | xargs echo

# Output from failing tests, interleaved with occasionally commentary on why they're failing:

cat > /dev/null <<EOF
==================== Test output for //go/vt/zkctl:go_default_test:
E0325 01:56:14.845566      14 zkctl.go:168] mkdir /vt: read-only file system
--- FAIL: TestLifeCycle (0.00s)
	zkctl_test.go:38: Init() err: mkdir /vt: read-only file system
FAIL
================================================================================

#
# Haven't tried this, but it seems like depending on the output from the mysqlctl
# go_binary instead of assuming that it'll be at $VTROOT/bin would help.
#
==================== Test output for //go/cmd/vtclient:go_default_test:
E0325 01:56:14.853234      14 local_cluster.go:216] Mysqlctl failed to start: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
--- FAIL: TestVtclient (0.00s)
	vtclient_test.go:78: InitSchemas failed: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
FAIL
================================================================================

#
# We don't need etcd support, so I think we should leave this permanently blacklisted.
#
==================== Test output for //go/vt/topo/etcd2topo:go_default_test:
--- FAIL: TestEtcd2Topo (0.00s)
	server_test.go:63: failed to start etcd: exec: "etcd": executable file not found in $PATH
FAIL
================================================================================

#
# vt/logutil seems to want to re-execute a subprocess version of the test,
# or something like that. The subprocess command is failing.
#
==================== Test output for //go/vt/logutil:go_default_test:
--- FAIL: TestConsoleLogger (0.00s)
	console_logger_test.go:69: cmd.Wait() error: exit status 1
--- FAIL: TestTeeConsoleLogger (0.00s)
	console_logger_test.go:69: cmd.Wait() error: exit status 1
FAIL
================================================================================

#
# vt/mysqlctl is trying to read some my.cnf files from VTROOT.
# Because of test sandboxing, --test_env=VTROOT=~/code/vitess will
# not immediately fix the problem. Some refactoring of the deps should
# be possible, though.
#
==================== Test output for //go/vt/mysqlctl:go_default_test:
--- FAIL: TestMycnf (0.00s)
	mycnf_test.go:51: err: open ~/code/vitess/src/vitess.io/vitess/config/mycnf/default.cnf: no such file or directory
	mycnf_test.go:68: failed reading, err value for key 'server-id' not set and no default value set
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x8 pc=0x84e2c6]

goroutine 7 [running]:
testing.tRunner.func1(0xc4201b41e0)
	GOROOT/src/testing/testing.go:711 +0x2d2
panic(0x8d9d40, 0xc3db20)
	GOROOT/src/runtime/panic.go:491 +0x283
vitess.io/vitess/go/vt/mysqlctl.TestMycnf(0xc4201b41e0)
	go/vt/mysqlctl/mycnf_test.go:73 +0x646
testing.tRunner(0xc4201b41e0, 0x9ab0c8)
	GOROOT/src/testing/testing.go:746 +0xd0
created by testing.(*T).Run
	GOROOT/src/testing/testing.go:789 +0x2de
================================================================================

==================== Test output for //go/vt/vtqueryserver:go_default_test:
E0325 01:56:15.048983      14 local_cluster.go:216] Mysqlctl failed to start: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
could not launch mysql: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
================================================================================

#
# We don't need consul support, so I think we should leave this permanently blacklisted.
#
==================== Test output for //go/vt/topo/consultopo:go_default_test:
--- FAIL: TestConsulTopo (0.00s)
	server_test.go:78: failed to start consul: exec: "consul": executable file not found in $PATH
FAIL
================================================================================

==================== Test output for //go/vt/vttablet/endtoend:go_default_test:
E0325 01:56:15.404370      14 local_cluster.go:216] Mysqlctl failed to start: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
could not launch mysql: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
================================================================================

#
# Looks like a possible sandboxing issue?...
#
==================== Test output for //go/vt/topo/zk2topo:go_default_test:
E0325 01:56:15.403694      14 zkctl.go:168] mkdir /vt: read-only file system
F0325 01:56:15.404144      14 zkctl_local.go:40] zkd.Init(1, 6702) failed: mkdir /vt: read-only file system
================================================================================

==================== Test output for //go/mysql/endtoend:go_default_test:
E0325 01:56:15.609298      14 local_cluster.go:216] Mysqlctl failed to start: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
could not launch mysql: fork/exec ~/code/vitess/bin/mysqlctl: no such file or directory
================================================================================

#
# Hmm. Could be a sandboxing issue?
#
==================== Test output for //go/cgzip:go_default_test:
--- FAIL: TestEofAndData (0.00s)
	eof_read_test.go:39: Cannot read file /home/dweitzman/.cache/bazel/_bazel_dweitzman/fbc7247e6832729f52b7565f332e027c/bazel-sandbox/5413822871940806006/execroot/__main__/bazel-out/k8-fastbuild/bin/go/cgzip/linux_amd64_stripped/go_default_test.runfiles/__main__/data/test/cgzip_eof.gz: open /home/dweitzman/.cache/bazel/_bazel_dweitzman/fbc7247e6832729f52b7565f332e027c/bazel-sandbox/5413822871940806006/execroot/__main__/bazel-out/k8-fastbuild/bin/go/cgzip/linux_amd64_stripped/go_default_test.runfiles/__main__/data/test/cgzip_eof.gz: no such file or directory
FAIL
================================================================================

#
# go/mysql has a dependency on mysqld. It'll pass if you have mysql install
# and set --test_env=VT_MYSQL_ROOT=/usr
# Ideally in a bazel world we'd have some kind of hermetic mysql from external
# repos so you could run this without installing mysql outside of bazel.
#
==================== Test output for //go/mysql:go_default_test:
--- FAIL: TestServer (0.00s)
	server_test.go:157: listening on address '127.0.0.1' port 37173
	server_test.go:770: binaryPath failed: mysql not found in any of ~/code/vitess/{sbin,bin}
--- FAIL: TestClearTextServer (0.00s)
	server_test.go:157: listening on address '127.0.0.1' port 39087
	server_test.go:770: binaryPath failed: mysql not found in any of ~/code/vitess/{sbin,bin}
--- FAIL: TestDialogServer (0.00s)
	server_test.go:157: listening on address '127.0.0.1' port 44379
	server_test.go:770: binaryPath failed: mysql not found in any of ~/code/vitess/{sbin,bin}
--- FAIL: TestTLSServer (0.30s)
	server_test.go:770: binaryPath failed: mysql not found in any of ~/code/vitess/{sbin,bin}
FAIL
EOF



