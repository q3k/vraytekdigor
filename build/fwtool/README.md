# fwtool

fwtool manipulates binary images for Draytek Vigor 167 modems, and possibly other devices which use the same file format (2RDH/HDR2).

## Usage

fwtool can function in two modes:

**1. Extract squashfs of root filesystem from an image:**

    fwtool -fw_in v167_50.all -squash_out rootfs.squash

**2. Replace the rootfilesystem squashfs in a firmware file, generating a new one:**

    fwtool -fw_in v167_50.all -squash_in rootfs-modified.squash -fw_out v167_50modified.all

It does not inspect the contained squashfs image at all, and treats it as an opaque blob. For manipulating this image, something like squashfs-tools-ng could be used.
