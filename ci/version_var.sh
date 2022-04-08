#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Release Parameters
BASE_VERSION=0.1.9
IS_RELEASE=false

# Project Parameters
SOURCE_REPO=mediator
PKG_NAME=mediator
RELEASE_REPO=ghcr.io/trustbloc
SNAPSHOT_REPO=ghcr.io/trustbloc-cicd

if [ ${IS_RELEASE} = false ]
then
  EXTRA_VERSION=snapshot-$(git rev-parse --short=7 HEAD)
  PROJECT_VERSION=${BASE_VERSION}-${EXTRA_VERSION}
  PROJECT_PKG_REPO=${SNAPSHOT_REPO}
else
  PROJECT_VERSION=${BASE_VERSION}
  PROJECT_PKG_REPO=${RELEASE_REPO}
fi

export MEDIATOR_TAG=$PROJECT_VERSION
export MEDIATOR_PKG=${PROJECT_PKG_REPO}/${PKG_NAME}
