{ buildGoModule
, protobuf
, go-protobuf
, lib
, ... }:

buildGoModule {
  name = "tarmanip";
  version = "0.1.0";

  src = lib.cleanSource ./.;

  nativeBuildInputs = [
    protobuf go-protobuf
  ];

  preBuild = ''
    go generate
  '';

  vendorSha256 = "1p5zzx7sb6hyrx8xr5wi67jb7wh7am7ah67k073hnq43ha03dn1l";
  runVend = true;
}
