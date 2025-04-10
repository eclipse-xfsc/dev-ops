#!/usr/bin/env bash

set -euo pipefail

# ---- Variablen ----
CONFIG_FILE="$1"  # 1. Argument: Pfad zur Harbor-Config-Datei

# ---- Config-File einlesen ----
if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "Fehler: Konfigurationsdatei '$CONFIG_FILE' nicht gefunden."
  exit 1
fi

while IFS='=' read -r key value; do
  key=$(echo "$key" | xargs)     # Whitespace trimmen
  value=$(echo "$value" | xargs) # Whitespace trimmen
  case "$key" in
    PROJECT)
      export HARBOR_PROJECT="$value"
      ;;
    USER)
      RAW_USER="$value"  # Benutzername merken
      export HARBOR_USERNAME="robot\$${value}"
      ;;
  esac
done < "$CONFIG_FILE"

# ---- Prüfen, ob HARBOR_CREDENTIALS gesetzt ist ----
if [[ -z "${HARBOR_CREDENTIALS:-}" ]]; then
  echo "Fehler: Umgebungsvariable HARBOR_CREDENTIALS ist nicht gesetzt."
  exit 1
fi

# ---- Credentials aus Base64-String decodieren und Passwort extrahieren ----
DECODED_JSON=$(echo "$HARBOR_CREDENTIALS" | base64 -d || {
  echo "Fehler: Base64-Dekodierung von HARBOR_CREDENTIALS fehlgeschlagen."
  exit 1
})

HARBOR_PASSWORD=$(echo "$DECODED_JSON" | jq -r --arg user "$RAW_USER" '.[$user]')

# ---- Passwort prüfen ----
if [[ -z "$HARBOR_PASSWORD" || "$HARBOR_PASSWORD" == "null" ]]; then
  echo "Fehler: Kein Passwort für Benutzer '$RAW_USER' gefunden."
  exit 1
fi

echo "::add-mask::$HARBOR_PASSWORD"

export HARBOR_PASSWORD=$HARBOR_PASSWORD
export HARBOR_PROJECT=$HARBOR_PROJECT
export HARBOR_USERNAME=$HARBOR_USERNAME
