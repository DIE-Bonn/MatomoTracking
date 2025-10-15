#!/bin/sh -eu
set -o pipefail || true

ini="/var/www/html/config/config.ini.php"
timeout="${INIT_WAIT_TIMEOUT:-300}"
min_lines="${INIT_MIN_LINES:-20}"

echo "INIT start: $(date)"
echo "uid=$(id -u) gid=$(id -g) host=$(hostname)"

# Wait until Matomo created the config and it has enough content
echo "Waiting for $ini (>= ${min_lines} lines, timeout ${timeout}s) ..."
start=$(date +%s)
while :; do
  if [ -f "$ini" ]; then
    lines=$(sed -n '$=' "$ini" 2>/dev/null || echo 0)
    if [ "${lines:-0}" -ge "$min_lines" ]; then
      break
    fi
  fi
  now=$(date +%s)
  if [ $((now - start)) -ge "$timeout" ]; then
    echo "Timeout: $ini missing or too small (< ${min_lines} lines)" >&2
    exit 1
  fi
  sleep 2
done
echo "Found: $ini (${lines} lines)"

# Ensure [General] section exists
grep -q '^\[General\]' "$ini" || printf '\n[General]\n' >> "$ini"

# Idempotent inserts
grep -Eq '^[[:space:]]*proxy_client_headers\[\]' "$ini" || sed -i '/^\[General\]/a\
proxy_client_headers[] = "HTTP_X_FORWARDED_FOR"' "$ini"

grep -Eq '^[[:space:]]*proxy_ip_read_last_in_list' "$ini" || sed -i '/^\[General\]/a\
proxy_ip_read_last_in_list = 0' "$ini"

# Ensure [Tracker] section exists
grep -q '^\[Tracker\]' "$ini" || printf '\n[Tracker]\n' >> "$ini"

# Set ip_address_mask_length = 0 (disable IP anonymization) within [Tracker]
if sed -n '/^\[Tracker\]/,/^\[/{/^[[:space:]]*ip_address_mask_length[[:space:]]*=/p;}' "$ini" | grep -q '='; then
  sed -i '/^\[Tracker\]/,/^\[/{s/^[[:space:]]*ip_address_mask_length[[:space:]]*=.*/ip_address_mask_length = 0/}' "$ini"
else
  sed -i '/^\[Tracker\]/a\
ip_address_mask_length = 0' "$ini"
fi

# Ensure full IP is used for enrichment as well
if sed -n '/^\[Tracker\]/,/^\[/{/^[[:space:]]*use_anonymized_ip_for_visit_enrichment[[:space:]]*=/p;}' "$ini" | grep -q '='; then
  sed -i '/^\[Tracker\]/,/^\[/{s/^[[:space:]]*use_anonymized_ip_for_visit_enrichment[[:space:]]*=.*/use_anonymized_ip_for_visit_enrichment = 0/}' "$ini"
else
  sed -i '/^\[Tracker\]/a\
use_anonymized_ip_for_visit_enrichment = 0' "$ini"
fi

chown -R 33:33 /var/www/html/config
chmod 660 "$ini" || true

echo "matomo-config-init done: $(date)"