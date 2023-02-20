package main

import (
	

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/api"
	"github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/client"
	prismclient "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/client"
	
	common "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/common/v1/config"
	"github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
	prism "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/prism/v4/config"
	prismapi "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/api"
	prismconfig "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/models/prism/v4/config"
)

// Configuration keeps all settings together
type Configuration struct {
	Port     	string `env:"NUTANIX_PORT" default:"9440"`
	Prism	    string `env:"NUTANIX_ENDPOINT" default:""`
	User    	string `env:"NUTANIX_USER" default:"admin"`
	Password	string `env:"NUTANIX_PASSWORD" default:""`
	Insecure 	string `env:"NUTANIX_INSECURE" default:"true"`
	Debug    	string `env:"DEBUG" default:"false"`
	Subnet   	string `env:"NUTANIX_SUBNET_NAME" default:""`
	SubnetUUID  string `env:"NUTANIX_SUBNET_UUID" default:""`
}
// IPItem contains IP Adress and ClientContext
type IPItem struct {
	ip string
	context string
}

type reservedIP struct {
    IP []string      `json:"reserved_ips"`
}

//V4NutanixClient contains API Objects
type V4NutanixClient struct {
	SubnetReserveUnreserveIPAPIClient *api.SubnetReserveUnreserveIpApi
	TasksAPIClient prismapi.TaskApi
	SubnetIPAPIClient *api.SubnetApi
	Configuration Configuration
}

//Connect to v4 API
func Connect(c Configuration) (n V4NutanixClient, err error){
	APIClientInstance := client.NewApiClient()
	APIClientInstance.Host = c.Prism // IPv4/IPv6 address or FQDN of the cluster
	Port, err:= strconv.Atoi(c.Port)
	if err != nil {
		return n,err
	} 
	APIClientInstance.Port = Port // Port to which to connect to
	APIClientInstance.Username = c.User // UserName to connect to the cluster
	APIClientInstance.Password = c.Password // Password to connect to the cluster
	
	if c.Debug == "true" {
		APIClientInstance.Debug = true
	}

	if c.Insecure=="true" {
		APIClientInstance.SetVerifySSL(false)
	} else {
		APIClientInstance.SetVerifySSL(true)
	}
	
	PrismAPIClientInstance := prismclient.NewApiClient()
	PrismAPIClientInstance.Host = c.Prism // IPv4/IPv6 address or FQDN of the cluster
	PrismAPIClientInstance.Port = Port // Port to which to connect to
	PrismAPIClientInstance.Username = c.User // UserName to connect to the cluster
	PrismAPIClientInstance.Password = c.Password // Password to connect to the cluster

	if c.Debug == "true" {
		PrismAPIClientInstance.Debug = true
	}

	if c.Insecure=="true" {
		PrismAPIClientInstance.SetVerifySSL(false)
	} else {
		PrismAPIClientInstance.SetVerifySSL(true)
	}

	n.SubnetReserveUnreserveIPAPIClient = api.NewSubnetReserveUnreserveIpApi(APIClientInstance)
	n.SubnetIPAPIClient = api.NewSubnetApi(APIClientInstance)
	n.TasksAPIClient = *prismapi.NewTaskApi(PrismAPIClientInstance)
	
	//Resolve Subnet Name to UUID if necessary
	if c.SubnetUUID=="" {
	
		Subnet, err:=findSubnetByName(n,serviceConfig.Subnet)
		if err != nil {
			panic (err)
		}
		c.SubnetUUID=*Subnet.ExtId
	}
	n.Configuration = c

	return n, nil
}

// HandleTask returns Task Status
// func HandleTask(n V4NutanixClient, TaskID *string) (error) {

