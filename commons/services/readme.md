<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">

# Services

Currently Dispatch code runs three main functionality providers:

- GRPC service: provides a singleton GRPC server that still conforms to the i_service interface.  The approach to using this is that any component that uses GRPC will manage its own registration with the singleton grpc service.

- HTTP service: provides a singleton HTTP server that conforms to the i_service interface. The approach to using this is that any component that uses HTTP will manage its own registration with the singleton HTTP service.

- DB service: provides a singleton DB server that conforms to the i_service interface. The approach to using this is that any component that uses DB will manage its own registration with the singleton DB service.



### Registering with SERVICES

The exposed function ```GetSERVICEServer()``` provides access to the singleton instance of the server of that service.  
The current approach is to have each component that relies on the SERVICE to specify their own proto file and generated service apis.  The component then must implement a service implementation that conforms to the i_service interface and handles its own registration with the SERVICE'S server.

for example in the case of grpc: 
```GO:
func NewServiceThatUsesGPRC() *NewServiceType {

  newService := &NewServiceType{
    //set initial values
  }
  // use the generated registration function from protobuf
  // that is specific to this component
  proto.RegisterNewServiceTypeGrpcServer(grpcServer, newService)

  return newService
}
``` 



### Using services that are registered with a service

In order to use your component that has registered itself the calling code should have an array of i_service interface

```GO:
type Server struct {
	services   []types.IService
}
```

Add all of the different services that are desired.  Of note, we want the commons grpc_service to be started last so that the other services are all registered prior. (May revisit this)
```GO:
  server.services = append(server.services, package.NewServiceThatUsesGPRC())
  server.services = append(server.services, services.NewGrpcService())
```

We want to have all of the services in a WaitGroup listening for incoming messages.  For all services that implement the i_service interface, we expect the `GO()` function for each service to be called to start it listening.

```GO:
// Run services.
var waitGroup sync.WaitGroup
for _, service := range server.services {
  log.WithFields(log.Fields{
    "method": "Server.Go",
  }).Info("starting " + service.Name() + "...")
  go service.Go(&waitGroup)
  waitGroup.Add(1)
}
waitGroup.Wait()
```

After this loop all desired services are registered and listening.  The last to start is the grpc_service which sets up the listener, registers everything with GRPC and starts serving requests:  Here is the GO() function for grpc_service:

```GO:
func (grpcService *GrpcService) Go(waitGroup *sync.WaitGroup) {

	grpcService.running = true
	listener, error := net.Listen("tcp", ":"+strconv.Itoa(grpcService.Port))
	if error != nil {
		log.Fatalf("failed to listen: %v", error)
	}

	// Serve.
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Go",
	}).Info("listening on " + strconv.Itoa(grpcService.Port))
	reflection.Register(GetGrpcServer())
	if error := GetGrpcServer().Serve(listener); error != nil {
		log.Fatalf("failed to serve: %v", error)
		grpcService.running = false
	}
}
```

