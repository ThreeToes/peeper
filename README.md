# peeper - a transparent HTTP proxy

## Building
Run `make`

## Configuring
Configuration is done via a TOML configuration file. See 
[example.toml](./config/example.toml) for a basic example

### Network
Network configration is relatively simple, with the only options
being `bind_interface` and `bind_port`

```toml
[network]
bind_interface = "0.0.0.0"
bind_port = 9090
```

### Endpoints
Endpoints are the basic configuration unit of peeper. One endpoint can
be forwarded to a single remote host, for example

```toml
[endpoints]
[endpoints.cats]
local_path = "/cats"
target_path = "https://cat-fact.herokuapp.com/facts"
local_method = "GET"
remote_method = "GET"
```

This will register the local `GET` endpoint `/cats` to forward to 
the [cat facts API](https://alexwohlbruck.github.io/cat-facts/docs/)

#### Authentication
Authentication is configured as part of an endpoint

##### HTTP Basic Auth
Basic authentication requires a username and password
```toml
[endpoints]
[endpoints.cats]
# ...
[endpoints.cats.basic_auth]
username = "bigboss"
password = "5n@ke3a7eR"
```

##### Static Key Authentication
If the remote service requires a keys in a header, you can configure
it with arbitrary key value pairs in the `headers` value
```toml

[endpoints]
[endpoints.cats]
# ...
[endpoints.cats.static_key]
headers = { x-api-key = "some key", x-client-id = "some client ID"}
```

##### OAuth2 Client Credentials Authentication
The [client credential flow](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/)
is supported, and uses Basic auth and a form encoded request body to get
the token. Any other required values (such as `audience` when using Auth0)
can be set as a dictionary in the `extra_form_values` value

```toml
[endpoints]
[endpoints.auth0]
local_path = "/oauth"
target_path = "https://testapi.com/api/banana_farms"
local_method = "GET"
remote_method = "GET"
[endpoints.auth0.oauth]
client_id = "client-abc"
client_secret = "secret"
token_endpoint = "https://testauth0provider.au.auth0.com/oauth/token"
extra_form_values = {audience = "https://testapi.com/api/"}
```

The token request is carried out with every request.