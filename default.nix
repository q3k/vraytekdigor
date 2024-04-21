let
  # Make a nixpkgs configured to cross-compile for mips, and with our custom
  # packages overlaid.
  nixpkgs = import (builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/19081514c247fbe95d0cd9094e530cafd43dbe7f.tar.gz";
    sha256 = "sha256:1kaghqzfp7f6pll96l5dxz44j3qrqx10w1vqml6bziwr1gywngfl";
  }) {
    crossSystem = {
      libc = "musl";
      config = "mips-unknown-linux-musl";
      openssl.system = "linux-generic32";
      withTLS = true;
      gcc = {
        abi = "32";
        arch = "mips32";
      } ;
      linuxArch = "mips";
      bfdEmulation = "elf32btsmip";
      rustc.config = "mips-unknown-linux-musl";
    };

    # Add our own tools, so that pkgs.* contains both upstream nixpkgs but also
    # the following:
    overlays = [
      (self: super: {
        fwtool = super.callPackage ./build/fwtool {};
        tarmanip = super.callPackage ./build/tarmanip {};
        alienlib = super.callPackage ./target/alienlib {};
        cfwdropbear = super.callPackage ./target/dropbear {};
        cfwlib = super.callPackages ./lib.nix {};
      })
    ];
  };

in

# Use the above nixpkgs. Everything below is in MIPS-land unless defined
# otherwise..
with nixpkgs; with builtins; let

  # Make a MOTD file containing the current revision.
  motd = pkgs.writeText "motd"
    ''
       _   __             ______    __     ___  _              
      | | / /______ ___ _/_  __/__ / /__  / _ \(_)__ ____  ____
      | |/ / __/ _ `/ // // / / -_)  '_/ / // / / _ `/ _ \/ __/
      |___/_/  \_,_/\_, //_/  \__/_/\_\ /____/_/\_, /\___/_/
                   /___/ CFW, git rev ${cfwlib.gitHash} /___/
    '';

in

{
  # Build 'model' custom firmware which gives root SSH access to /bin/sh and
  # implements auth/ssh pubkey authorization compatible with the DrayOS web
  # panel.
  cfw = cfwlib.makeCustomFirmware {
    modelSlug = "v167";
    gitRevision = cfwlib.gitHash;
    originalZip = pkgs.fetchurl {
      name = "vigor167-5.0.1";
      # Original URL, taken down.
      #url = "http://draytek.com/download_de/Firmwares-Modem/Vigor160-Serie/Vigor167/Vigor167_v5.0.1_STD.zip";
      # My mirror.
      url = "https://object.ceph-eu.hswaw.net/q3k-personal/2ebc6fa7ae6ce1c3a8fa6c94e0b9e8b386ad02bc7b953d30730304a30f5855a9.zip";
      sha256 = "1aamb07s6103fcq3v5bvph1av1mkx2wy153czalc7qbcmsknzg1f";
    };
  
    # tarmanip script describing the actual firmware modification. See
    # tarmanip/README.md for more information about the format of this script.
    script = ''
      ## Lighten up firmware by removing unused crap.
      # Homenas stuff - this is being started by rcS, but we don't care.
      change { remove { path: "/userfs/bin/homenas" } }
      change { remove { path: "/userfs/bin/ntfs-3g" } }
      change { remove { path: "/userfs/bin/mkntfs" } }
      change { remove { path: "/userfs/bin/ntfslabel" } }
      change { remove { path: "/userfs/bin/mobile-manager" } }
      # Old/BSP? webserver, not even started.
      change { remove { path: "/boaroot/" recursive: true } }
      
      # Modify firmware version reported.
      change { jsonpatch {
        path: "/usr/etc/draytek_info"
        source: "${writeText "patch.json" (toJSON [
          {
            op = "replace";
            path = "/fw_ver";
            value = "5.0.1 VraytekDigor ${cfwlib.gitHash}";
          }
        ])}"
      } }
  
      # Add SSH public key field to admin configuration.
      # This modifies the 'seeds' config form description, which is reflected in
      # the admin panel and the draysh shell. By default, adding a field doesn't
      # change any behaviour. Instead, this SSHKey field is consumed by alienlib
      # which uses it to provide ssh pubkey auth to dropbear.
      change { jsonpatch {
        path: "/usr/etc/seeds/config_form.json"
        source: "${writeText "patch.json" (toJSON [
          {
            op = "add";
            path = "/0ADM_PASSWORD/form/-";
            value = {
              name = "SSHKey";
              title = "SSH Public Key";
              method = "text";
              method_opt = [];
              flag = [];
              style = "generic";
              pattern = "";
              tip = "SSH public key to tie to this account (eg. ssh-ed25519 ... foo@example.com)";
              data_type = "string";
            };
          }
          {
            op = "add";
            path = "/0ADM_PASSWORD/default_form/SSHKey";
            value = "";
          }
          {
            op = "add";
            path = "/0ADM_PASSWORD/default_data/0/SSHKey";
            value = "";
          }
        ])}"
      } }
  
      # Add a spiffy MOTD.
      change { create {
        path: "/usr/etc/motd"
        mode: 0644
      } }
      change { write {
        path: "/usr/etc/motd"
        source: "${motd}"
      } }
  
      # Serve /bin/sh instead of draysh over ssh/telnet.
      change {
        binreplace {
          path: "/usr/etc/lib/libfeeds.so.1.0.0"
          from: "/usr/bin/draysh\" >> /etc/passwd\x00"
          to:   "/bin/sh\" >> /etc/passwd\x00"
        }
      }
  
      # Allow password login over SSH.
      change { binreplace {
          path: "/usr/etc/lib/libfeeds.so.1.0.0"
          from: "/userfs/bin/dropbear -s -p %d"
          to: "/userfs/bin/dropbear -p %d"
      } }
  
      # Replace dropbear with newer version which also supports auth against
      # draytek config (seeds) via alienlib.
      change { write {
        path: "/userfs/bin/dropbear"
        source: "${cfwdropbear}/bin/dropbear"
      } }
    '';
  };

  # Expose useful tools for hacking around.
  inherit (pkgs.buildPackages) fwtool tarmanip;
}
