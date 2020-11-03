#!/bin/bash

enforcerd run --tag keytabs="superman:birdman" --tag secrets="secret1:secret2" --service-name oidc-shell bash -- "$@"