// }
// ReserveIP returns single IP, needs Subnet UUID and ClientContext
func ReserveIP(n V4NutanixClient, ClientContext string) (common.IPAddress, error) {
	var ClientCount int64 = 1
	ReservedIP:=common.NewIPAddress()
	
	ipReserveInput:=*config.NewIpReserveInput()
	ipReserveInput.ClientContext = &ClientContext
	ipReserveInput.Count = &ClientCount
	ipReserveInput.ReserveType = config.RESERVETYPE_IP_ADDRESS_COUNT.Ref()

	//response, err := n.SubnetReserveUnreserveIPAPIClient.ReserveIps(ipReserveInput, SubnetUUID)   
	response, err := n.SubnetReserveUnreserveIPAPIClient.ReserveIps(&ipReserveInput, &n.Configuration.SubnetUUID)   
	if err != nil {
		return *ReservedIP , err
	} 
	data := response.GetData().(prism.TaskReference)
	responsetask, err:= n.TasksAPIClient.TaskGet(data.ExtId)
	if err != nil {
		return *ReservedIP , err
	}
	status, err := responsetask.GetData().(prismconfig.Task).Status.MarshalJSON()
	if string(status) == "\"FAILED\"" {
		return *ReservedIP, fmt.Errorf(*responsetask.Data.GetValue().(prismconfig.Task).LegacyErrorMessage)
	}

	ReservedIPv4:=common.NewIPv4Address()
	ipResponse:=reservedIP{}
	output := responsetask.GetData().(prismconfig.Task)
	
	for _ ,details:= range output.CompletionDetails {
		s:=details.Value.GetValue().(string)
		json.Unmarshal([]byte(s), &ipResponse)
		ReservedIPv4.Value=&ipResponse.IP[0]
	}
	ReservedIP.Ipv4=ReservedIPv4
	return *ReservedIP,nil
}

// UnreserveIP returns Err of nil if release was successful, needs Subnet UUID and ClientContext
func UnreserveIP(n V4NutanixClient, IP string, ClientContext string) (error) {
	IPUnreserveInput:=config.NewIpUnreserveInput()

	if IP=="" {
		IPUnreserveInput.UnreserveType= config.UNRESERVETYPE_CONTEXT.Ref()
		IPUnreserveInput.ClientContext=&ClientContext
	
	} else {
		ip:=common.NewIPAddress()
		ip.Ipv4 = common.NewIPv4Address()
		ip.Ipv4.Value = &IP
		IPUnreserveInput.UnreserveType= config.UNRESERVETYPE_IP_ADDRESS_LIST.Ref()
		IPUnreserveInput.ClientContext=&ClientContext
		IPUnreserveInput.IpAddresses = append(IPUnreserveInput.IpAddresses, *ip)
	}
	response, err := n.SubnetReserveUnreserveIPAPIClient.UnreserveIps(IPUnreserveInput,&n.Configuration.SubnetUUID)
	if err != nil {
		return err
	}
	data := response.GetData().(prism.TaskReference)
	resp, err := n.TasksAPIClient.TaskGet(data.ExtId)
	status, err := resp.GetData().(prismconfig.Task).Status.MarshalJSON()
	if string(status) == "\"FAILED\"" {
		return fmt.Errorf(*resp.Data.GetValue().(prismconfig.Task).LegacyErrorMessage)
	}

	if err != nil {
		return err
	}
	return nil
}

//FetchIPList returns List of all reserved IPs in given Subnet
func FetchIPList(n V4NutanixClient, SubnetUUID string) (IPList []IPItem, err error) {

	response, err := n.SubnetReserveUnreserveIPAPIClient.FetchSubnetAddressAssignments(&n.Configuration.SubnetUUID)   
	if err != nil {
		return nil, err
	}

	if (response.GetData())== nil {
		return nil, fmt.Errorf("no reserved IPs in subnet %s",n.Configuration.SubnetUUID)
	}
	for _, data := range response.GetData().([]config.AddressAssignmentInfo) {
		if *data.IsReserved {
			var ipitem IPItem
			ipitem.ip = *data.IpAddress.Ipv4.Value
			if data.ReservedDetails.ClientContext != nil {
				ipitem.context = *data.ReservedDetails.ClientContext
			}
			IPList = append(IPList, ipitem)
		}
	}

	return IPList, nil

}

//findSubnetByName returns Subnet UUID, needs name
func findSubnetByName(n V4NutanixClient, name string) (*config.Subnet, error) {
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
	found:=config.NewSubnet()
	for _, data := range response.GetData().([]config.Subnet) {
		found=&data
	}
	return found, nil
}
