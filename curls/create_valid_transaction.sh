#!/usr/bin/env bash

curl -X POST -d '{"hash":"d4076ff453736aad7c495c281e0dc39015adcc8934e4e23acd786238a4934a90","from":"19b1edf2e533fd638bad65bd2644e8af7591b3e7","to":"a0be735d740ed1a79ff1cfbe0a8dc7f53276c64d","value":10,"time":"2018-03-03T17:34:51.617181-08:00","signature":"3fd985992dc395b6918a04c077ff492212bb0db81e6290adf021265d00b429ad7e61b3f1ace73e65fdae440b2b4ed183c6a3661001f9ef0324d6a74109a4122300"}' 'http://localhost:1975/v1/transactions'
