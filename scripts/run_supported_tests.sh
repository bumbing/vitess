# NOTE(dweitzman): These tests pass on my devapp but fail on my mac laptop:
#
# //go/vt/vtctld:go_default_test
# //go/vt/vttablet:go_default_test
# //go/vt/tabletmanager:go_default_test
# //go/vt/worker:go_default_test
#
# All fail for the same reason, which seems fixable at some point:
#
# panic: agent.Start(cell:"cell1" uid:42 ) failed: FullyQualifiedHostname: failed to lookup the IP of this machine's hostname (dweitzman-0S7H03Y.dyn.pinadmin.com): lookup dweitzman-0S7H03Y.dyn.pinadmin.com: no such host

bazel test -- go/... `./scripts/print_not_yet_supported_tests.sh | xargs -n 1 | xargs -I{} echo -{} | xargs echo`
