---
title: 'Authentication'
weight: 5
---
# Authentication

GateKeeper support JWT authentication using the [JSON Web Token](https://jwt.io/) (JWT) standard.

## Configuration

GateKeeper requires some form of authentication to be configured. In the event that you do not, a default authentication provider is used with a random password that will be printed to the logs.

There are two types of authentication providers that can be configured:

### HTPasswd

HTPasswd is a simple authentication provider that uses a [htpasswd](https://httpd.apache.org/docs/2.4/programs/htpasswd.html) file to authenticate users. 

#### Example Configuration

```yaml
Auth:
  AuthType: htpasswd
  HTPasswdFile: /path/to/htpasswd
```


### OIDC: OpenID Connect

GateKeeper supports OpenID Connect (OIDC) authentication using the OpenID Connect Provider (OP) specification.

#### Example Configuration

```yaml
Auth:
  AuthType: oidc
  IssuerURL: https://accounts.google.com
  ClientID: your-client-id
  ClientSecretVar: GOOGLE_CLIENT_SECRET
  RedirectURL: http://localhost:5173/auth/callback
  Scopes:
    - openid
    - profile
    - email
```

Note that the callback slug is always `auth/callback`, so this MUST be present in the `RedirectURL` field.
