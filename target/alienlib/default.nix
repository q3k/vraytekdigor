{ rust-cbindgen
, rustPlatform
, lib

, buildPackages
}:

rustPlatform.buildRustPackage rec {
  pname = "alienlib";
  version = "0.1.0";

  target = "mips-unknown-linux-musl";

  src = lib.cleanSource ./.;

  cargoSha256 = "1brsd0cn8h5cvwlgpvs8ifp9yk94d7md1rma4ldq3cwj4ck9jyy6";

  nativeBuildInputs = [ buildPackages.rust-cbindgen ];

  postInstall =
    ''
      cbindgen --config bindgen.toml --output $out/include/alien.h
    '';
}
