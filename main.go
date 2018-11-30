package main

import (
	"fmt"
	"github.com/CS-SI/LocalDriver/local"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/IPVersion"
)

func main() {
	tenant := make(map[string]interface{})
	tenant["uri"] = "qemu:///system"
	tenant["minioEndpoint"] ="localhost:9000"
	tenant["minioAccessKeyID"] = "accesKey"
	tenant["minioSecretAccessKey"] = "secretKey"
	tenant["minioUseSSL"] = false

	local, err := (&local.Client{}).Build(tenant)
	if err != nil {
		fmt.Println("Build failed : ", err.Error())
		return
	}

	networkRequest := model.NetworkRequest{
		Name:	 "network_test",
		IPVersion: IPVersion.IPv4,
		CIDR:	 "10.8.0.1/24",
		DNSServers:	[]string{},
	}
	network, err := local.CreateNetwork(networkRequest)
	if err != nil {
		fmt.Println("Create network failed : ", err.Error())
		return
	}
	fmt.Println(network)



	hostRequest := model.HostRequest{
		ResourceName:	"vm_de_test",
		PublicIP: 		true,
		NetworkIDs:		[]string{network.ID},
		TemplateID:		"03014bb3-9096-49a7-bf5e-0f9e440ad7c6",
		ImageID:		"8891e5fc-b42b-49a0-b852-569cc1f1062d",
	}
	host, err := local.CreateHost(hostRequest)
	if err != nil {
		fmt.Println("Create Host failed : ", err.Error())
		return
	}
	fmt.Println(host)
	//local.DeleteHost("ded58a11-79a7-4c1c-9f6b-a992fdbda21c")

	//fmt.Println(local.StartHost("e46a237e-827d-4060-8e16-ccc9b29caf8f"))

	//fmt.Println(local.RebootHost("e46a237e-827d-4060-8e16-ccc9b29caf8f"))

	//local.DeleteHost(host.ID)

	//local.GetHost("247bdbc5-014f-4825-8995-97f5c110c2eb")
	//local.GetHost("a1a77375-058f-47ab-adc2-9b6535f4597f")



	//
	//gatewayRequest := model.GWRequest{
	//	NetworkID: network.ID,
	//	GWName: 	"",
	//	ImageID: 	"8891e5fc-b42b-49a0-b852-569cc1f1062d",
	//	TemplateID: "03014bb3-9096-49a7-bf5e-0f9e440ad7c6",
	//	KeyPair:	 nil,
	//}
	//
	//fmt.Println(local.CreateGateway(gatewayRequest))
}
