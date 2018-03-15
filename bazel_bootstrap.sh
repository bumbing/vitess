#!/bin/bash

# Creates a bazel version of the govendor vendor.json file
# and runs gazelle to generate BUILD.bazel files for all the
# to files.
#
# You'll need to run this before building for the first time,
# or if you've added go files or changed their imports.

if [ ! -e "go_repos.bzl" ] ; then
  echo "def register_go_repos():" > go_repos.bzl
  echo "  pass" >> go_repos.bzl
fi
bazel build :generate_go_repos
cp bazel-genfiles/go_repos.bzl go_repos.bzl
chmod +w go_repos.bzl
bazel run :gazelle
