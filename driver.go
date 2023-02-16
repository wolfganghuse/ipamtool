package main

import (
	//"github.com/golang/glog"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"

	networkingapi "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/api"
	networkingconfig "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/networking/v4/config"
	prism "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/prism/v4/config"
	tasksapi "github.com/nutanix-core/ntnx-api-go-sdk-internal/tasks_go_sdk/v16/api"

	//"github.com/nutanix-core/ntnx-api-go-sdk-internal/tasks_go_sdk/v16/models/common/v1/config"
	tasksprism "github.com/nutanix-core/ntnx-api-go-sdk-internal/tasks_go_sdk/v16/models/prism/v4/config"

	common "github.com/nutanix-core/ntnx-api-go-sdk-internal/networking_go_sdk/v16/models/common/v1/config"
)

// Configuration keeps all settings together
type Configuration struct {
	Port     	string `env:"NUTANIX_PORT" default:"9440"`
	Prism	    string `env:"NUTANIX_ENDPOINT" default:"10.48.38.43"`
	User    	string `env:"NUTANIX_USER" default:"admin"`
	Password	string `env:"NUTANIX_PASSWORD" default:"Nutanix.123"`
	Insecure 	string `env:"NUTANIX_INSECURE" default:"true"`
	Debug    	string `env:"DEBUG" default:"false"`
	Subnet   	string `env:"NUTANIX_SUBNET_NAME" default:"IPAM_nestedPool_2"`
	SubnetUUID  string `env:"NUTANIX_SUBNET_UUID" default:""`
}
type reservedIP struct {
    IP []string      `json:"reserved_ips"`
}

//V4NutanixClient contains API Objects
type V4NutanixClient struct {
	SubnetReserveUnreserveIPAPIClient *networkingapi.SubnetReserveUnreserveIpApi
	TasksAPIClient *tasksapi.TaskApi
	SubnetIPAPIClient *networkingapi.SubnetApi
}

//Connect to v4 API
func Connect(c Configuration) (n V4NutanixClient, err error){
	n.SubnetReserveUnreserveIPAPIClient = networkingapi.NewSubnetReserveUnreserveIpApi()
	n.SubnetReserveUnreserveIPAPIClient.ApiClient.BasePath = "https://" + c.Prism + ":" + c.Port
	n.SubnetReserveUnreserveIPAPIClient.ApiClient.SetVerifySSL(!stob(c.Insecure))
	n.SubnetReserveUnreserveIPAPIClient.ApiClient.Debug = stob(c.Debug)
	n.SubnetReserveUnreserveIPAPIClient.ApiClient.DefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Basic %s",
			b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.User, c.Password)))),
	}


	n.TasksAPIClient = tasksapi.NewTaskApi()
	n.TasksAPIClient.ApiClient.BasePath = "https://" + c.Prism + ":" + c.Port
	n.TasksAPIClient.ApiClient.SetVerifySSL(!stob(c.Insecure))
	n.TasksAPIClient.ApiClient.Debug = stob(c.Debug)
	n.TasksAPIClient.ApiClient.DefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Basic %s",
			b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.User, c.Password)))),
	}
 
	n.SubnetIPAPIClient = networkingapi.NewSubnetApi()
	n.SubnetIPAPIClient.ApiClient.BasePath = "https://" + c.Prism + ":" + c.Port
	n.SubnetIPAPIClient.ApiClient.SetVerifySSL(!stob(c.Insecure))
	n.SubnetIPAPIClient.ApiClient.Debug = stob(c.Debug)
	n.SubnetIPAPIClient.ApiClient.DefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Basic %s",
			b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.User, c.Password)))),
	}
	
	return n, nil
}

// ReserveIP returns single IP, needs Subnet UUID and ClientContext
func ReserveIP(n V4NutanixClient, SubnetUUID string, ClientContext string) (*common.IPAddress, error) {
	var ClientCount int64 = 1
	ipReserveInput:=networkingconfig.NewIpReserveInput()
	ipReserveInput.ClientContext = &ClientContext
	ipReserveInput.Count = &ClientCount
	ipReserveInput.ReserveType = networkingconfig.RESERVETYPE_IP_ADDRESS_COUNT.Ref()

	response, err := n.SubnetReserveUnreserveIPAPIClient.ReserveIps(ipReserveInput, SubnetUUID)   
	if err != nil {
		return nil , err
	} 
	data := response.GetData().(prism.TaskReference)

	responsetask, err := n.TasksAPIClient.TaskGet(*data.ExtId)
	if err != nil {
		return nil , err
	}

	ReservedIPv4:=common.NewIPv4Address()
	ipResponse:=reservedIP{}
	output := responsetask.GetData().(tasksprism.Task)
	for _ ,details:= range output.CompletionDetails {
		s:=details.Value.GetValue().(string)
		json.Unmarshal([]byte(s), &ipResponse)
		ReservedIPv4.Value=&ipResponse.IP[0]
	}
	ReservedIP:=common.NewIPAddress()
	ReservedIP.Ipv4=ReservedIPv4
	return ReservedIP,nil
}

// UnreserveIP returns Err of nil if release was successful, needs Subnet UUID and ClientContext
func UnreserveIP(n V4NutanixClient, ClientContext string) (error) {
	IPUnreserveInput:=networkingconfig.NewIpUnreserveInput()
	IPUnreserveInput.UnreserveType= networkingconfig.UNRESERVETYPE_CONTEXT.Ref()
	IPUnreserveInput.ClientContext=&ClientContext
	response, err := n.SubnetReserveUnreserveIPAPIClient.UnreserveIps(IPUnreserveInput,serviceConfig.SubnetUUID)
	if err != nil {
		panic (err)
	}
	data := response.GetData().(prism.TaskReference)
	_, err = n.TasksAPIClient.TaskGet(*data.ExtId)
	if err != nil {
		return err
	}
	return nil
}

//findSubnetByName returns Subnet UUID, needs name
func findSubnetByName(n V4NutanixClient, name string) (*networkingconfig.Subnet, error) {
	page := 0
	limit := 20
	filter := fmt.Sprintf("name eq '%[1]v'", name)
	response, err := n.SubnetIPAPIClient.ListSubnets(
		&page, &limit, &filter, nil)
	if err != nil {
		return nil , err
	}

	if *response.Metadata.TotalAvailableResults > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please use subnet_uuid argument or use additional filters instead")
	}

	if *response.Metadata.TotalAvailableResults == 0{
		return nil, fmt.Errorf("subnet with the given name, not found")
	}

	if response.GetData() == nil {
		return nil, fmt.Errorf("subnet query call failed")
	}
	found:=networkingconfig.NewSubnet()
	for _, data := range response.GetData().([]networkingconfig.Subnet) {
		found=&data
	}
	return found, nil
}
