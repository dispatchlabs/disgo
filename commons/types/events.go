package types

type commonServicesEvents struct {
	DbServiceInitFinished       string
	GrpcServiceInitFinished     string
	HttpServiceInitFinished     string
	DisGoverServiceInitFinished string
	DAPoSServiceInitFinished    string
	DVMServiceInitFinished      string
}

var (
	// Events - `services` events
	Events = commonServicesEvents{
		DbServiceInitFinished:       "DbServiceInitFinished",
		GrpcServiceInitFinished:     "GrpcServiceInitFinished",
		HttpServiceInitFinished:     "HttpServiceInitFinished",
		DisGoverServiceInitFinished: "DisGoverServiceInitFinished",
		DAPoSServiceInitFinished:    "DAPoSServiceInitFinished",
		DVMServiceInitFinished:      "DVMServiceInitFinished",
	}
)
