#!/usr/bin/env bash

curl -X POST -d '{"from": "123f681646d4a755815f9cb19e1acc8565a0c2ac", "to": "553f681646d4a755815f9cb19e1acc8533a0c2ac", "value": 1}' 'http://localhost:1973/v1/transactions'
