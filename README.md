# VraytekDigor

This is q3k's experimental Custom Firmware (CFW) project for his Draytek Vigor
167 VDSL modem.

## Disclaimer

Before we go any further, a few things must be stated:

1. This is my personal project. You probably shouldn't use this. I recommend
   you don't use it. This might brick your devices. This probably also voids
   any warranty you have.
2. No, seriously, don't depend on it. This is meant as an exploratory project
   for people interested in embedded development, not a ready end-user product.
   It probably doesn't provide anything useful to most people.
3. This is not supported by DrayTek, not supported by me, not supported by
   anyone or anything. You are on your own.

# Features

The custom firmware is based on firmware version 5.0.1 of the Draytek Vigor 167
VDSL modem, with the following modifications:

 - A modern Dropbear server (2020.81) running /bin/sh instead of a limited
   shell.
 - SSH password authentication using the same passwords as the admin panel.
 - SSH public key authentication, configured via a new field in the admin
   panel.

# Usage

## Building

You will need [nix or NixOS](https://nixos.org/download.html).

    $ # Build everything. This will take a bit on first run, as a bunch of
    $ # toolchains for MIPS must be built...
    $ export NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1
    $ nix-build -A cfw
    /nix/store/v7ihha3j4j2swz3ildaylz9vqaqrl78r-vraytek-custom-518f426f
    $ # Note: your hash will differ, as it's based on the Git revision of this
    $ # repository at build time.

    $ ls /nix/store/v7ihha3j4j2swz3ildaylz9vqaqrl78r-vraytek-custom-518f426f
    v167_cfw518f426f.all
    $ # Note: your firmware name will differ, as it contains the Git revision
    $ # of this repository at build time.

Once you have a firmware file like `v167_cfw518f426f.all`, it can be uploaded
to the web interface under System Maintenance -> Firmware.

## SSH

After installing and rebooting to the new firmware, you should be able to SSH
as admin onto the modem, using the same password as for the web panel.

    $ ssh admin@192.0.2.1
    admin@192.0.2.1's password:
     _   __             ______    __     ___  _
    | | / /______ ___ _/_  __/__ / /__  / _ \(_)__ ____  ____
    | |/ / __/ _ `/ // // / / -_)  '_/ / // / / _ `/ _ \/ __/
    |___/_/  \_,_/\_, //_/  \__/_/\_\ /____/_/\_, /\___/_/
                 /___/ CFW, git rev 518f426f /___/
    # uname -a
    Linux draytek 3.18.21 #4 SMP Fri May 7 16:22:06 CST 2021 mips unknown

For public key authentication, add an SSH admin key in the web panel, in System
Maintenance -> Accounts -> SSH Public Key. **You will need to enter your
current password and a new password twice (can be the same as the existing
password) alongside the SSH public key to save it.** This is due to how the
behaviour of the password form is implemented in the web interface.

# Customization

Currently, only a 'model' custom firmware is built by this repository, defined
in default.nix. Poke around this file (especially the 'script') to add your own
modifications. It should be documented well enough to understand what's going
on and why.

In the future, it might be possible to import this repository into another Nix
derivation and extend it (this can already be somewhat done using lib.nix's
makeCustomFirmware, but that means you have to reimplement all the basic
modifications as per default.nix).

# License and Binaries

This repository contains only source code licensed under an open source license
(MIT license, see COPYING). This does not make the resulting build artifacts
open source software, though.

The original DrayTek firmware is a proprietary piece of software not
distributed under an open source license. I don't have any rights to
redistribute it, and probably neither do you. Custom firmware built by the code
in this repository derives from that original firmware. To be clear, this
repository does not contain either original nor custom firmware, just code
which in turn, when run, builds custom firmware.

In addition, the original DrayTek firmware seemingly contains compiled code of
works originally licensed under copyleft licenses like the GPL, and no
correspondig source code is available at the time of writing. This means that
redistributing the firmware might infringe not only on the rights of proprietary
DrayTek code, but the authors of what appears to be code redistributed under
these copyleft licenses.

All in all, custom firmwares are a legally gray area, and you should do your
own research on how this concerns you, the potential user of anything built by
this codebase.

Considering the above, **no binary builds of the custom firmware will ever be
provided**. You must build everything yourself, and do your own legal research
on whether whatever you're doing is even legal.

# Vigor 167 - General Software/Hardware information

The modem runs on a EcoNet EN751627 SOC (2 cores / 4 threads), has a bit over
100M of RAM available to Linux and 128MB of flash (split into a set of
primary/secondary partitions).

**Speculation below:**

DrayOS 5 is based on Linux 3.18.21. It seems to have been bult from a buildroot
BSP that might have also been used in previous DrayOS builds? Hard to tell.

The firmware contains a lot of references to a 'TC3162' but that seems to be a
red herring, what looks like a standalone ADSL SoC from Trendchip that now has
become a standardized userland interface for some class of DSL home gateways?
It seems to pop up across various vendors of different classes of DSL modems
across years of random public projects.  People have been writing
[parsers](https://github.com/zoka/modstat/blob/master/modstat.py) for `cat
/proc/tc3162/adsl_stats` for a while now. A whole bunch of kernel modules (to
which there are no sources...) interact with and implement this mysterious
tc3162 world, including what seems to be the main ethernet/switch driver
(`eth.ko`). The switch chip / MAC itself might be a MT7530.

More research would have to be done into the actual driver/firmware stack
involved to make an educated opinion on whether something like OpenWRT could be
ported to this device. Having a reliable root shell helps :).
