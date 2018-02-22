#!/usr/bin/env bash

curl -X POST -d '{"from": "aef26adf2dd6a622dbd9c66f34b4722b1c63524bba9dd3d11394c4bcdd7eaf86", "to": "123f681646d4a755815f9cb19e1acc8565a0c2ac", "value": 0.10}' 'http://localhost:1973/v1/transactions'
