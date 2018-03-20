#!/usr/bin/env bash

curl -X POST -d '{"hash":"5abdb41f762ca24f1915fc217b70f59b9c164d96eaee36f2265df339b4fcceb0","from":"7777f2b40aacbef5a5127f65418dc5f951280833","to":"c296220327589dc04e6ee01bf16563f0f53895bb","value":100,"time":"2018-03-16T15:29:04.287216-07:00","signature":"38fbce7db5d744700cd3deb9162d6783e4fee32039d7bb0210d27e316fa69f3e5a965fa6a8ff1bc8d1f3dac793277e2321a0425d5deab7278bdcb61cc9f9537a00"}' 'http://localhost:1975/v1/transactions'
