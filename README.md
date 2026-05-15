# Traefik Request ID Middleware

A Traefik middleware plugin that adds or propagates an `X-Request-ID` header.

If the incoming request already contains the configured request ID header, the plugin preserves it and also adds the same value to the response headers.

If the incoming request does not contain the header, the plugin generates a new UUID v7, injects it into the request, and returns the same value in the response.

## Why use this plugin?

In distributed systems, reverse proxies, APIs, workers, and logs need a common identifier to correlate what happened during a request.

This plugin helps you:

- correlate Traefik access logs with application logs;
- propagate request IDs to backend services;
- return the request ID to clients for debugging;
- standardize request tracing across services;
- avoid generating different IDs in every layer.

## Behavior

### When the request already has an ID

Request:

```http
GET / HTTP/1.1
Host: app.example.com
X-Request-ID: client-123
````

The plugin keeps the existing value.

Forwarded request to backend:

```http
X-Request-ID: client-123
```

Response to client:

```http
X-Request-ID: client-123
```

### When the request does not have an ID

Request:

```http
GET / HTTP/1.1
Host: app.example.com
```

The plugin generates a UUID v7.

Forwarded request to backend:

```http
X-Request-ID: 018f8f2e-7c3b-7cc2-9f1a-2d0f3c6d9e21
```

Response to client:

```http
X-Request-ID: 018f8f2e-7c3b-7cc2-9f1a-2d0f3c6d9e21
```

## Features

* Adds `X-Request-ID` to incoming requests when missing.
* Preserves existing request IDs.
* Adds the same request ID to response headers.
* Supports custom header names.
* Uses UUID v7 for generated IDs.
* Works as a Traefik HTTP middleware plugin.

## Configuration

### Static configuration

Add the plugin to your Traefik static configuration.

```yaml
experimental:
  plugins:
    requestid:
      moduleName: github.com/jrCleber/traefik-request-id
      version: v0.1.0
```

For local development:

```yaml
experimental:
  localPlugins:
    requestid:
      moduleName: github.com/jrCleber/traefik-request-id
```

## Dynamic configuration

### File provider

```yaml
http:
  middlewares:
    request-id:
      plugin:
        requestid:
          headerName: X-Request-ID
```

Then attach the middleware to a router:

```yaml
http:
  routers:
    app:
      rule: Host(`app.example.com`)
      entryPoints:
        - websecure
      service: app-service
      middlewares:
        - request-id

  services:
    app-service:
      loadBalancer:
        servers:
          - url: http://app:3000
```

### Docker provider

```yaml
services:
  app:
    image: your-app:latest
    labels:
      - traefik.enable=true

      - traefik.http.routers.app.rule=Host(`app.example.com`)
      - traefik.http.routers.app.entrypoints=websecure
      - traefik.http.routers.app.service=app-service
      - traefik.http.routers.app.middlewares=request-id

      - traefik.http.services.app-service.loadbalancer.server.port=3000

      - traefik.http.middlewares.request-id.plugin.requestid.headerName=X-Request-ID
```

## Options

| Option       | Type   | Default        | Description                                                  |
| ------------ | ------ | -------------- | ------------------------------------------------------------ |
| `headerName` | string | `X-Request-ID` | Header name used to read, inject, and return the request ID. |

## Example with custom header

```yaml
http:
  middlewares:
    request-id:
      plugin:
        requestid:
          headerName: X-Correlation-ID
```

Request without header:

```bash
curl -i https://app.example.com
```

Expected response:

```http
X-Correlation-ID: 018f8f2e-7c3b-7cc2-9f1a-2d0f3c6d9e21
```

## Testing

### Request without `X-Request-ID`

```bash
curl -i https://app.example.com
```

Expected:

```http
X-Request-ID: <generated-uuid-v7>
```

### Request with `X-Request-ID`

```bash
curl -i \
  -H "X-Request-ID: test-123" \
  https://app.example.com
```

Expected:

```http
X-Request-ID: test-123
```

## Local development

Example folder structure:

```txt
plugins-local/
└── src/
    └── github.com/
        └── jrCleber/
            └── traefik-request-id/
                ├── go.mod
                ├── go.sum
                ├── requestid.go
                └── .traefik.yml
```

Traefik static configuration:

```yaml
experimental:
  localPlugins:
    requestid:
      moduleName: github.com/code-chat-br/traefik-request-id
```

Docker Compose example:

```yaml
services:
  traefik:
    image: traefik:v3.6.7
    command:
      - --configFile=/etc/traefik/traefik.yml
    volumes:
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./plugins-local:/plugins-local
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - public

networks:
  public:
    external: true
```

## Plugin metadata

Example `.traefik.yml`:

```yaml
displayName: Request ID Middleware
type: middleware

import: github.com/jrCleber/traefik-request-id

summary: Adds or propagates X-Request-ID in request and response headers.

testData:
  headerName: X-Request-ID
```

## Recommended logging usage

To get the full benefit of this plugin, your backend application should read the request ID header and include it in logs.

Example in Node.js / Express:

```js
app.use((req, res, next) => {
  const requestId = req.headers['x-request-id']

  req.requestId = requestId
  res.setHeader('X-Request-ID', requestId)

  next()
})
```

Example log:

```json
{
  "level": "info",
  "request_id": "018f8f2e-7c3b-7cc2-9f1a-2d0f3c6d9e21",
  "message": "request processed"
}
```

## Request ID vs distributed tracing

This plugin is useful for request correlation.

It is not a full distributed tracing solution.

For complete tracing across services, use OpenTelemetry and the `traceparent` header.

Recommended usage:

* Use `X-Request-ID` for simple request correlation.
* Use `traceparent` for distributed tracing.
* Use both when you need simple debugging plus full observability.

## Security considerations

This plugin preserves an incoming request ID if the client sends one.

That is useful when clients or upstream gateways already generate request IDs.

However, if you do not trust external clients, be careful when using client-provided IDs in logs or dashboards.

Do not treat `X-Request-ID` as an authentication, authorization, or security boundary.

## Compatibility

This plugin is designed for Traefik HTTP routers and middlewares.

It does not apply to TCP or UDP routers.

## License

MIT