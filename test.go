package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"


	"github.com/dispatchlabs/disgover"
)

func main2() {


	var seedNodeIP = os.Getenv("SEED_NODE_IP") // Needed when run from Kubernetes
	if len(seedNodeIP) == 0 {
		seedNodeIP = "127.0.0.1"
	}

	var dsg = disgover.NewDisgover(
		disgover.NewContact(),
		[]*disgover.Contact{
			&disgover.Contact{
				Id: "k0s66Hm6K85Jlg==",
				Endpoint: &disgover.Endpoint{
					Host: seedNodeIP,
					Port: 1975,
				},
			},
		},
	)
	dsg.ThisContact.Id = "NODE-2"
	dsg.ThisContact.Endpoint.Port = 9002
	dsg.Run()

	node2, _ := dsg.Find("k0s66Hm6K85Jlg==", dsg.ThisContact)

	if node2 == nil {
		fmt.Println("DISGOVER: Find() -> NOT FOUND")
	} else {
		fmt.Println(fmt.Sprintf("DISGOVER: Find() -> %s on [%s : %d]", node2.Id, node2.Endpoint.Host, node2.Endpoint.Port))
	}

	// party := party2.NewParty()
	// val, _ := party.GetVersion(context.Background(), &party2.Empty{})
	// fmt.Printf("Party version %s\n", val)
	// rslt := party.Join(dsg.ThisContact)
	// fmt.Printf("%s\n", rslt)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	<-done

}
