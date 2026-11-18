# XFSC Keycloak + Identity Controller Helm chart

Deploys Keycloak and identity controller separately, generates persistent Kubernetes secrets, and provisions a confidential `xfsc-provisioner` client in the `master` realm with the **realm role `admin`**. This role grants broad cross-realm administration; restrict it after validating your requirements.

## Install

`helm upgrade --install identity ./helm -n identity --create-namespace`

On initial installation, the chart generates three secrets: `<release>-keycloak-admin`, `<release>-keycloak-db`, `<release>-identity-controller-credentials`. Existing secrets are not replaced on upgrades (`lookup`); secrets have a `keep` policy. **Note:** Helm client-side dry-runs cannot look up live secrets. Secret regeneration after deletion or reinstall can break existing database/admin credentials. Back up these secrets. The database password is generated locally; it must also match the PostgreSQL server's configured user password.

## Operator injection

Set `injection.keycloak.enabled=true` and/or `injection.controller.enabled=true`. Keycloak then omits `KC_DB_URL`, `KC_DB_USERNAME`, `KC_DB_PASSWORD`, and `NATS_URL`; controller omits `NATS_URL` when enabled. The operator must supply these values through matching ResourceProviders. The operator recognizes annotations on workload/pod template, with `inject.xfsc.io/needs` and `inject.xfsc.io/env-replacements`. Example: `POSTGRES_URL` maps to `KC_DB_URL`; `POSTGRES_USER` to `KC_DB_USERNAME`; `POSTGRES_PASSWORD` to `KC_DB_PASSWORD`. Verify your actual ResourceProvider output names; a provider using other names requires adjusted replacements. No env-prefix is configured.

`resourceClaim.enabled` is **false** by default. When enabled, `resourceClaim.externalSecretName` must point to a real ESO ExternalSecret. ResourceClaim reconciliation is not equivalent to workload injection; the operator provisions the backend via its ExternalSecret contract.

## Provisioning

The Helm post-install/post-upgrade Job authenticates with Keycloak's bootstrap admin account, creates or updates the confidential client, enables service accounts, sets its Secret, and assigns the `master` realm role `admin` to the service-account user. The job uses the internal Keycloak Service URL and retries readiness. The provisioner script is mounted from a ConfigMap. Ensure `admin-cli` direct access grants and the bootstrap admin password are valid; existing Keycloak installations with a changed bootstrap admin password require supplying the current admin credentials via the configured Secret. `helm --wait` is recommended for readiness; hook execution still requires a running Keycloak API.

## Caveats

- The original chart has no database deployment: an external PostgreSQL instance must already be configured. Operator injection alone does not create the database.
- `KC_HOSTNAME` is the public issuer/hostname; set it consistently with your Keycloak deployment.
- Verify the controller image consumes the documented `KEYCLOAK_*` environment variables and that its image is reachable.
- The `admin` role is powerful; use dedicated least-privilege roles when practical.
- For existing Keycloak installations, the generated bootstrap password may not match the existing admin account. Use an existing Kubernetes Secret.
