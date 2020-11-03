package main

# {
#   "claims": {
#     "alg": "EC",
#     "kid": "donut",
#     "iss": "abc123",
#     "exp": 1599844897,
#     "aud": "daisy",
#     "service": {
#       "keytab": "user1@example.com,user2@example.com"
#     }
#   },
#   "principal": "user1@example.com",
#   "nonces": ["daisy", "abigale", "poppy"]
# }

default auth_get_nonce = false
default auth_get_keytab = false
default auth_get_secret = false

auth_base {
   # Match Issuer
   input.claims.iss == "abc123"
}

auth_get_nonce {
   auth_base
}

auth_nonce {
   # The input contains a set of all of the current valid nonces. For our
   # example here we expect the claim audience to have a nonce that will match
   # one of tne entries in the nonces set.
   input.nonces[_] == input.claims.aud
}

auth_get_keytab {
   # The nonce must be validated and then the principal. This is done by splitting the
   # principals in the claim service.keytab by the comma into a set and checking for
   # match with requested principal
   auth_base
   auth_nonce
   split(input.claims.service.keytab,",")[_] == input.principal
}

auth_get_keytab {
   auth_base
   auth_nonce
}