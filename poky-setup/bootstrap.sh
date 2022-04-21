#!/usr/bin/env bash
#
set -exo pipefail

# install poky dependencies
sudo apt install -y gawk wget git diffstat unzip texinfo gcc build-essential chrpath socat cpio python3 python3-pip python3-pexpect xz-utils debianutils iputils-ping python3-git python3-jinja2 libegl1-mesa libsdl1.2-dev pylint3 xterm python3-subunit mesa-common-dev zstd liblz4-tool
# pull down poky
git clone git://git.yoctoproject.org/poky --branch honister --depth 2
cd poky
# pull down meta-raspberry layer and apply custom fixes
# - remove gstreamer (breaks the build)
# - fix up kernel revs (these revs were rebased)
git clone git://git.yoctoproject.org/meta-raspberrypi
cd meta-raspberrypi
git checkout -b honister 157f72b8084756d9ca994e9ca30a07cfc4074137
git apply ../../meta-rpi.patch
cd ..
# setup build environment
. ./oe-init-build-env build
bitbake-layers add-layer ../meta-raspberrypi
cat >> conf/local.conf <<EOF
SSTATE_MIRRORS = "\\
file://.* http://sstate.yoctoproject.org/dev/PATH;downloadfilename=PATH \\n \\
file://.* http://sstate.yoctoproject.org/3.3.5/PATH;downloadfilename=PATH \\n \\
file://.* http://sstate.yoctoproject.org/3.4.3/PATH;downloadfilename=PATH \\n \\
"
BB_HASHSERVE_UPSTREAM = "typhoon.yocto.io:8687"
MACHINE ?= "raspberrypi"
EOF
echo "this script is now done setting up."
echo "you will need to source oe-init-build-env again."
echo "then you can try something like 'bitbake core-image-minimal'"
