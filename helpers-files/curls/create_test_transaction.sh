#!/usr/bin/env bash

curl -X POST -d '{"privateKey":"dbb9eb135089c47e7ae678eed35933e13efa79c88731794add26c1a370b9efc9","from":"9d6fa5845833c42e1aa4b768f944c5e09fe968b0","to":"c296220327589dc04e6ee01bf16563f0f53895bb","value":100}' 'http://localhost:1975/v1/test_transaction'
