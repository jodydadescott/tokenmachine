# Tokenmachine
Grants SharedSecrets, Kerberos Keytabs, and Nonces to bearers of authorized tokens.
Authorization is accompolished by executing an OPA/Rego policy configured by the operator.
Shared Secrets are for applications that require a common shared secret. Keytabs are for
Kerberos authentication and Nonces are provided as a method to prevent replay attacks.

## Operation

When an application or users (client) desires a resource from the Tokenmachine they must first
acquire an OAUTH compliant token (bearer token) from their Identity Provider (IDP). Using 
this bearer token the client should request a Nonce from the Tokenmachine. The Tokenmachine will
authorize the Nonce request and if authroized return a time scoped Nonce. The Nonce will contain
both a secret and an expiration time. The client should obtain a new token from their IDP with
the Nonce secret encoded in the claims. The client should then use the bearer token to make one
or more request to the Tokenmachine before the Nonce expires.

SharedSecrets contain a Secret and an expiration (exp) time in epoch seconds. If the SharedSecret has
reached half life then the fields NextSecret and NextExp will be present. These fields represent the
next period Secret and expiration. It is up to the user to make use of the SharedSecret and cordinate
the period changes.

Keytabs contain a base64 Keytab file in the field Base64File and a Kerberos principal in the 
field Principal. It is not currently possible to provide the next Keytab as it is for a 
SharedSecret. Generally the user should obtain a Keytab when required and then discard it.

Authentication and authorization is performed by validating a bearers token signature and executing
a pre-configured OPA/Rego policy. The signature validation works by following the tokens issuer and
obtaining the public key. The issuer name must match the TLS certificate in the HTTPS request. The
OPA/Rego policy may also be used to prevent replay attack by checking to see if a nonce pattern is
inside the payload of the bearer token.

The Tokenmachine server works by being preconfigured with SharedSecrets and Keytabs. We will
referfer to these as entities for short. Both entities must be configured with a seed. This 
seed should be kept secret as it will be used to generate the secret for the SharedSecret
and the password for the Keytab principal. Each entity must also be configured with a Lifetime
in seconds. Each entity will have a time period calculated by taking the seconds since epoch and
dividing it by the Lifetime and removing the remainder. This will provide us with the current
period and by adding the lifetime to this we can determine the next period.

When a request for a SharedSecret is made the current time period is determined and the
seed is used to obtain a one time code. This is then appended to the seed and a SHA-256
hash is calculated. This will be converted into a alpha-numeric string.

Keytabs operate in a similiar method but there are signifigant differences. At the top of each
period Keytabs are created using a password that is derived in a similiar method to how SharedSecrets
are created. This is stored in a map that is read from when a request for a Keytab is made.

Resilency or redundancy can be achieved by running more then one instance of the Tokenmachine server.
This should work without conflict as long as the seeds for each entity are the same. Once again it is
important that the seeds remain secret.


## Installation

For Windows it is intended to run the server as a Windows service. The binary is able
to install itself as a service and even start itself. Installation consist of placing
the binary in a desired location, setting the config(s), installing the service and
finally starting the service.

It is importnat that the service be configured to run as a Directory Admin. This is
required for the creation of keytabs.

The config will be covered below but on Windows a comma delineated list of configs is
stored in the Windows registry. This can be set and shown using the same binary.

Create a directory such as 'C:\Program Files\Tokens2Secrets' and install the binary
'tokens2Secrets.exe' into said direcotry. Using the Windows command prompt change
directory to the installation directory and perform the following task.

```bash
# Install the service
.\tokenmachine.exe service install

# Create an example config then edit it
.\tokenmachine.exe config example > example.yaml

# Set the config(s) where $CONFIG_FILES with the location to one or more config files or
#  http(s) urls
.\tokenmachine.exe service config set $CONFIG_FILES

# When ready start the service
.\tokenmachine.exe service start
```

For Linux and Darwin download and install the binary tokenmachine. Note that valid Keytabs will NOT
be issued on Linux and Darwin.

```bash
# Create an example config then edit it
./tokenmachine config example > example.yaml

# Run the server with
./tokenmachine --config example.yaml
```

## Example

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
