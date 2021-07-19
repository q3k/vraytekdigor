{ dropbear
, zlib
, alienlib
}:

# A hacked dropbear from upstream nixpkgs, with some changes:
#  - authentication using alienlib
#  - static build, stripped as much as possible
dropbear.overrideAttrs (oa: {
    configureFlags = [
      "LDFLAGS=-static"
      "LIBS=-lalien"
      "--disable-shadow"
    ];
    buildInputs = oa.buildInputs ++ [ zlib.static alienlib ];
    dontDisableStatic = true;
    stripAllList = [ "$out" ];
    stripDebugList = [ "$out" ];
    patches = [
      ./dropbear-alienlib.patch
    ];
})
