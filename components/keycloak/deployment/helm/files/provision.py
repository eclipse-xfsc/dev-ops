import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

base = os.environ["KEYCLOAK_URL"].rstrip("/")
client_id = os.environ["CLIENT_ID"]
client_secret = os.environ["CLIENT_SECRET"]
username = os.environ["ADMIN_USERNAME"]
password = os.environ["ADMIN_PASSWORD"]
role_name = os.environ.get("ADMIN_ROLE", "admin")

def request(method, path, token=None, payload=None, form=False):
    headers = {}
    if token:
        headers["Authorization"] = "Bearer " + token
    if payload is not None:
        if form:
            data = urllib.parse.urlencode(payload).encode()
            headers["Content-Type"] = "application/x-www-form-urlencoded"
        else:
            data = json.dumps(payload).encode()
            headers["Content-Type"] = "application/json"
    else:
        data = None
    req = urllib.request.Request(base + path, data=data, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=10) as response:
        raw = response.read()
        return json.loads(raw) if raw else None

for attempt in range(60):
    try:
        auth = request("POST", "/realms/master/protocol/openid-connect/token", payload={
            "grant_type": "password", "client_id": "admin-cli", "username": username, "password": password
        }, form=True)
        token = auth["access_token"]
        break
    except (urllib.error.URLError, ValueError, KeyError) as exc:
        if attempt == 59:
            raise RuntimeError("Keycloak admin authentication unavailable after retries") from exc
        time.sleep(3)

clients = request("GET", "/admin/realms/master/clients?" + urllib.parse.urlencode({"clientId": client_id}), token)
client = next((c for c in clients if c.get("clientId") == client_id), None)
config = {"clientId": client_id, "enabled": True, "publicClient": False,
          "serviceAccountsEnabled": True, "clientAuthenticatorType": "client-secret",
          "standardFlowEnabled": False, "directAccessGrantsEnabled": False,
          "secret": client_secret}
if client:
    request("PUT", "/admin/realms/master/clients/" + client["id"], token, config)
    internal_id = client["id"]
else:
    request("POST", "/admin/realms/master/clients", token, config)
    clients = request("GET", "/admin/realms/master/clients?" + urllib.parse.urlencode({"clientId": client_id}), token)
    internal_id = next(c["id"] for c in clients if c.get("clientId") == client_id)

service_account = request("GET", "/admin/realms/master/clients/" + internal_id + "/service-account-user", token)
role = request("GET", "/admin/realms/master/roles/" + urllib.parse.quote(role_name, safe=""), token)
request("POST", "/admin/realms/master/users/" + service_account["id"] + "/role-mappings/realm", token, [role])
print("Provisioned confidential service-account client and master realm role:", client_id, role_name)
