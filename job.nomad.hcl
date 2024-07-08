job "logout" {
  type = "service"
  namespace = "auth"

  group "logout" {
    network {
      port "http" { }
    }

    service {
      name     = "logout"
      port     = "http"
      provider = "nomad"
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.logout.rule=Host(`logout.datasektionen.se`)",
        "traefik.http.routers.logout.tls.certresolver=default",

        "traefik.http.routers.logout-login2.rule=Host(`login2.datasektionen.se`)",
        "traefik.http.routers.logout-login2.tls.certresolver=default",
        "traefik.http.routers.logout-login2.middlewares=redirect-to-logout",
        "traefik.http.middlewares.redirect-to-logout.redirectregex.regex=^https://login2.datasektionen.se/(.*)$",
        "traefik.http.middlewares.redirect-to-logout.redirectregex.replacement=https://logout.datasektionen.se/$${1}",

        "traefik-internal.enable=true",
        "traefik-internal.http.routers.logout.rule=Host(`logout.nomad.dsekt.internal`)",
      ]
    }

    task "logout" {
      driver = "docker"

      config {
        image = var.image_tag
        ports = ["http"]
      }

      template {
        data        = <<ENV
PORT={{ env "NOMAD_PORT_http" }}

{{ with nomadVar "nomad/jobs/logout" }}
KTH_CLIENT_SECRET={{ .kth_client_secret }}
OIDC_PROVIDER_KEY={{ .oidc_provider_key }}
DATABASE_URL=postgresql://logout:logout@localhost:5432/logout
DATABASE_URL=postgresql://logout:{{ .database_password }}@postgres.dsekt.internal:5432/logout
{{ end }}

KTH_ISSUER_URL=https://login.ug.kth.se/adfs
KTH_CLIENT_ID=bad94f41-8323-4c26-8c59-fb6d6b8384db
KTH_RP_ORIGIN=https://login2.datasektionen.se

OIDC_PROVIDER_ISSUER_URL=https://logout.datasektionen.se/op

ORIGIN=https://logout.datasektionen.se
DEV=false
# LDAP_PROXY_URL=TODO
PLS_URL=https://pls.datasektionen.se
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
  default = "ghcr.io/datasektionen/logout:latest"
}
