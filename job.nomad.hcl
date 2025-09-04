job "sso" {
  type = "service"
  namespace = "auth"

  group "sso" {
    network {
      port "http_external" { }
      port "http_internal" { }
    }

    service {
      name     = "sso"
      port     = "http_external"
      provider = "nomad"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.sso.rule=Host(`sso.datasektionen.se`)",
        "traefik.http.routers.sso.tls.certresolver=default",

        "traefik.http.routers.sso-login2.rule=Host(`login2.datasektionen.se`)",
        "traefik.http.routers.sso-login2.tls.certresolver=default",
        "traefik.http.routers.sso-login2.middlewares=redirect-to-sso",
        "traefik.http.middlewares.redirect-to-sso.redirectregex.regex=^https://[^.]*.datasektionen.se/(.*)$",
        "traefik.http.middlewares.redirect-to-sso.redirectregex.replacement=https://sso.datasektionen.se/$${1}",
      ]
    }

    service {
      name     = "sso-internal"
      port     = "http_internal"
      provider = "nomad"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.sso-internal.rule=Host(`sso.nomad.dsekt.internal`)",
        "traefik.http.routers.sso-internal.entrypoints=web-internal",
      ]
    }

    task "sso" {
      driver = "docker"

      config {
        image = var.image_tag
        ports = ["http_external", "http_internal"]
      }

      template {
        data        = <<ENV
PORT_EXTERNAL={{ env "NOMAD_PORT_http_external" }}
PORT={{ env "NOMAD_PORT_http_internal" }}

{{ with nomadVar "nomad/jobs/sso" }}
KTH_CLIENT_SECRET={{ .kth_client_secret }}
OIDC_PROVIDER_KEY={{ .oidc_provider_key }}
DATABASE_URL=postgresql://sso:{{ .database_password }}@postgres.dsekt.internal:5432/sso
SPAM_API_KEY={{ .spam_api_key }}
HIVE_API_KEY={{ .hive_api_key }}
{{ end }}

KTH_ISSUER_URL=https://login.ug.kth.se/adfs
KTH_CLIENT_ID=bad94f41-8323-4c26-8c59-fb6d6b8384db
KTH_RP_ORIGIN=https://login2.datasektionen.se

OIDC_PROVIDER_ISSUER_URL=https://sso.datasektionen.se/op

ORIGIN=https://sso.datasektionen.se
DEV=false
LDAP_PROXY_URL=http://ldap-proxy.dsekt.internal:38980
PLS_URL=https://pls.datasektionen.se
HIVE_URL=http://hive.nomad.dsekt.internal
SPAM_URL=https://spam.datasektionen.se
ENV
        destination = "local/.env"
        env         = true
      }

      resources {
        memory = 40
      }
    }
  }
}

variable "image_tag" {
  type = string
  default = "ghcr.io/datasektionen/sso:latest"
}
