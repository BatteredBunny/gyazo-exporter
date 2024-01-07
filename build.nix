{buildGoModule}:
buildGoModule {
  src = ./.;

  name = "gyazo-exporter";
  vendorHash = "sha256-2caHIvsCdgDcQkquw4887cOLolRLd9+0ysDE5aA4KVU=";

  ldflags = [
    "-s"
    "-w"
  ];
}
