---
title: 'Authentication'
weight: 5
---
# Authentication

GateKeeper support JWT authentication using the [JSON Web Token](https://jwt.io/) (JWT) standard.

## JWT Secret

GateKeeper signs its tokens with the secret in the `GATEKEEPER_JWT_SECRET` environment variable. When the `Web` module is configured and this variable is unset or empty, GateKeeper refuses to start and logs:

```
GATEKEEPER_JWT_SECRET must be set when Web is configured
```

Use a long random value, for example:

```bash
export GATEKEEPER_JWT_SECRET="$(openssl rand -base64 32)"
```

Tokens are valid for 7 days. Changing the secret logs everybody out.

## Configuration

There are two ways to log in, and password login is always available.

### Password Login (htpasswd)

Password login checks users against an [htpasswd](https://httpd.apache.org/docs/2.4/programs/htpasswd.html) file, named by `HTPasswdFile` in the `Web` section:

```yaml
Web:
  Address: :8085
  HTPasswdFile: /path/to/.htpasswd
```

The file is read again on every login, so you can add or change users without a restart.

#### First Start

When `HTPasswdFile` is set and no file exists at that path, GateKeeper creates it on startup with one user, `admin`, a random password (stored as a bcrypt hash, file mode `0600`), and logs the password once:

```
WARNING - Created htpasswd file /path/to/.htpasswd; log in as admin with password: <password>
```

The password stays valid across restarts. Change it with `htpasswd -B /path/to/.htpasswd admin`.

#### Fallback Password

If GateKeeper still cannot read an htpasswd file when it starts (the key is not set, or the file is unreadable), it makes a random password and logs it once, at startup:

```
WARNING - No htpasswd file; log in as admin with password: <password>
```

Log in as `admin` with that password. It stays the same until GateKeeper restarts. Create an htpasswd file for anything beyond a first look.

Passwords are never written to the log; failed and successful logins are logged with the username only.


### OIDC: OpenID Connect

GateKeeper supports OpenID Connect (OIDC) authentication using the OpenID Connect Provider (OP) specification.

#### Example Configuration

```yaml
Auth:
  AuthType: oidc
  IssuerURL: https://accounts.google.com
  ClientID: your-client-id
  ClientSecretVar: GOOGLE_CLIENT_SECRET
  RedirectURL: https://your-domain.com/auth/callback
  Scopes:
    - openid
    - profile
    - email
```

Note that the callback slug is always `auth/callback`, so this MUST be present in the `RedirectURL` field.

You will also need to configure the `WebURL` in the Web section of your config:

```yaml
Web:
  Address: :8085
  WebURL: https://your-domain.com
```

After an OIDC login GateKeeper sets the `oauth_token` cookie. It is `HttpOnly` (scripts cannot read it), `SameSite=Lax`, lasts as long as the token (7 days), and is `Secure` when `WebURL` starts with `https://`, so serve GateKeeper over HTTPS when you use OIDC.

#### OIDC Parameters

| Key               | Description                                                                                                    | Default |
| ----------------- | ------------------------------------------------------------------------------------------------------------- | ------- |
| AuthType          | Must be set to `oidc`                                                                                         |         |
| IssuerURL         | The URL of your OpenID Connect provider                                                                      |         |
| ClientID          | The client ID from your OIDC provider                                                                        |         |
| ClientSecretVar   | Environment variable name containing the client secret                                                       |         |
| RedirectURL       | The callback URL (must include `/auth/callback`)                                                             |         |
| Scopes            | OAuth scopes to request (typically `openid`, `profile`, `email`)                                            | [openid, profile, email] |
| RoleClaim         | The ID token claim that lists the user's groups                                                              | groups |
| Roles             | Map from a group in `RoleClaim` to a GateKeeper role (`admin` or `viewer`)                                  | the groups `admin` and `viewer` map to themselves |

## Roles

Every user has one of two roles:

- `admin` can do everything.
- `viewer` can open every page but cannot change anything: the API answers `403` to any change, and the UI hides the edit buttons.

Password (htpasswd) users are always `admin`. An OIDC user gets the role their groups map to through `Roles`; when a user is in groups for both roles, `admin` wins. **An OIDC user whose groups map to no role cannot log in.**

```yaml
Auth:
  AuthType: oidc
  # ...
  RoleClaim: groups
  Roles:
    gatekeeper-admins: admin
    gatekeeper-viewers: viewer
```

With Authentik, the `groups` claim is sent by the default `profile` scope mapping: create the two groups and add users to them.
