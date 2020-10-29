#!/bin/bash
# shellcheck disable=SC2059,SC2145
########################################################################################
# Example Script showing how to get Keytab from Keytab Broker Server, mount CIFS
# filesystem and write data
#
# This script expands on kinit_client.bash
# 
# Process Overview
# 1) Get a token from the token server
# 2) Using the token get a nonce from the Keytab server
# 3) Using the nonce attribute value get a new token from the token server with the
#    audience (aud) set to the nonce value. This is what prevents a replay attack.
# 4) Using the new token, the nonce value and the name of the desired principal
#    request the keytab from the Keytab server
# 5) Decode the base64file attribute from the Keytab into a file
# 6) Using the principal name attribute from the Keytab principal attribute
#    obtain a TGT from the Kerberos server. Note that the principal attribute
#    will differ from the original principal.
# 7) Mount CIFS filesystem
# 8) Write Random string to mounted filesystem
# 9) Un-mount CIFS filesystem
########################################################################################

PRINCIPAL="superman@EXAMPLE.COM"
SERVER="35.153.18.49:8080"
TOKEN_SERVER="169.254.254.1"

NC='\033[0m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
PURPLE='\033[0;35m'

function main() {

  which curl > /dev/null 2>&1 || { err "curl not found in path"; return 2; }
  which jq > /dev/null 2>&1 || { err "jq not found in path"; return 2; }

  [[ $PRINCIPAL ]] || { err "Missing env var PRINCIPAL"; return 2; }
  [[ $SERVER ]] || { err "Missing env var SERVER"; return 2; }
  [[ $TOKEN_SERVER ]] || { err "Missing env var TOKENB_SERVER"; return 2; }

  tmp=$(mktemp -d)
  trap cleanup EXIT

  run || return $?
  write || return $?
  return 0
}

function run() {
  local token
  local nonce

  log "${YELLOW}Get token->${NC}\n"
  token=$(httpGet -H 'X-Aporeto-Metadata: secrets' "http://${TOKEN_SERVER}/token?type=OAUTH&audience=initial") || {
    log_fail
    return 3
  }
  print_token "$token"

  err

  log "${YELLOW}Get nonce with token from above->${NC}\n"
  nonce=$(httpGet -H "Authorization: Bearer $token" "${SERVER}/getnonce") || {
    log_fail
    return 3
  }
  echo "$nonce" | jq
  nonce=$(echo "$nonce" | jq -r '.value')

  err

  log "${YELLOW}Get token with audience (aud field) set to nonce $nonce->${NC}\n"
  token=$(httpGet -H 'X-Aporeto-Metadata: secrets' "http://${TOKEN_SERVER}/token?type=OAUTH&audience=${nonce}") || {
    log_fail
    return 3
  }
  print_token "$token"

  log "${YELLOW}Get keytab with token from above and principal ${PURPLE}${PRINCIPAL}${YELLOW}->${NC}\n"
  keytab=$(httpGet -H "Authorization: Bearer $token" "${SERVER}"/getkeytab\?principal="${PRINCIPAL}") || {
    log_fail
    return 3
  }
  echo "$keytab" | jq

  echo "$keytab" | jq -r '.base64file' | base64 -d > "$tmp/keytab"
  local principal_alias
  principal_alias=$(echo "$keytab" | jq -r '.principal')

  log "${YELLOW}Authenticate with Active Directory / Kerberos Server: ${NC}"
  sudo /usr/bin/kinit -V -k -t "$tmp/keytab" "${principal_alias}" > /dev/null 2>&1 || {
    log_fail
    return 3
  }
  log_ok
}

function write() {
  log "${YELLOW}Mount Windows CIFS share with user ${PURPLE}${PRINCIPAL}${YELLOW}: ${NC}"
  sudo mount /data || { log_fail; return 3; }
  log_ok

  local random_msg
  random_msg="This is my random message $(openssl rand -base64 18)"
  log "${YELLOW}Write random message ${PURPLE}$random_msg${YELLOW} to /data/random.txt: ${NC}"
  echo "$random_msg" >> /data/random.txt || { log_fail; return 3; }
  log_ok
  sudo umount /data > /dev/null 2>&1
}

function httpGet() {
  local code
  cat /dev/null > "$tmp/response" || return 2
  code=$(curl -s -o "$tmp/response" -w "%{http_code}" "$@")
  local emsg
  local response
  response=$(<"$tmp/response")
  [ "$code" == "200" ] || {
    emsg=$(echo "$response" | jq -r .error)
    [[ $emsg ]] && err "Keytab server returned error: code->$code, message->$emsg"
    return 10
  }
  echo "$response"
  return 0
}

function cleanup() { [[ "$tmp" ]] && rm -rf "$tmp"; }

function print_token() {
  local data
  data=$(echo "$@" | cut -d "." -f 2 | base64 -d 2> /dev/null)
  [[ $data ]] || { err "Invalid Token"; return 3; }

  local first_char=${data:0:1}
  local last_char="${data: -1}"

  [[ "$first_char" == "" ]] && { err "No data"; return 3; }
  [[ "$first_char" == "{" ]] || { err "Invalid JWT"; return 3; }

  if [ "$last_char" == "\"" ]; then
    data+="}"
  else
    data+="\"}"
  fi

  echo "$data" | jq
}

err() { printf "$@\n" 1>&2; }
log() { printf "$@" 1>&2; }
log_ok() { printf "${GREEN}Successful${NC}\n" 1>&2; }
log_fail() { printf "${RED}Failed${NC}\n" 1>&2; }

main "$@"
