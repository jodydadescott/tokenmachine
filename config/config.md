# Example

This example assumes that you already have a token. Your token should look something like this
```json
{
  "service": {
    "keytabs": "superman:birdman",
    "scopes": "default",
    "secrets": "secret1:secret2"
  },
  "aud": "initial",
  "exp": 1604444245,
  "iat": 1604440645,
  "iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
  "sub": "5fa1d23bc3a26d00019b0458"
}
```

Get a nonce with the command

```bash
curl -H "Authorization: Bearer $token ${YOUR_SERVER}/getnonce"
```

You should get a Nonce object like this

```json
{
  "exp": 1604440705,
  "value": "eAoiph3AMuFPQM5NG7Rr5beQBIcPCPmtE1GhRzGoGHC5YYK8o2yh0Q5vt0EJw1iP"
}
```

Now get a new token with the audience (aud) field set to the value from the Nonce. This assume that we have a OPA/Rego policy that will use the audience field. The nonce is how we prevent a replay attack. The new token should look like this

```json
{
  "service": {
    "keytabs": "superman:birdman",
    "scopes": "default",
    "secrets": "secret1:secret2"
  },
  "aud": "eAoiph3AMuFPQM5NG7Rr5beQBIcPCPmtE1GhRzGoGHC5YYK8o2yh0Q5vt0EJw1iP",
  "exp": 1604444245,
  "iat": 1604440645,
  "iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
  "sub": "5fa1d23bc3a26d00019b0458"
}
```

Use this token to get a Keytab or SharedSecret. To get a Keytab use the command

```bash
curl -H "Authorization: Bearer $token ${SERVER}/getkeytab\?name=${NAME}"
```

Or use this command to get a SharedSecret

```bash
curl -H "Authorization: Bearer $token ${SERVER}/getsec\?name=${NAME}"
```

Get a new token with the audience (aud) field set to the value of the nonce and then get a Keytab with the command



```bash
curl -H "Authorization: Bearer $token" "${SERVER}"/getkeytab\?name="${NAME}"
```

[Example client scripts](example)






NAME="superman"
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

  [[ $NAME ]] || { err "Missing env var NAME"; return 2; }
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

  log "${YELLOW}Get keytab with token from above and NAME ${PURPLE}${NAME}${YELLOW}->${NC}\n"
  keytab=$(httpGet -H "Authorization: Bearer $token" "${SERVER}"/getkeytab\?name="${NAME}") || {
    log_fail
    return 3
  }
  echo "$keytab" | jq

  echo "$keytab" | jq -r '.base64file' | base64 -d > "$tmp/keytab"
  local principal
  principal=$(echo "$keytab" | jq -r '.principal')

  log "${YELLOW}Authenticate with Active Directory / Kerberos Server: ${NC}"
  sudo /usr/bin/kinit -V -k -t "$tmp/keytab" "${principal}" > /dev/null 2>&1 || {
    log_fail
    return 3
  }
  log_ok
}

function write() {
  log "${YELLOW}Mount Windows CIFS share with user ${PURPLE}${NAME}${YELLOW}: ${NC}"
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
