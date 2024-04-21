{ buildGoModule
, lib
, ... }:

buildGoModule {
  name = "fwtool";
  version = "0.1.0";

  src = lib.cleanSource ./.;

  vendorHash = "sha256:1b6sa6q30kid89mc3f06ncnl0735h7ws5dzcknq18zw8z68pqskc";
  proxyVendor = true;
}
