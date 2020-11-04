package main # The package must be declared main

# The functions auth_get_nonce, auth_get_keytab and auth_get_secret must be
# implemented and must return a boolean value

default auth_get_nonce = false

default auth_get_keytab = false

default auth_get_secret = false

auth_base {
	# Here we match the authorized OAUTH issuer or issuers. It is very important that
	# this be defined and only for authorized providers.
	input.claims.iss == "abc123"
}

auth_get_nonce {
	# Here we authenticate who can get a Nonce. In this example we are just calling
	# the default auth_base
	auth_base
}

auth_nonce {
	# The input contains a set of all of the current valid nonces. For our
	# example here we expect the claim audience to have a nonce that will match
	# one of tne entries in the nonces set. You might ask your self what prevents
	# an attacker from getting a valid nonce and then launching a replay attack
	# with a captured token. This would require the attacker to modify the token
	# to include the Nonce which would corrupt the signature.
	input.nonces[_] == input.claims.aud
}

auth_get_keytab {
	# Here we call the default auth_base. Then we call auth_nonce to validate the nonce.
	# Finally we check to see if the token bearer is authorized to obtain the Keytab
	# by the name provided in the request. Notice that we expect the claim to have
	# zero or more entries delineated by colon.
	auth_base
	auth_nonce
	split(input.claims.service.keytabs, ":")[_] == input.name
}

auth_get_secret {
	# This is almost identical to auth_get_keytab. The only difference is that the
	# claim has been changed from service.keytabs to service.secrets
	auth_base
	auth_nonce
	split(input.claims.service.secrets, ":")[_] == input.name
}
