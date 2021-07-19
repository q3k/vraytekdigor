{ buildGoModule
, lib
, ... }:

buildGoModule {
  name = "fwtool";
  version = "0.1.0";

  src = lib.cleanSource ./.;

  vendorSha256 = "0vkr1wn02li57whq1c2k28m0vn42gjckprvchb2sqqk2gsnii8x9";
  runVend = true;
}
