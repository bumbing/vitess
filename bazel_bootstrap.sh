#!/bin/bash

# Creates a bazel version of the govendor vendor.json file
# and runs gazelle to generate BUILD.bazel files for all the
# to files.
#
# You'll need to run this before building for the first time,
# or if you've added go files or changed their imports.

if [ ! -e "bazel/go_repos.bzl" ] ; then
  echo "def register_go_repos():" > bazel/go_repos.bzl
  echo "  pass" >> bazel/go_repos.bzl
fi
bazel build --symlink_prefix=/  //bazel:generate_go_repos
cp `bazel info bazel-genfiles`/bazel/go_repos.bzl bazel/go_repos.bzl
bazel run --symlink_prefix=/  :gazelle
