{ runCommand
, writeText
, unzip
, squashfs-tools-ng
, git

, fwtool
, tarmanip
}:

{

  # Make a customized firmware from a vendor distribution zip.
  #
  # This extracts the original firmware from zip, extracts the squashfs blob
  # via fwtool, converts the squashfs to a tarball, applies a tarmanip script
  # on the rootfs tarball, converts the rootfs tarball back into squashfs, and
  # finally inserts the squashfs back into the original file, yielding a custom
  # firmware.
  makeCustomFirmware = {
    # The shorthand name of the device as seen in update files, eg. v167 for
    # Vigor 167 devices.
    modelSlug,

    # The git revision of the vraytekdigor repository, used as version strings.
    # Can be obtained using lib.gitHash.
    gitRevision,

    # Original zip from vendor, containing modelSlut_*.all.
    originalZip,

    # tarmanip script to apply to this firmware.
    script,
  }: let

    # Extract the firmware .zip. It contains two files, named the
    # following way:
    #
    #   ${modelSlug}_${version}.(all|rst)
    #
    # For Vigor 167 modems, `modemSlug` is v167.
    # `version` has been observed to be 50 for version 5.0, but
    # could technically be any arbistrary string.
    # The file extension (all or rst) makes the firmware either
    # keep the existing settings (all) or reset them to defaults
    # (rst). The files themselves are identical, only the
    # extension is parsed.
    #
    # After the zip is extracted, we run fwtool and sqfs2tar to
    # retrieve the builtin squashfs and convert it to a tarball
    # for further processing.
    unpacked = runCommand "vraytek-extracted" {
      nativeBuildInputs = [
        unzip
        fwtool
        squashfs-tools-ng
      ];
    } ''
        unzip ${originalZip}

        mkdir -p $out
        mv ${modelSlug}_*all $out/original.bin

        fwtool -squash_out squashfs.bin -fw_in $out/original.bin
        sqfs2tar squashfs.bin > $out/rootfs.tar
      '';
    
    # Build custom firmware. This is done by applying the given
    # script using tarmanip on the rootfs tarball, then repacking
    # it back using fwtool.
    customized = runCommand "vraytek-custom-${gitRevision}" {
      # This makes tar2sqfs give us a nice non-epoch root
      # filesystem timestamp, to make things slighly nicer. The
      # particular time doesn't matter, just as long as it's not
      # ancient history.
      SOURCE_DATE_EPOCH = "1626462487";

      nativeBuildInputs = [
        fwtool
        tarmanip
        squashfs-tools-ng
      ];
    } ''
        tarmanip \
            -in ${unpacked}/rootfs.tar \
            -out rootfs.tar \
            -script ${writeText "script.pb.text" script}
        tar2sqfs \
           -c lzma \
           -X level=9,extreme \
           -T -x -d mode=0755 \
           -j 8 \
           squashfs.bin < rootfs.tar

        mkdir -p $out
        fwtool \
          -fw_in ${unpacked}/original.bin \
          -squash_in squashfs.bin \
          -fw_out $out/${modelSlug}_cfw${gitRevision}.all
      '';
  in customized;

  # Get 8-character shorthash of the commit checked out in this repository.
  # IFD, very bad, not good.
  gitHash = let
    ifd = runCommand "gitHashIFD" {
      nativeBuildInputs = [ git ];
    } ''
        cd ${builtins.path { path = ./.git; name = "git"; }}
        short=$(git rev-parse --short=8 HEAD)
        mkdir -p $out
        cat >$out/default.nix <<EOF
        {
           short = "$short";
        }
        EOF
      '';
  in (import "${ifd}").short;

}
