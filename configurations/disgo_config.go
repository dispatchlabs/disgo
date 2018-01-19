package configurations

var (
	Configuration *DisgoConfig
)

type DisgoConfig struct {
	HttpPort int
	HttpHostIp string
	GrpcPort int
}
