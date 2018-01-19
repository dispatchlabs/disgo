package configurations

var (
	Configuration *DisgoNodeConfig
)

type DisgoNodeConfig struct {
	HttpPort int
	HttpHostIp string
	RpcPort int
}
