{ lib, config, pkgs, inputs, ... }:
let cfg = config.services.pots;
in {
  options = with lib; {
    services.pots = {
      enable = lib.mkEnableOption "Enable pots";

      port = mkOption {
        type = types.int;
        default = 8888;
        description = ''
          Port to listen on string.
        '';
      };

      user = mkOption {
        type = with types; oneOf [ str int ];
        default = "pots";
        description = ''
          The user the service will use.
        '';
      };

      group = mkOption {
        type = with types; oneOf [ str int ];
        default = "pots";
        description = ''
          The group the service will use.
        '';
      };

      keyPath = mkOption {
        type = types.path;
        default = "";
        description = ''
          Path to the GitHub API key file
        '';
      };

      dataDir = mkOption {
        type = types.path;
        default = "/var/lib/pots";
        description = "Path pots home directory";
      };

      package = mkOption {
        type = types.package;
        default = pkgs.pots;
        defaultText = literalExpression "pkgs.pots";
        description = "The package to use for pots";
      };

      envFile = mkOption {
        type = types.path;
        default = "/run/secrets/pots_env_file";
        description = ''
          Path to a file containing the pots token information
        '';
      };
    };
  };

  config = lib.mkIf (cfg.enable) {
    users.groups.${cfg.group} = { };
    users.users.${cfg.user} = {
      description = "pots service user";
      isSystemUser = true;
      home = "${cfg.dataDir}";
      createHome = true;
      group = "${cfg.group}";
    };

    systemd.services.pots = {
      enable = true;
      description = "pots server";
      after = [ "network-online.target" ];
      wantedBy = [ "multi-user.target" ];
      wants = [ "network-online.target" ];
      environment = { HOME = "${cfg.dataDir}"; };

      serviceConfig = {
        User = cfg.user;
        Group = cfg.group;

        ExecStart =
          "${cfg.package}/bin/pots -port ${toString cfg.port}";
        EnvironmentFile = cfg.envFile;
      };
    };
  };
}
