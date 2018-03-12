#!/usr/bin/env bash

curl -X POST -d '{"hash":"153abd30bbae632dcd8ce47c482a0f5ac7c7fff0e9aec8407f8937bd16f2ae47","from":"428273ddce244807d8dcb135b1657b81fa18e7f5","to":"428273ddce244807d8dcb135b1657b81fa18e7f5","value":10,"time":"2018-03-12T09:05:33.273633-07:00","signature":"f55ef22b42033ac9ca84e68d5decfa9d81f677e059b08cc08e8afb2e516163247fc4d1eed4bdf4735c170725bb982f58cd7219f5a114925cded687d3f9ea0b5b00"}' 'http://localhost:1975/v1/transactions'
