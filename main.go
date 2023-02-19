package main

import (
	"fmt"
	"net"
	"os"
)

/* Tag names to load configuration from environment variable */
const (
	ENV     = "env"
	DEFAULT = "default"
	ClientContext = "IPAMTool"
)

/* Non-exported instance to avoid accidental overwrite */
var serviceConfig Configuration

func main() {


	setConfig()
	//fmt.Printf("Service configuration : %+v\n ", serviceConfig)
	if serviceConfig.Debug=="false" {
		os.Stderr, _ = os.Open(os.DevNull)
	}

	
	nutanixClient, _ :=Connect(serviceConfig)

	
	// Check commandline-args
	if len(os.Args) < 2 {
        fmt.Println("expected 'reserve', 'unreserve IP / ClientContext' or 'fetch'")
        os.Exit(1)
    }

    switch os.Args[1] {

    case "reserve":
		//ClientContext := uuid.NewString()
		myIP, err:= ReserveIP(nutanixClient,ClientContext)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(*myIP.Ipv4.Value)
		

    case "unreserve":

		var ReleaseContext string
		var ipToRelease string

	
		if len(os.Args) < 3 {
			fmt.Println("expected IP or ClientContext to unreserve")
			os.Exit(1)
		}
		
		if net.ParseIP(os.Args[2]) == nil {
			//not an IP
			ReleaseContext=os.Args[2]
		} else {
			ipToRelease= os.Args[2]
		}
		err:= UnreserveIP(nutanixClient,ipToRelease, ReleaseContext)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	
	case "fetch":
		IPList, err:= FetchIPList(nutanixClient,serviceConfig.SubnetUUID)
		if err != nil {
			panic (err)
		}
		for _, ipitem := range (IPList) {
			fmt.Println(ipitem.ip, ipitem.context)
		}

	default:
        fmt.Println("expected 'reserve', 'unreserve a.b.c.d' or 'fetch")
        os.Exit(1)
    }

  
}
