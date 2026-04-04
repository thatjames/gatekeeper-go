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

#### HTPasswd Parameters

| Key          | Description                                              | Default |
| ------------ | -------------------------------------------------------- | ------- |
| AuthType     | Must be set to `htpasswd`                               |         |
| HTPasswdFile | Path to the htpasswd file                               |         |


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

#### OIDC Parameters

| Key               | Description                                                                                                    | Default |
| ----------------- | ------------------------------------------------------------------------------------------------------------- | ------- |
| AuthType          | Must be set to `oidc`                                                                                         |         |
| IssuerURL         | The URL of your OpenID Connect provider                                                                      |         |
| ClientID          | The client ID from your OIDC provider                                                                        |         |
| ClientSecretVar   | Environment variable name containing the client secret                                                       |         |
| RedirectURL       | The callback URL (must include `/auth/callback`)                                                             |         |
| Scopes            | OAuth scopes to request (typically `openid`, `profile`, `email`)                                            | [openid, profile, email] |
