## Routes

```sh
grep -r 'http.Handle(' --no-filename services/ | sed 's/\s\+//'
```

## Development

### Running with automatic recompiling & rerunning

```sh
find -name '*.templ' | entr go generate ./...
```

```sh
find -name "*.go" | entr -r go run .
```

### Mocking an OIDC provider

Start a vault dev server (tested with version v1.16.2):

```sh
vault server -dev
```

In another terminal, set it up with the following:

```sh
export VAULT_ADDR=http://127.0.0.1:8200

vault auth enable userpass
vault auth tune -listing-visibility=unauth userpass
vault write auth/userpass/users/turetek password=turetek
vault write identity/entity name=turetek
vault write identity/entity-alias \
    name=turetek \
    canonical_id=$(vault read -field=id identity/entity/name/turetek) \
    mount_accessor=$(vault auth list -detailed -format json | jq -r '.["userpass/"].accessor')

vault write identity/oidc/scope/profile template=$(echo '{"username":{{identity.entity.name}}}' | base64 -)
vault write identity/oidc/provider/default scopes_supported="profile"
vault write identity/oidc/client/logout redirect_uris="http://localhost:3000/oidc/kth/callback" assignments=allow_all

echo -n "
KTH_ISSUER_URL=$(vault read -field=issuer identity/oidc/provider/default)
KTH_CLIENT_ID=$(vault read -field=client_id identity/oidc/client/logout)
KTH_CLIENT_SECRET=$(vault read -field=client_secret identity/oidc/client/logout)
KTH_RP_ORIGIN=http://localhost:3000
"
```

Then set the env variables printed at the end. Then you can log in using
`turetek` as username and password when you've registered `turetek` in the
system. Note that it doesn't mock all the claims we get from kth.

## Example claims from KTH:

```js
{
    "affiliation": [
        "member",
        "student"
    ],
    "appid": "bad94f41-8323-4c26-8c59-fb6d6b8384db",
    "apptype": "Confidential",
    "aud": "bad94f41-8323-4c26-8c59-fb6d6b8384db",
    "auth_time": 1717086052,
    "authmethod": "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport",
    "email": "turetek@kth.se",
    "exp": 1717090261,
    "iat": 1717086661,
    "iss": "https://login.ug.kth.se/adfs",
    "kthid": "u1jwkms6",
    "nbf": 1717086661,
    "scp": "openid",
    "sid": "S-1-1-11-1111111111-1111111111-1111111111-1111111",
    "sub": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "unique_name": [
        "Ture Teknokrat",
        "UG\\turetek"
    ],
    "upn": "turetek@ug.kth.se",
    "username": "turetek",
    "ver": "1.0"
}
```
