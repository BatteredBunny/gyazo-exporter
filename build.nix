{buildGoModule}:
buildGoModule {
  src = ./.;

  name = "gyazo-exporter";
  vendorSha256 = "sha256-2caHIvsCdgDcQkquw4887cOLolRLd9+0ysDE5aA4KVU=";

  ldflags = [
    "-s"
    "-w"
  ];
}
