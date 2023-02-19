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

	//Resolve Subnet Name to UUID if necessary
	if serviceConfig.SubnetUUID=="" {
	
		Subnet, err:=findSubnetByName(nutanixClient,serviceConfig.Subnet)
		if err != nil {
			panic (err)
		}
		serviceConfig.SubnetUUID=*Subnet.ExtId
	}

	// Check commandline-args
	if len(os.Args) < 2 {
        fmt.Println("expected 'reserve', 'unreserve uuid' or 'fetch'")
        os.Exit(1)
    }

    switch os.Args[1] {

    case "reserve":
		//ClientContext := uuid.NewString()
		myIP, err:= ReserveIP(nutanixClient,serviceConfig.SubnetUUID,ClientContext)
		if err != nil {
			panic (err)
		}
		fmt.Println(*myIP.Ipv4.Value)
		

    case "unreserve":

		var ReleaseContext string
		var ipToRelease string

	
		if len(os.Args) < 3 {
			fmt.Println("expected IP to unreserve")
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
			panic (err)
		}		
		//output := responsetask.GetData().(tasksprism.Task)
		fmt.Println("SUCCESS")
	
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
