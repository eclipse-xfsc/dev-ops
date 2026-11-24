# XFSC Identity Tenant Management

The XFSC Identity Tenant Management component provides a multi-tenant identity management solution based on Keycloak and NATS.

It bundles:

- A customized Keycloak theme.
- A global Keycloak event listener publishing user lifecycle events as CloudEvents.
- An optional NATS-to-Keycloak Admin API controller.
- Automated tenant and OpenID Connect client management.
- Docker Compose and Kubernetes Helm deployments.

## Architecture

```text
                         NATS
                           |
          +----------------+----------------+
          |                                 |
  Management Commands              User Lifecycle Events
          |                                 ^
          v                                 |
  Identity Controller               XFSC Event Listener
          |                                 |
          | Keycloak Admin API              |
          v                                 |
       Keycloak ---------------------------+
          |
          v
      PostgreSQL
```

Each tenant is represented by a dedicated Keycloak realm. Applications are represented by OpenID Connect clients within the corresponding realm.

The identity controller receives management commands through NATS and translates them into Keycloak Admin API operations.

## Prerequisites

- Docker and Docker Compose
- NATS Server
- NATS CLI
- PostgreSQL
- Helm and kubectl for Kubernetes deployments
- jq (optional, for formatting JSON responses)

### Install the NATS CLI

macOS:

```sh
brew install nats-io/nats-tools/nats
```

Linux or macOS with Go installed:

```sh
go install github.com/nats-io/natscli/nats@latest
```

