#!/bin/bash

echo GIT_REVISION $(git rev-parse HEAD)
echo GIT_BRANCH $(git rev-parse --abbrev-ref HEAD)
echo JENKINS_BUILD_NUM $BUILD_ID
