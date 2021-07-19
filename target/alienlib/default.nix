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

  cargoSha256 = "0sv7jdia6kknz66z4bga129apm8gsbn9mps4nppj34g2n4295q8n";

  nativeBuildInputs = [ buildPackages.rust-cbindgen ];

  postInstall =
    ''
      cbindgen --config bindgen.toml --output $out/include/alien.h
    '';
}
