package services

type commonServicesEvents struct {
	DbServiceInitFinished   string
	GrpcServiceInitFinished string
	HttpServiceInitFinished string
}

var (
	// Events - `services` events
	Events = commonServicesEvents{
		DbServiceInitFinished:   "DbServiceInitFinished",
		GrpcServiceInitFinished: "GrpcServiceInitFinished",
		HttpServiceInitFinished: "HttpServiceInitFinished",
	}
)
