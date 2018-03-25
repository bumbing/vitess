# Note: the images from dockerhub don't have the additional Pinterest-specific
# govendor deps. We have a modified bootstrap percona dockerfile that fetches the
# deps we need.
# The "-pull false" prevents the default behavior of re-fetching a stock image from dockerhub.
# Pass -unit as an argument to this command to only run unit tests without integration tests.
#
# The continuous build for vitess splits the tests into 5 shards which take 20+
# minutes each when run in parallel. In serial it would take almost 2 hours to
# run all the tests.
# Vitess open source CI history: https://travis-ci.org/vitessio/vitess/builds

./docker/bootstrap/build.sh percona
bazel run --run_under="cd $PWD && " :vitess -- -pull=false -flavor percona $@
