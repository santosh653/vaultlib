# ----------------------------------------------------------------------------
#
# Package		: vaultlib
# Version		: v0.6.1, v0.6.0, 0.5.4
# Source repo	: https://github.com/mhamann/vaultlib
# Tested on		: UBI 8.4
# Script License: Apache License, Version 2 or later
# Maintainer	: Santosh Kulkarni<santoshkulkarni70@gmail.com>/Priya Seth<sethp@us.ibm.com>
#
# Disclaimer: This script has been tested in non-root mode on given
# ==========  platform using the mentioned version of the package.
#             It may not work as expected with newer versions of the
#             package and/or distribution. In such case, please
#             contact "Maintainer" of this script.
#
# ----------------------------------------------------------------------------
#!/bin/bash

if [ -z "$1" ]; then
  export VERSION=master
else
  export VERSION=$1
fi

if [ -d "vaultlib" ] ; then
  rm -rf vaultlib 
fi

# Dependency installation
sudo dnf install -y git golang 
# Download the repos
git clone  https://github.com/mhamann/vaultlib


# Build and Test vaultlib
cd vaultlib
git checkout $VERSION
ret=$?
if [ $ret -eq 0 ] ; then
 echo "$Version found to checkout "
else
 echo "$Version not found "
 exit
fi

#Build and test
go get -v -t ./...

ret=$?
if [ $ret -ne 0 ] ; then
  echo "Build failed "
else
  go test -v ./...
  if [ $ret -ne 0 ] ; then
    echo "Tests failed "
  else
    echo "Build & unit tests Success "
  fi
fi
