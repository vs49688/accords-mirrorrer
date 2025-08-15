{ buildGoModule
, version
}:

buildGoModule {
  inherit version;

  pname = "accords-mirrorrer";

  src = ./.;

  ldflags = [
    "-X" "git.vs49688.net/zane/accords-mirrorrer.Version=0.0.${version}-nix"
  ];

  vendorHash = null;
}
