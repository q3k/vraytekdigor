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

  vendorSha256 = "0snc9dqfr95535zds9dim6gsafxbl5vh0j7vlcr55lqjwmhja9f8";
  proxyVendor = true;
}
