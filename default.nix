{ buildGoModule
, version
}:

buildGoModule {
  inherit version;

  pname = "accords-mirrorrer";

  src = ./.;

  ldflags = [
    "-X" "github.com/vs49688/accords-mirrorrer.Version=0.0.${version}-nix"
  ];

  vendorHash = null;
}