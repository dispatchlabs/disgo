package types

type events struct {
	DbServiceInitFinished       string
	GrpcServiceInitFinished     string
	HttpServiceInitFinished     string
	DisGoverServiceInitFinished string
	DAPoSServiceInitFinished    string
	DVMServiceInitFinished      string
}

var (
	Events = events{
		DbServiceInitFinished:       "DbServiceInitFinished",
		GrpcServiceInitFinished:     "GrpcServiceInitFinished",
		HttpServiceInitFinished:     "HttpServiceInitFinished",
		DisGoverServiceInitFinished: "DisGoverServiceInitFinished",
		DAPoSServiceInitFinished:    "DAPoSServiceInitFinished",
		DVMServiceInitFinished:      "DVMServiceInitFinished",
	}
)
