#!/usr/bin/env bash

# curl -X POST http://localhost:1975/v1/transactions -d '{"from":"e6098cc0d5c20c6c31c4d69f0201a02975264e94","code":"6060604052600160005534610000575b6101168061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609d565b005b6000816000540290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a723058203f0887849cabeb36c6f72cc345c5ff3521d889356357e6815dd8dbe9f7c41bbe0029"}'


curl -X POST http://localhost:1175/v1/transactions -d '{"hash":"0bb8c64779b0ab9d04b84b1d33d8cff40d4802c91cea815afcbee21b06895254","type":0,"from":"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c","to":"","value":0,"code":"6060604052600160005534610000575b6101168061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609d565b005b6000816000540290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a723058203f0887849cabeb36c6f72cc345c5ff3521d889356357e6815dd8dbe9f7c41bbe0029","method":"","time":1526859114441,"signature":"62411c9fe1b084ed6ccbe1f5a2ffc89484abce8fc7b482948d8012ab0e462b866a96d19bfd48f304c2d906c3ab950758c1d28ae7aba14e1bfb29ec8949967b4c01","hertz":0,"fromName":"","toName":""}'