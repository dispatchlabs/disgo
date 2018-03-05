#!/usr/bin/env bash

curl 'http://localhost:1975/v1/transactions/new' -X POST -d '{"hash":"1","from":"dl","to":"NODE-Nicolae","value":10,"time":"2018-01-01T17:34:51.617181-08:00"}'
