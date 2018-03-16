#!/usr/bin/env bash

curl -X POST -d '{"privateKey":"9dc7a0f09dba1ae2fec78c5238a0917208bd6012e335eda0f6bef87bb7a15a30","from":"7777f2b40aacbef5a5127f65418dc5f951280833","to":"c296220327589dc04e6ee01bf16563f0f53895bb","value":100}' 'http://localhost:1975/v1/test_transaction'
