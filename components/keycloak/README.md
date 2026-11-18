# XFSC Identity Tenant Management

This component bundles the theme, a global Keycloak event listener, and an optional NATS-to-Admin-API controller.

## Build

From `components/keycloak`:

```sh
docker build --no-cache -f deployment/docker/Dockerfile -t xfsc-keycloak .
docker build --no-cache -f controller/deployment/docker/Dockerfile -t xfsc-identity-controller controller
# Alle Container stoppen und entfernen
docker compose down --remove-orphans

# Alle selbst gebauten Images ohne Cache neu bauen
docker compose build --no-cache --pull

# Alle Images aktualisieren, auch fertige Images
docker compose pull --ignore-buildable

# Alles neu erstellen und starten
docker compose up -d --force-recreate

# Status prüfen
docker compose ps
```

The controller uses the included snapshot of `nats-message-library` to ensure the wire contracts match. Replace this snapshot with a pinned released module when available.

## Subjects

Request payloads use `common.Request` fields `tenant_id`, `request_id`, optional `group_id`. Publish commands to:

- `tenant.management.identity.create`, `.update`, `.delete`
- `tenant.management.identity.app.create`, `.app.update`, `.app.delete`

For a reply, set the NATS reply inbox. The controller returns `common.Reply` (and `appId` for app commands). A successful reply has `error: null`. This implementation uses Core NATS queue subscriptions; delivery, deduplication, authorization and persistence must be addressed before production use. `request_id` is a correlation ID, not currently an idempotency key.

### Create tenant

```json
{"tenant_id":"demo","request_id":"req-1","enabled":true,"registrationAllowed":true}
```

### Create authorization-code application

```json
{"tenant_id":"demo","request_id":"req-2","appId":"portal","name":"Portal","client":{"type":"public"},"authorization":{"grantTypes":["authorization_code"],"redirectUris":["https://portal.example.com/callback"],"pkceRequired":true,"pkceMethod":"S256"}}
```

Public clients require PKCE S256. Confidential clients need credential lifecycle handling outside this API. Updates are partial but PKCE attribute removal is not currently supported.

## Self-service events

The built-in global listener publishes CloudEvents 1.0 JSON to `tenant.management.identity.user.created` and `tenant.management.identity.user.deleted`. Their CloudEvent types are `org.eclipse.xfsc.identity.user.created` and `.deleted`. These events are published after commit, but Core NATS does not guarantee durable delivery.

## Helm

`helm install identity deployment/helm/xfsc-keycloak` deploys Keycloak and (by default) the controller. Configure database and bootstrap secrets, controller credentials, images, and NATS URL in values. The controller requires a dedicated service-account client with narrowly scoped realm/client-management permissions. Do not use bootstrap admin credentials in the controller.

## Security and limitations

- Restrict access to management subjects with NATS permissions; otherwise any publisher could create tenants or clients.
- Use TLS for production NATS and Admin API endpoints.
- Controller uses client-credentials grant, and its client must be provisioned separately.
- Core NATS does not provide replay or exactly-once processing. Add JetStream durable consumers, idempotency, and an outbox for reliable provisioning.
- The event listener logs failures but does not retry durable delivery.
- Do not expose raw Admin API errors to untrusted requesters without sanitization.
