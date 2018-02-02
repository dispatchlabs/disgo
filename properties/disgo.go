package properties

var (
	Properties *DisgoProperties
)

type DisgoProperties struct {
	HttpPort          int
	HttpHostIp        string
	GrpcPort          int
	GrpcTimeout       int
	UseQuantumEntropy bool
}
