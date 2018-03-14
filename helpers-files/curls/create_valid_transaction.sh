#!/usr/bin/env bash

curl -X POST -d '{"hash":"fb585ef5f78ea05b50e3a2dd4a0a8c9214e1f23d291cb3db0b9fdee7d08b0dd9","from":"9d6fa5845833c42e1aa4b768f944c5e09fe968b0","to":"c296220327589dc04e6ee01bf16563f0f53895bb","value":100,"time":"2018-03-13T14:54:02.378452-07:00","signature":"d4ae31aefc2659f9a48ba3b96a7b63a0d864de920581f464950699b7f8cbd7eb2a073bb291317dec45dfb25f2467618fe687aa01fbee8a1a51950dd0f65276e100"}' 'http://localhost:1975/v1/transactions'
