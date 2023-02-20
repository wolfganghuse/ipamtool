# ipamtool

Small CLI Helper Tool to reserve IPs from Nutanix IPAM-enabled Subnet.

## Prerequisite
Prism Central 2022.9+

## Configuration
```
Environment variables need to be set:
NUTANIX_PORT default:"9440"
NUTANIX_ENDPOINT
NUTANIX_USER default:"admin"
NUTANIX_PASSWORD
NUTANIX_INSECURE default:"true"
DEBUG default:"false"
NUTANIX_SUBNET_NAME
NUTANIX_SUBNET_UUID
```

For Subnet only one of NAME or UUID needs to be set, if Name is provided we try to evaluate this by ourselves.
(ToDo: Does not work if multiple Networks with same Name exists)

## Usage

```
ipamtool fetch
```
Shows all existing reservations and Client Contexts

```
ipamtool reserve (optionally Client Context)
```
Reserves a new IP Adress, if a Context is provided it will be assigned to this IP

```
ipamtool unreserve a.b.c.d (and/or) Client Context
```
To release an IP Adress you can provide the individual IP, if a context was configured this needs to be passed as additional parameter.
If you just provide the Client Context ALL IP Adresses with matching Context will be released.