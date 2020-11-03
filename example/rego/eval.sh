#!/bin/bash

opa eval -i input.json -d policy.rego "query = data.main.auth_get_keytab"