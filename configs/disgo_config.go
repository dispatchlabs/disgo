package configs

var (
	Config *DisgoConfig
)

type DisgoConfig struct {
	HttpPort int
	HttpHostIp string
	GrpcPort int
	GrpcTimeout int
}
