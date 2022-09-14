package main

import (
	//"github.com/golang/glog"
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

type reservedIP struct {
    IP []string      `json:"reserved_ips"`
}

func ReserveIP(SubnetReserveUnreserveIpApiClient *networkingapi.SubnetReserveUnreserveIpApi, TasksApiClient *tasksapi.TaskApi, Subnet_UUID string, ClientContext string) (*common.IPAddress, error) {
	var ClientCount int64 = 1
	ipReserveInput:=networkingconfig.NewIpReserveInput()
	ipReserveInput.ClientContext = &ClientContext
	ipReserveInput.Count = &ClientCount
	ipReserveInput.ReserveType = networkingconfig.RESERVETYPE_IP_ADDRESS_COUNT.Ref()

	response, err := SubnetReserveUnreserveIpApiClient.ReserveIps(ipReserveInput, Subnet_UUID)   
	if err != nil {
		return nil , err
	} 
	data := response.GetData().(prism.TaskReference)

	responsetask, err := TasksApiClient.TaskGet(*data.ExtId)
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

func UnreserveIP(SubnetReserveUnreserveIPAPIClient *networkingapi.SubnetReserveUnreserveIpApi, TasksAPIClient *tasksapi.TaskApi, ClientContext string) (error) {
	IPUnreserveInput:=networkingconfig.NewIpUnreserveInput()
	IPUnreserveInput.UnreserveType= networkingconfig.UNRESERVETYPE_CONTEXT.Ref()
	IPUnreserveInput.ClientContext=&ClientContext
	response, err := SubnetReserveUnreserveIPAPIClient.UnreserveIps(IPUnreserveInput,serviceConfig.SubnetUUID)
	if err != nil {
		panic (err)
	}
	data := response.GetData().(prism.TaskReference)
	_, err = TasksAPIClient.TaskGet(*data.ExtId)
	if err != nil {
		return err
	}
	return nil
}


func findSubnetByName(SubnetIpApiClient *networkingapi.SubnetApi, name string) (*networkingconfig.Subnet, error) {
	page := 0
	limit := 20
	filter := fmt.Sprintf("name eq '%[1]v'", name)
	response, err := SubnetIpApiClient.ListSubnets(
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