Alternatively, download a binary from the official [NATS CLI releases](https://github.com/nats-io/natscli/releases).

Verify the installation:

```sh
nats --version
```

## Build

From `components/keycloak`:

### Build Keycloak

```sh
docker build --no-cache \
  -f deployment/docker/Dockerfile \
  -t xfsc-keycloak .
```

### Build the identity controller

```sh
docker build --no-cache \
  -f controller/deployment/docker/Dockerfile \
  -t xfsc-identity-controller \
  controller
```

The controller uses the included snapshot of `nats-message-library` to ensure compatibility with the XFSC wire contracts.

Replace this snapshot with a pinned released module when available.

## Docker Compose

### Start the environment

```sh
docker compose up -d
```

### Rebuild the complete environment

Stop and remove existing containers:

```sh
docker compose down --remove-orphans
```

Rebuild all locally built images:

```sh
docker compose build --no-cache --pull
```

Pull externally maintained images:

```sh
docker compose pull --ignore-buildable
```

Recreate and start all containers:

```sh
docker compose up -d --force-recreate
```

Verify the deployment:

```sh
docker compose ps
```

Inspect logs:

```sh
docker compose logs -f
```

## NATS Configuration

Configure the NATS server address:

```sh
export NATS_URL="nats://localhost:4222"
```

Create a reusable NATS CLI context:

```sh
nats context save xfsc \
  --server "$NATS_URL"
```

Select the context:

```sh
nats context select xfsc
```

Inspect the active context:

```sh
nats context info
```

For authenticated environments, configure the context with the appropriate credentials and TLS settings.

For example:

```sh
nats context save xfsc-production \
  --server "tls://nats.example.com:4222" \
  --creds ./identity.creds
```

## NATS Management API

The identity controller subscribes to the following management subjects.

| Subject | Operation |
|---|---|
| `tenant.management.identity.create` | Create a tenant realm |
| `tenant.management.identity.update` | Update a tenant realm |
| `tenant.management.identity.delete` | Delete a tenant realm |
| `tenant.management.identity.app.create` | Create an OIDC application |
| `tenant.management.identity.app.update` | Update an OIDC application |
| `tenant.management.identity.app.delete` | Delete an OIDC application |

### Request format

Management requests use the fields defined by `common.Request`.

| Field | Required | Description |
|---|---|---|
| `tenant_id` | Yes | Tenant identifier |
| `request_id` | Yes | Request correlation identifier |
| `group_id` | No | Optional group identifier |

Additional fields depend on the requested operation.

The `request_id` is used for correlation. It is not currently an idempotency key.

### Request-reply communication

Use NATS request-reply communication to receive the operation result:

```sh
nats request \
  tenant.management.identity.create \
  '{"tenant_id":"demo","request_id":"req-1","enabled":true}' \
  --timeout=30s
```

The NATS CLI automatically creates a reply inbox.

The controller returns a response based on `common.Reply`. Application management responses additionally contain `appId`.

A successful response contains:

```json
{
  "error": null
}
```

This example shows the relevant success field, not necessarily the complete response envelope.

### Publish-only communication

Commands can also be published without waiting for a response:

```sh
nats pub \
  tenant.management.identity.create \
  '{"tenant_id":"demo","request_id":"req-1","enabled":true}'
```

However, request-reply communication is recommended for provisioning operations because publish-only communication does not confirm whether the operation succeeded.

## Tenant Management

### Create a tenant

Create a dedicated Keycloak realm and enable self-service registration.

```sh
nats request \
  tenant.management.identity.create \
  '{
    "tenant_id": "demo",
    "request_id": "req-tenant-create-1",
    "enabled": true,
    "registrationAllowed": true
  }' \
  --timeout=30s
```

### Update a tenant

Disable self-service registration while keeping the tenant enabled.

```sh
nats request \
  tenant.management.identity.update \
  '{
    "tenant_id": "demo",
    "request_id": "req-tenant-update-1",
    "enabled": true,
    "registrationAllowed": false
  }' \
  --timeout=30s
```

### Disable a tenant

```sh
nats request \
  tenant.management.identity.update \
  '{
    "tenant_id": "demo",
    "request_id": "req-tenant-disable-1",
    "enabled": false
  }' \
  --timeout=30s
```

### Enable a tenant

```sh
nats request \
  tenant.management.identity.update \
  '{
    "tenant_id": "demo",
    "request_id": "req-tenant-enable-1",
    "enabled": true
  }' \
  --timeout=30s
```

### Delete a tenant

**Warning:** Deleting a tenant realm is destructive and may permanently remove its users, clients, roles, and associated configuration.

```sh
nats request \
  tenant.management.identity.delete \
  '{
    "tenant_id": "demo",
    "request_id": "req-tenant-delete-1"
  }' \
  --timeout=30s
```

## Application Management

Applications are represented by OpenID Connect clients within a tenant realm.

The tenant must exist before creating an application.

### Create a public authorization-code application

Create a public OpenID Connect client using Authorization Code Flow with PKCE S256.

```sh
nats request \
  tenant.management.identity.app.create \
  '{
    "tenant_id": "demo",
    "request_id": "req-app-create-1",
    "appId": "portal",
    "name": "Portal",
    "client": {
      "type": "public"
    },
    "authorization": {
      "grantTypes": [
        "authorization_code"
      ],
      "redirectUris": [
        "https://portal.example.com/callback"
      ],
      "pkceRequired": true,
      "pkceMethod": "S256"
    }
  }' \
  --timeout=30s
```

Public clients require PKCE S256.

### Create a confidential authorization-code application

```sh
nats request \
  tenant.management.identity.app.create \
  '{
    "tenant_id": "demo",
    "request_id": "req-app-create-2",
    "appId": "backend-portal",
    "name": "Backend Portal",
    "client": {
      "type": "confidential"
    },
    "authorization": {
      "grantTypes": [
        "authorization_code"
      ],
      "redirectUris": [
        "https://backend.example.com/callback"
      ]
    }
  }' \
  --timeout=30s
```

Confidential clients require credential lifecycle management outside this API.

Client secrets must not be distributed through unrestricted NATS subjects or application logs.

### Update an application

Update the application name and redirect URIs.

```sh
nats request \
  tenant.management.identity.app.update \
  '{
    "tenant_id": "demo",
    "request_id": "req-app-update-1",
    "appId": "portal",
    "name": "XFSC Portal",
    "authorization": {
      "grantTypes": [
        "authorization_code"
      ],
      "redirectUris": [
        "https://portal.example.com/callback",
        "https://portal.example.com/auth/callback"
      ],
      "pkceRequired": true,
      "pkceMethod": "S256"
    }
  }' \
  --timeout=30s
```

Updates are partial. Removing existing PKCE attributes is not currently supported.

### Delete an application

```sh
nats request \
  tenant.management.identity.app.delete \
  '{
    "tenant_id": "demo",
    "request_id": "req-app-delete-1",
    "appId": "portal"
  }' \
  --timeout=30s
```

Deleting an application removes the corresponding client configuration from the tenant realm.

## Using JSON Payload Files

For larger requests, store the payload in a JSON file.

Create `create-tenant.json`:

```json
{
  "tenant_id": "demo",
  "request_id": "req-tenant-create-1",
  "enabled": true,
  "registrationAllowed": true
}
```

Send the request:

```sh
nats request \
  tenant.management.identity.create \
  "$(cat create-tenant.json)" \
  --timeout=30s
```

Alternatively, generate requests dynamically using `jq`.

```sh
TENANT_ID="demo"
REQUEST_ID="req-$(date +%s)-$$"

PAYLOAD="$(jq -n \
  --arg tenant "$TENANT_ID" \
  --arg request "$REQUEST_ID" \
  '{
    tenant_id: $tenant,
    request_id: $request,
    enabled: true,
    registrationAllowed: true
  }')"

nats request \
  tenant.management.identity.create \
  "$PAYLOAD" \
  --timeout=30s
```

These examples work on Linux and macOS.

## Self-Service User Events

The built-in global Keycloak event listener publishes user lifecycle events to NATS.

Events are encoded as CloudEvents 1.0 JSON.

| NATS subject | CloudEvent type |
|---|---|
| `tenant.management.identity.user.created` | `org.eclipse.xfsc.identity.user.created` |
| `tenant.management.identity.user.deleted` | `org.eclipse.xfsc.identity.user.deleted` |

Events are published after the corresponding Keycloak transaction commits.

Core NATS does not guarantee durable delivery.

### Subscribe to user creation events

```sh
nats sub \
  tenant.management.identity.user.created
```

### Subscribe to user deletion events

```sh
nats sub \
  tenant.management.identity.user.deleted
```

### Subscribe to all user lifecycle events

```sh
nats sub \
  'tenant.management.identity.user.>'
```

### Subscribe to all identity management traffic

```sh
nats sub \
  'tenant.management.identity.>'
```

This subscription includes management commands and user lifecycle events.

### Display JSON events

```sh
nats sub \
  --raw \
  'tenant.management.identity.user.>' \
  | jq --unbuffered .
```

### Publish a synthetic CloudEvent

The following example can be used to test downstream CloudEvent consumers without creating an actual Keycloak user.

```sh
nats pub \
  tenant.management.identity.user.created \
  "{
    \"specversion\": \"1.0\",
    \"id\": \"test-$(date +%s)-$$\",
    \"source\": \"urn:xfsc:identity:test\",
    \"type\": \"org.eclipse.xfsc.identity.user.created\",
    \"time\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"datacontenttype\": \"application/json\",
    \"data\": {
      \"tenant_id\": \"demo\",
      \"user_id\": \"test-user\"
    }
  }"
```

This is a synthetic test event. The `data` object is illustrative and must be aligned with the actual event listener contract when testing consumers.

Publishing this event does not create a user in Keycloak.

## End-to-End Testing

### 1. Monitor identity events

In a separate terminal:

```sh
nats sub 'tenant.management.identity.>'
```

### 2. Create a tenant

```sh
nats request \
  tenant.management.identity.create \
  '{
    "tenant_id": "demo",
    "request_id": "e2e-tenant-1",
    "enabled": true,
    "registrationAllowed": true
  }' \
  --timeout=30s
```

### 3. Verify the tenant

Configure the public Keycloak URL:

```sh
export KEYCLOAK_URL="https://auth.example.com"
```

Retrieve the OpenID Connect discovery document:

```sh
curl -fsS \
  "$KEYCLOAK_URL/realms/demo/.well-known/openid-configuration" \
  | jq .
```

Inspect the relevant endpoints:

```sh
curl -fsS \
  "$KEYCLOAK_URL/realms/demo/.well-known/openid-configuration" \
  | jq '{
      issuer,
      authorization_endpoint,
      token_endpoint,
      jwks_uri
    }'
```

### 4. Create an application

```sh
nats request \
  tenant.management.identity.app.create \
  '{
    "tenant_id": "demo",
    "request_id": "e2e-app-1",
    "appId": "portal",
    "name": "XFSC Portal",
    "client": {
      "type": "public"
    },
    "authorization": {
      "grantTypes": [
        "authorization_code"
      ],
      "redirectUris": [
        "https://portal.example.com/callback"
      ],
      "pkceRequired": true,
      "pkceMethod": "S256"
    }
  }' \
  --timeout=30s
```

### 5. Test user registration

Open the tenant account console:

```text
https://auth.example.com/realms/demo/account/
```

Register a new user using the self-service registration flow.

Verify that the NATS subscriber receives a user creation event.

### 6. Clean up

Only execute these commands against disposable test tenants.

```sh
nats request \
  tenant.management.identity.app.delete \
  '{
    "tenant_id": "demo",
    "request_id": "e2e-cleanup-app-1",
    "appId": "portal"
  }' \
  --timeout=30s
```

```sh
nats request \
  tenant.management.identity.delete \
  '{
    "tenant_id": "demo",
    "request_id": "e2e-cleanup-tenant-1"
  }' \
  --timeout=30s
```

## Kubernetes Deployment

### Helm installation

Deploy Keycloak and the identity controller:

```sh
helm upgrade --install identity \
  deployment/helm/xfsc-keycloak \
  --namespace identity \
  --create-namespace \
  --values values.yaml
```

The identity controller is enabled by default.

Configure the following deployment parameters:

- Keycloak and controller container images.
- PostgreSQL connection settings.
- Database credentials.
- Keycloak bootstrap administrator credentials.
- Identity controller service-account credentials.
- NATS connection settings.
- Public Keycloak hostname.
- Operator injection settings, where applicable.

### Reverse proxy configuration

When TLS is terminated at a Kubernetes Gateway, Keycloak can communicate with the gateway over HTTP.

Typical Keycloak environment variables:

```yaml
- name: KC_HOSTNAME
  value: "https://auth.example.com"

- name: KC_HTTP_ENABLED
  value: "true"

- name: KC_PROXY_HEADERS
  value: "xforwarded"
```

Only trust forwarded headers from controlled reverse proxies.

The HTTPRoute must forward all required Keycloak paths, including `/realms`, `/resources`, and `/admin` where administrative access is intended.

### PostgreSQL configuration

When database connection parameters are injected individually, Keycloak can construct the JDBC URL using:

```yaml
- name: KC_DB
  value: "postgres"

- name: KC_DB_URL_HOST
  value: "postgres"

- name: KC_DB_URL_PORT
  value: "5432"

- name: KC_DB_URL_DATABASE
  value: "keycloak"
```

Alternatively, configure a complete JDBC URL:

```yaml
- name: KC_DB_URL
  value: "jdbc:postgresql://postgres:5432/keycloak"
```

Do not assign a plain PostgreSQL hostname to `KC_DB_URL`.

### Controller authentication

The identity controller uses the OAuth 2.0 Client Credentials Grant to authenticate against the Keycloak Admin API.

The controller requires a dedicated confidential client with service accounts enabled and sufficient permissions for the intended realm and client management operations.

Do not use bootstrap administrator credentials in the controller.

The service-account client must be provisioned before the controller can perform management operations.

Cross-realm provisioning may require elevated permissions. Apply least-privilege principles wherever possible.

## Accessing NATS in Kubernetes

If NATS is not exposed externally, use Kubernetes port-forwarding.

Find the NATS service:

```sh
kubectl get svc -A | grep nats
```

Forward the NATS client port:

```sh
kubectl port-forward \
  -n infrastructure \
  svc/nats \
  4222:4222
```

Configure the local NATS CLI:

```sh
export NATS_URL="nats://localhost:4222"

nats context save xfsc-kubernetes \
  --server "$NATS_URL"

nats context select xfsc-kubernetes
```

Send management commands:

```sh
nats request \
  tenant.management.identity.create \
  '{
    "tenant_id": "kubernetes-demo",
    "request_id": "k8s-tenant-1",
    "enabled": true
  }' \
  --timeout=30s
```

Adjust the NATS service name, namespace, authentication and TLS settings to match the actual deployment.

## Troubleshooting

### NATS: No responders available

```text
nats: no responders available for request
```

The NATS server has no matching active responder for the requested subject.

Verify:

- The identity controller is running.
- The controller is connected to the correct NATS server.
- The controller subscribed to the expected management subjects.
- NATS account and subject permissions are correct.

Inspect Kubernetes workloads:

```sh
kubectl get pods -n identity
kubectl get deployments -n identity
```

Inspect controller logs using the actual deployment name:

```sh
kubectl logs \
  -n identity \
  deployment/<controller-deployment> \
  --tail=100
```

### NATS request timeout

A request timeout does not necessarily mean that the operation failed.

The controller may have received the request but failed to return a response within the configured timeout.

Check the controller logs and the resulting Keycloak state before retrying a management command.

Requests are not currently guaranteed to be idempotent.

### Keycloak Admin API: HTTP 401

Verify:

- Client ID and client secret.
- Service-account configuration.
- Authentication realm.
- Token endpoint configuration.

### Keycloak Admin API: HTTP 403

The controller authenticated successfully but does not have sufficient permissions.

Verify the service-account role mappings and cross-realm management permissions.

### Keycloak cannot connect to PostgreSQL

An error such as:

```text
Driver does not support the provided URL
```

usually indicates an invalid JDBC connection string.

Verify that `KC_DB_URL` contains a complete JDBC URL or use the individual database connection parameters.

### Keycloak cannot connect to NATS

An error such as:

```text
Unable to connect to NATS servers: [nats://localhost:4222]
```

may indicate that the configured NATS address is missing or incorrect.

Inside Kubernetes, `localhost` refers to the current pod, not the NATS service.

Verify the NATS configuration expected by the event listener and the environment variables injected into the Keycloak container.

### User lifecycle events are missing

Start a subscriber before registering or deleting a user:

```sh
nats sub 'tenant.management.identity.user.>'
```

Verify:

- The XFSC event listener is installed and enabled.
- Keycloak is connected to NATS.
- The user transaction completed successfully.
- NATS publish permissions allow the event subjects.
- The subscriber is connected to the correct NATS account.

Core NATS does not replay events published before the subscription was established.

### Keycloak Admin Console fails to load

Verify the public hostname, proxy headers and Gateway routing.

Check the discovery endpoint:

```sh
curl -i \
  "$KEYCLOAK_URL/realms/master/.well-known/openid-configuration"
```

Check the administration translation endpoint:

```sh
curl -i \
  "$KEYCLOAK_URL/resources/master/admin/en"
```

Compare the public endpoint with a direct pod port-forward to distinguish Keycloak errors from Gateway routing issues.

## Security

- Restrict management subjects using NATS account and subject permissions.
- Separate management command publishers from user lifecycle event consumers.
- Restrict access to NATS reply inboxes.
- Use TLS for production NATS and Keycloak Admin API connections.
- Use dedicated service-account credentials for the identity controller.
- Store credentials in Kubernetes Secrets or an external secret management system.
- Do not commit credentials to source control.
- Restrict public access to administrative Keycloak endpoints.
- Do not expose raw Admin API errors to untrusted requesters without sanitization.
- Validate tenant identifiers, client identifiers and redirect URIs.
- Avoid logging access tokens, client secrets or sensitive user data.

## Delivery Guarantees and Limitations

The current implementation uses Core NATS queue subscriptions.

| Capability | Current implementation |
|---|---|
| Request-reply | Supported |
| Queue subscriptions | Supported |
| Durable message storage | Not supported |
| Event replay | Not supported |
| Exactly-once processing | Not guaranteed |
| Request correlation | Supported |
| Request deduplication | Not implemented |
| Durable event retries | Not implemented |
| Transactional outbox | Not implemented |

The Keycloak event listener logs publishing failures but does not provide durable delivery retries.

The `request_id` field is a correlation identifier, not an idempotency key.

## Production Roadmap

Recommended improvements for production deployments:

- [ ] Introduce JetStream-backed durable command processing.
- [ ] Introduce durable user lifecycle event delivery.
- [ ] Implement request deduplication and idempotency.
- [ ] Implement retry policies with exponential backoff.
- [ ] Introduce dead-letter subjects.
- [ ] Implement a transactional outbox for user lifecycle events.
- [ ] Introduce structured audit logging.
- [ ] Add controller health and readiness endpoints.
- [ ] Add NATS connectivity monitoring.
- [ ] Add provisioning metrics.
- [ ] Add automated integration and end-to-end tests.
- [ ] Implement client credential rotation.
- [ ] Replace the bundled message library snapshot with a pinned released dependency.

## References

- [Eclipse XFSC](https://github.com/eclipse-xfsc)
- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [Keycloak Server Administration](https://www.keycloak.org/docs/latest/server_admin/)
- [NATS Documentation](https://docs.nats.io/)
- [NATS CLI](https://github.com/nats-io/natscli)
- [CloudEvents Specification](https://cloudevents.io/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [External Secrets Operator](https://external-secrets.io/)
