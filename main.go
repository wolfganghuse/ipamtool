package main

import (
	b64 "encoding/base64"
	"fmt"
	"os"

	//"github.com/golang/glog"
	"github.com/google/uuid"
	networkingapi "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/api"
	tasksapi "github.com/nutanix-core/ntnx-api-go-sdk-internal/tasks_go_sdk/v16/api"
	//tasksprism "github.com/nutanix-core/ntnx-api-go-sdk-internal/tasks_go_sdk/v16/models/prism/v4/config"
	//prism "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/prism/v4/config"
	networkingconfig "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/networking/v4/config"
	// common "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/common/v1/config"
	//"github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/common/v1/response"
)

/* Tag names to load configuration from environment variable */
const (
	ENV     = "env"
	DEFAULT = "default"
)

// Configuration keeps all settings together
type Configuration struct {
	Port     	string `env:"NUTANIX_PORT" default:"9440"`
	Prism	    string `env:"NUTANIX_ENDPOINT" default:"10.19.227.151"`
	User    	string `env:"NUTANIX_USER" default:"admin"`
	Password	string `env:"NUTANIX_PASSWORD" default:"Nutanix.123"`
	Insecure 	string `env:"NUTANIX_INSECURE" default:"true"`
	Debug    	string `env:"DEBUG" default:"true"`
	Subnet   	string `env:"NUTANIX_SUBNET_NAME" default:"test-domain-managed"`
	SubnetUUID  string `env:"NUTANIX_SUBNET_UUID" default:""`
}

/* Non-exported instance to avoid accidental overwrite */
var serviceConfig Configuration

func main() {
	setConfig()
	//fmt.Printf("Service configuration : %+v\n ", serviceConfig)
	

	SubnetReserveUnreserveIPAPIClient := networkingapi.NewSubnetReserveUnreserveIpApi()
	SubnetReserveUnreserveIPAPIClient.ApiClient.BasePath = "https://" + serviceConfig.Prism + ":" + serviceConfig.Port
	SubnetReserveUnreserveIPAPIClient.ApiClient.SetVerifySSL(!stob(serviceConfig.Insecure))
	SubnetReserveUnreserveIPAPIClient.ApiClient.Debug = stob(serviceConfig.Debug)
    
	SubnetReserveUnreserveIPAPIClient.ApiClient.DefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Basic %s",
			b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", serviceConfig.User, serviceConfig.Password)))),
	}


	TasksAPIClient := tasksapi.NewTaskApi()
	TasksAPIClient.ApiClient.BasePath = "https://" + serviceConfig.Prism + ":" + serviceConfig.Port
	TasksAPIClient.ApiClient.SetVerifySSL(!stob(serviceConfig.Insecure))
	TasksAPIClient.ApiClient.Debug = stob(serviceConfig.Debug)
	
	TasksAPIClient.ApiClient.DefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Basic %s",
			b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", serviceConfig.User, serviceConfig.Password)))),
	}
 
	//Resolve Subnet Name to UUID if necessary
	if serviceConfig.SubnetUUID=="" {
		SubnetIPAPIClient := networkingapi.NewSubnetApi()
		SubnetIPAPIClient.ApiClient.BasePath = "https://" + serviceConfig.Prism + ":" + serviceConfig.Port
		SubnetIPAPIClient.ApiClient.SetVerifySSL(!stob(serviceConfig.Insecure))
		SubnetIPAPIClient.ApiClient.Debug = stob(serviceConfig.Debug)
		
		SubnetIPAPIClient.ApiClient.DefaultHeaders = map[string]string{
			"Authorization": fmt.Sprintf("Basic %s",
				b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", serviceConfig.User, serviceConfig.Password)))),
		}
		Subnet, err:=findSubnetByName(SubnetIPAPIClient,serviceConfig.Subnet)
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
		myIP, err:= ReserveIP(SubnetReserveUnreserveIPAPIClient,TasksAPIClient,serviceConfig.SubnetUUID,ClientContext)
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
		err:= UnreserveIP(SubnetReserveUnreserveIPAPIClient,TasksAPIClient,ClientContext)
		if err != nil {
			panic (err)
		}		
		//output := responsetask.GetData().(tasksprism.Task)
		fmt.Println("SUCCESS")
	
	case "fetch":
		response, err := SubnetReserveUnreserveIPAPIClient.FetchSubnetAddressAssignments(serviceConfig.SubnetUUID)   
		if err != nil {
			panic (err)
		} 
		for _, data := range response.GetData().([]networkingconfig.AddressAssignmentInfo) {
			if *data.IsReserved {
				fmt.Printf("%s - %s\n",*data.IpAddress.Ipv4.Value, *data.ReservedDetails.ClientContext)
			}
		}

	default:
        fmt.Println("expected 'reserve', 'unreserve a.b.c.d' or 'fetch")
        os.Exit(1)
    }

	


    // output:=networkingconfig.NewIpReserveOutput()
 

    


  
}
