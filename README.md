# ![](https://storage.googleapis.com/material-icons/external-assets/v4/icons/svg/ic_info_outline_black_24px.svg) This is the Dispatch Node
__Run this code to be part of the best business enabling, open-source chain on the planet. Period.__

# ![](https://storage.googleapis.com/material-icons/external-assets/v4/icons/svg/ic_directions_run_black_24px.svg) Run

- Install `Go` for your platform. The 1-2-3 steps are [here](https://github.com/dispatchlabs/samples/tree/master/golang-setup)

##### Dev Box
- `go get github.com/dispatchlabs/disgo`
- `cd ~/go/src/dispatchlabs/disgo`
- `go run main.go`

##### As Service
- `go get github.com/dispatchlabs/disgo`
- `cd $GOPATH/src/github.com/dispatchlabs/disgo`
- `go build`
- `sudo mkdir /go-binaries`
- `sudo mv ./disgo /go-binaries/`
- `sudo cp -r ./properties /go-binaries/`
- `sudo nano /etc/systemd/system/dispatch-disgo-node.service`
```shell
[Unit]
Description=Dispatch Disgo Node
After=network.target

[Service]
WorkingDirectory=/go-binaries
ExecStart=/go-binaries/disgo -asSeed -nodeId=NODE-Seed-001 -thisIp=35.227.162.40
Restart=on-failure

User=dispatch-services
Group=dispatch-services

[Install]
WantedBy=multi-user.target
```
- `sudo useradd dispatch-services -s /sbin/nologin -M`
- `sudo systemctl enable dispatch-disgo-node`
- `sudo systemctl start dispatch-disgo-node`
- `sudo journalctl -f -u dispatch-disgo-node`
- `sudo systemctl daemon-reload` if you change the service



# ![](https://storage.googleapis.com/material-icons/external-assets/v4/icons/svg/ic_code_black_24px.svg) Develop
- Install [protoc](https://github.com/google/protobuf/releases) compiler manually or by homebrew `$ brew install protobuf`
- Install `protoc-gen-go plugin`: `go get -u github.com/golang/protobuf/protoc-gen-go`
- Build Go bindings from `.proto` file. `protoc --go_out=plugins=grpc:. party/party.proto
- Use [GoLand](https://github.com/dispatchlabs/samples/tree/master/docker-debug-go-goland) or [VSCode](https://github.com/dispatchlabs/samples/tree/master/docker-debug-go-vscode)