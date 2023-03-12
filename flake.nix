{
  description = "Transfer.sh";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      transfer-sh = pkgs: pkgs.buildGoModule {
        src = self;
        name = "transfer.sh";
        vendorSha256 = "sha256-bgQUMiC33yVorcKOWhegT1/YU+fvxsz2pkeRvjf3R7g=";
      };
    in

      flake-utils.lib.eachDefaultSystem (
        system:
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in
            rec {
              packages = flake-utils.lib.flattenTree {
                transfer-sh = transfer-sh pkgs;
              };
              defaultPackage = packages.transfer-sh;
              apps.transfer-sh = flake-utils.lib.mkApp { drv = packages.transfer-sh; };
              defaultApp = apps.transfer-sh;
            }
      ) // rec {

        nixosModules = {
          transfer-sh = { config, lib, pkgs, ... }: with lib; let
            RUNTIME_DIR = "/var/lib/transfer.sh";
            cfg = config.services.transfer-sh;

            general_options = {

              enable = mkEnableOption "Transfer.sh service";
              listener = mkOption { default = 80; type = types.int; description = "port to use for http (:80)"; };
              profile-listener = mkOption { default = 6060; type = types.int; description = "port to use for profiler (:6060)"; };
              force-https = mkOption { type = types.nullOr types.bool; description = "redirect to https"; };
              tls-listener = mkOption { default = 443; type = types.int; description = "port to use for https (:443)"; };
              tls-listener-only = mkOption { type = types.nullOr types.bool; description = "flag to enable tls listener only"; };
              tls-cert-file = mkOption { type = types.nullOr types.str; description = "path to tls certificate"; };
              tls-private-key = mkOption { type = types.nullOr types.str; description = "path to tls private key "; };
              http-auth-user = mkOption { type = types.nullOr types.str; description = "user for basic http auth on upload"; };
              http-auth-pass = mkOption { type = types.nullOr types.str; description = "pass for basic http auth on upload"; };
              http-auth-htpasswd = mkOption { type = types.nullOr types.str; description = "htpasswd file path for basic http auth on upload"; };
              http-auth-ip-whitelist = mkOption { type = types.nullOr types.str; description = "comma separated list of ips allowed to upload without being challenged an http auth"; };
              ip-whitelist = mkOption { type = types.nullOr types.str; description = "comma separated list of ips allowed to connect to the service"; };
              ip-blacklist = mkOption { type = types.nullOr types.str; description = "comma separated list of ips not allowed to connect to the service"; };
              temp-path = mkOption { type = types.nullOr types.str; description = "path to temp folder"; };
              web-path = mkOption { type = types.nullOr types.str; description = "path to static web files (for development or custom front end)"; };
              proxy-path = mkOption { type = types.nullOr types.str; description = "path prefix when service is run behind a proxy"; };
              proxy-port = mkOption { type = types.nullOr types.str; description = "port of the proxy when the service is run behind a proxy"; };
              ga-key = mkOption { type = types.nullOr types.str; description = "google analytics key for the front end"; };
              email-contact = mkOption { type = types.nullOr types.str; description = "email contact for the front end"; };
              uservoice-key = mkOption { type = types.nullOr types.str; description = "user voice key for the front end"; };
              lets-encrypt-hosts = mkOption { type = types.nullOr (types.listOf types.str); description = "hosts to use for lets encrypt certificates"; };
              log = mkOption { type = types.nullOr types.str; description = "path to log file"; };
              cors-domains = mkOption { type = types.nullOr (types.listOf types.str); description = "comma separated list of domains for CORS, setting it enable CORS "; };
              clamav-host = mkOption { type = types.nullOr types.str; description = "host for clamav feature"; };
              rate-limit = mkOption { type = types.nullOr types.int; description = "request per minute"; };
              max-upload-size = mkOption { type = types.nullOr types.int; description = "max upload size in kilobytes  "; };
              purge-days = mkOption { type = types.nullOr types.int; description = "number of days after the uploads are purged automatically "; };
              random-token-length = mkOption { type = types.nullOr types.int; description = "length of the random token for the upload path (double the size for delete path)"; };

            };

            provider_options = {

                aws = {
                  enable = mkEnableOption "Enable AWS backend";
                  aws-access-key = mkOption { type = types.str; description = "aws access key"; };
                  aws-secret-key = mkOption { type = types.str; description = "aws secret key"; };
                  bucket = mkOption { type = types.str; description = "aws bucket "; };
                  s3-endpoint = mkOption {
                    type = types.nullOr types.str;
                    description = ''
                      Custom S3 endpoint. 
                      If you specify the s3-region, you don't need to set the endpoint URL since the correct endpoint will used automatically.
                    '';
                  };
                  s3-region = mkOption { type = types.str; description = "region of the s3 bucket eu-west-"; };
                  s3-no-multipart = mkOption { type = types.nullOr types.bool; description = "disables s3 multipart upload "; };
                  s3-path-style = mkOption { type = types.nullOr types.str; description = "Forces path style URLs, required for Minio. "; };
                };

                storj = {
                  enable = mkEnableOption "Enable storj backend";
                  storj-access = mkOption { type = types.str; description = "Access for the project"; };
                  storj-bucket = mkOption { type = types.str; description = "Bucket to use within the project"; };
                };

                gdrive = {
                  enable = mkEnableOption "Enable gdrive backend";
                  gdrive-client-json = mkOption { type = types.str; description = "oauth client json config for gdrive provider"; };
                  gdrive-chunk-size = mkOption { default = 8; type = types.nullOr types.int; description = "chunk size for gdrive upload in megabytes, must be lower than available memory (8 MB)"; };
                  basedir = mkOption { type = types.str; description = "path storage for gdrive provider"; default = "${cfg.stateDir}/store"; };
                  purge-interval = mkOption { type = types.nullOr types.int; description = "interval in hours to run the automatic purge for (not applicable to S3 and Storj)"; };

                };

                local = {
                  enable = mkEnableOption "Enable local backend";
                  basedir = mkOption { type = types.str; description = "path storage for local provider"; default = "${cfg.stateDir}/store"; };
                  purge-interval = mkOption { type = types.nullOr types.int; description = "interval in hours to run the automatic purge for (not applicable to S3 and Storj)"; };
                };

              };
          in
            {
              options.services.transfer-sh = fold recursiveUpdate {} [
                general_options
                {
                  provider = provider_options;
                  user = mkOption {
                    type = types.str;
                    description = "User to run the service under";
                    default = "transfer.sh";
                  };
                  group = mkOption {
                    type = types.str;
                    description = "Group to run the service under";
                    default = "transfer.sh";
                  };
                  stateDir = mkOption {
                    type = types.path;
                    description = "Variable state directory";
                    default = RUNTIME_DIR;
                  };
                }
              ];

              config = let

                mkFlags = cfg: options:
                  let
                    mkBoolFlag = option: if cfg.${option} then [ "--${option}" ] else [];
                    mkFlag = option:
                      if isBool cfg.${option}
                      then mkBoolFlag option
                      else [ "--${option}" "${cfg.${option}}" ];

                  in
                    lists.flatten (map (mkFlag) (filter (option: cfg.${option} != null && option != "enable") options));

                aws-config = (mkFlags cfg.provider.aws (attrNames provider_options)) ++ [ "--provider" "aws" ];
                gdrive-config = mkFlags cfg.provider.gdrive (attrNames provider_options.gdrive) ++ [ "--provider" "gdrive" ];
                storj-config = mkFlags cfg.provider.storj (attrNames provider_options.storj) ++ [ "--provider" "storj" ];
                local-config = mkFlags cfg.provider.local (attrNames provider_options.local) ++ [ "--provider" "local" ];

                general-config = concatStringsSep " " (mkFlags cfg (attrNames general_options));
                provider-config = concatStringsSep " " (
                  if cfg.provider.aws.enable && !cfg.provider.storj.enable && !cfg.provider.gdrive.enable && !cfg.provider.local.enable then aws-config
                  else if !cfg.provider.aws.enable && cfg.provider.storj.enable && !cfg.provider.gdrive.enable && !cfg.provider.local.enable then storj-config
                  else if !cfg.provider.aws.enable && !cfg.provider.storj.enable && cfg.provider.gdrive.enable && !cfg.provider.local.enable then gdrive-config
                  else if !cfg.provider.aws.enable && !cfg.provider.storj.enable && !cfg.provider.gdrive.enable && cfg.provider.local.enable then local-config
                  else throw "transfer.sh requires exactly one provider (aws, storj, gdrive, local)"
                );

              in
                lib.mkIf cfg.enable
                  {
                    systemd.tmpfiles.rules = [
                      "d ${cfg.stateDir} 0750 ${cfg.user} ${cfg.group} - -"
                    ] ++ optional cfg.provider.gdrive.enable cfg.provider.gdrive.basedir
                    ++ optional cfg.provider.local.enable cfg.provider.local.basedir;

                    systemd.services.transfer-sh = {
                      wantedBy = [ "multi-user.target" ];
                      after = [ "network.target" ];
                      serviceConfig = {
                        User = cfg.user;
                        Group = cfg.group;
                        ExecStart = "${transfer-sh pkgs}/bin/transfer.sh ${general-config} ${provider-config} ";
                      };
                    };

                    networking.firewall.allowedTCPPorts = [ cfg.listener cfg.profile-listener cfg.tls-listener ];
                  };
            };

          default = { self, pkgs, ... }: {
            imports = [ nixosModules.transfer-sh ];
            # Network configuration.

            # useDHCP is generally considered to better be turned off in favor
            # of <adapter>.useDHCP
            networking.useDHCP = false;
            networking.firewall.allowedTCPPorts = [];

            # Enable the inventaire server.
            services.transfer-sh = {
              enable = true;
              provider.local = {
                enable = true;
              };
            };

            nixpkgs.config.allowUnfree = true;
          };
        };


        nixosConfigurations."container" = nixpkgs.lib.nixosSystem {
          system = "x86_64-linux";
          modules = [
            nixosModules.default
            ({ ... }: { boot.isContainer = true; })
          ];
        };

      };
}
