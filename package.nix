{
	buildGoModule,
}: buildGoModule {
	pname = "wwise-cli";
	version = "0.1.0";
	
	src = ./.;

	vendorHash = "sha256-twRZm4LmKwAj3yK1i20Qqm+hP8jMPjSGMy/UexzfZTU=";
}
