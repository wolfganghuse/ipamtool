package main

import (

	"log"
	"io/ioutil"

	"fmt"
	"os"

	"github.com/google/uuid"
	//networkingconfig "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/networking/v4/config"
	"github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
)

/* Tag names to load configuration from environment variable */
const (
	ENV     = "env"
	DEFAULT = "default"
)

/* Non-exported instance to avoid accidental overwrite */
var serviceConfig Configuration

func main() {

	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)

	setConfig()
	//fmt.Printf("Service configuration : %+v\n ", serviceConfig)
	
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
		ClientContext := uuid.NewString()
		myIP, err:= ReserveIP(nutanixClient,serviceConfig.SubnetUUID,ClientContext)
		if err != nil {
			panic (err)
		}
		fmt.Println(*myIP.Ipv4.Value)
		

    case "unreserve":
		if len(os.Args) < 3 {
			fmt.Println("expected uuid to unreserve")
			os.Exit(1)
		}
		ClientContext:= os.Args[2]
		err:= UnreserveIP(nutanixClient,ClientContext)
		if err != nil {
			panic (err)
		}		
		//output := responsetask.GetData().(tasksprism.Task)
		fmt.Println("SUCCESS")
	
	case "fetch":
		response, err := nutanixClient.SubnetReserveUnreserveIPAPIClient.FetchSubnetAddressAssignments(&serviceConfig.SubnetUUID)   
		if err != nil {
			panic (err)
		} 
		for _, data := range response.GetData().([]config.AddressAssignmentInfo) {
			if *data.IsReserved {
				fmt.Printf("%s - %s\n",*data.IpAddress.Ipv4.Value, *data.ReservedDetails.ClientContext)
			}
		}

	default:
        fmt.Println("expected 'reserve', 'unreserve a.b.c.d' or 'fetch")
        os.Exit(1)
    }

  
}
