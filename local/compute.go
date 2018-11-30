/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package local

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"github.com/CS-SI/LocalDriver/metadata"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/HostProperty"
	"github.com/CS-SI/LocalDriver/model/enums/HostState"
	propsv1 "github.com/CS-SI/LocalDriver/model/properties/v1"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

const imagesJsonPath string = "/home/armand/Iso/images.json"
const templatesJsonPath string = "/home/armand/Iso/templates.json"
const libvirtStorage string = "/home/armand/LibvirtStorage"

//-------------IMAGES---------------------------------------------------------------------------------------------------

// ListImages lists available OS images
func (client *Client) ListImages(all bool) ([]model.Image, error) {
	// TODO implement ListAllImages(False)
	if ! all {
		return nil, fmt.Errorf("all==False not implemented yet")
	}

	jsonFile, err := os.Open(imagesJsonPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s : ", imagesJsonPath, err.Error())
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s : ", imagesJsonPath, err.Error())
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	imagesJson := result["images"].([]interface{})
	images := []model.Image{}
	for _, imageJson := range imagesJson {
		image := model.Image{
			imageJson.(map[string]interface{})["imageID"].(string),
			imageJson.(map[string]interface{})["imageName"].(string),
		}
		images = append(images, image)
	}

	return images, nil
}

// GetImage returns the Image referenced by id
func (client *Client) GetImage(id string) (*model.Image, error) {
	jsonFile, err := os.Open(imagesJsonPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s : ", imagesJsonPath, err.Error())
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s : ", imagesJsonPath, err.Error())
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	imagesJson := result["images"].([]interface{})
	for _, imageJson := range imagesJson {
		if imageID, _ := imageJson.(map[string]interface{})["imageID"]; imageID == id {
			return &model.Image{
				imageJson.(map[string]interface{})["imageID"].(string),
				imageJson.(map[string]interface{})["imageName"].(string),
			}, nil
		}
	}

	return nil, fmt.Errorf("Image with id=%s not found", id)
}



//-------------TEMPLATES------------------------------------------------------------------------------------------------

// ListTemplates overload OpenStack ListTemplate method to filter wind and flex instance and add GPU configuration
func (client *Client) ListTemplates(all bool) ([]model.HostTemplate, error) {
	// TODO implement ListAllTemplates(False)
	if ! all {
		return nil, fmt.Errorf("all==False not implemented yet")
	}

	jsonFile, err := os.Open(templatesJsonPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s : ", templatesJsonPath, err.Error())
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s : ", templatesJsonPath, err.Error())
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	templatesJson := result["templates"].([]interface{})
	templates := []model.HostTemplate{}
	for _, templateJson := range templatesJson {
		template := model.HostTemplate{
			HostTemplate : &propsv1.HostTemplate{
				HostSize: &propsv1.HostSize{
					Cores:     int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["coresNumber"].(float64)),
					RAMSize:   float32(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["ramSize"].(float64)),
					DiskSize:  int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["diskSize"].(float64)),
					GPUNumber: int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["gpuNumber"].(float64)),
					GPUType:   templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["gpuType"].(string),
				},
				ID:   templateJson.(map[string]interface{})["templateID"].(string),
				Name: templateJson.(map[string]interface{})["templateName"].(string),
			},
		}
		templates = append(templates, template)
	}

	return templates, nil
}

//GetTemplate overload OpenStack GetTemplate method to add GPU configuration
func (client *Client) GetTemplate(id string) (*model.HostTemplate, error) {
	jsonFile, err := os.Open(templatesJsonPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s : ", templatesJsonPath, err.Error())
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s : ", templatesJsonPath, err.Error())
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	templatesJson := result["templates"].([]interface{})
	for _, templateJson := range templatesJson {
		if templateID, _ := templateJson.(map[string]interface{})["templateID"]; templateID == id {
			return &model.HostTemplate{
				HostTemplate : &propsv1.HostTemplate{
					HostSize: &propsv1.HostSize{
						Cores:     int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["coresNumber"].(float64)),
						RAMSize:   float32(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["ramSize"].(float64)),
						DiskSize:  int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["diskSize"].(float64)),
						GPUNumber: int(templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["gpuNumber"].(float64)),
						GPUType:   templateJson.(map[string]interface{})["templateSpecs"].(map[string]interface{})["gpuType"].(string),
					},
					ID:   templateJson.(map[string]interface{})["templateID"].(string),
					Name: templateJson.(map[string]interface{})["templateName"].(string),
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("Template with id=%s not found", id)
}



//-------------SSH KEYS-------------------------------------------------------------------------------------------------

// CreateKeyPair creates and import a key pair
func (client *Client) CreateKeyPair(name string) (*model.KeyPair, error) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	pubKey := string(pubBytes)

	priBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	priKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priBytes,
		},
	)

	priKey := string(priKeyPem)
	return &model.KeyPair{
		ID:         uuid.NewV4().String(),
		Name:       name,
		PublicKey:  pubKey,
		PrivateKey: priKey,
	}, nil
}

// GetKeyPair returns the key pair identified by id
func (client *Client) GetKeyPair(id string) (*model.KeyPair, error) {
	panic("Not implemented yet")
}

// ListKeyPairs lists available key pairs
func (client *Client) ListKeyPairs() ([]model.KeyPair, error) {
	panic("Not implemented yet")
}

// DeleteKeyPair deletes the key pair identified by id
func (client *Client) DeleteKeyPair(id string) error {
	panic("Not implemented yet")
}



//-------------HOST MANAGEMENT------------------------------------------------------------------------------------------
// getImagePathFromID retrieve the storage path of an image from this image ID
func getImagePathFromID (id string) (string, error) {
	jsonFile, err := os.Open(imagesJsonPath)
	if err != nil {
		return "", fmt.Errorf("Failed to open %s : ", imagesJsonPath, err.Error())
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", fmt.Errorf("Failed to read %s : ", imagesJsonPath, err.Error())
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	imagesJson := result["images"].([]interface{})
	for _, imageJson := range imagesJson {
		if imageID, _ := imageJson.(map[string]interface{})["imageID"]; imageID == id {
			return imageJson.(map[string]interface{})["imagePath"].(string), nil
		}
	}

	return "", fmt.Errorf("Image with id=%s not found", id)
}

func getVolumesFromDomain(domain *libvirt.Domain, libvirtService *libvirt.Connect) ([]*libvirtxml.StorageVolume, error) {
	volumeDescriptions := []*libvirtxml.StorageVolume{}
	domainVolumePaths := []string{}

	//List paths of domain disks
	domainXML, err := domain.GetXMLDesc(0)
	if err != nil{
		return nil, fmt.Errorf(fmt.Sprintf("Failed get xml description of a domain : %s", err.Error()))
	}
	domainDescription := &libvirtxml.Domain{}
	err = xml.Unmarshal([]byte(domainXML), domainDescription)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed unmarshall the domain description : %s", err.Error()))
	}
	domainDisks := domainDescription.Devices.Disks

	for _, disk := range domainDisks {
		domainVolumePaths = append(domainVolumePaths, disk.Source.File.File)
	}


	//Check which volumes match these paths
	pools, err := libvirtService.ListAllStoragePools(2)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed list pools : %s", err.Error()))
	}
	for _, pool := range pools {
		volumes, err := pool.ListAllStorageVolumes(0)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Failed list storage volumes : %s", err.Error()))
		}
		for _, volume := range volumes {
			volumeXML, err := volume.GetXMLDesc(0)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Failed get xml description of a volume : %s", err.Error()))
			}
			volumeDescription := &libvirtxml.StorageVolume{}
			err = xml.Unmarshal([]byte(volumeXML), volumeDescription)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Failed unmarshall the volume description : %s", err.Error()))
			}

			for _, domainVolumePath := range domainVolumePaths {
				if volumeDescription.Key == domainVolumePath {
					volumeDescriptions = append(volumeDescriptions, volumeDescription)
				}
			}

		}
	}
	return volumeDescriptions, nil
}

//stateConvert convert libvirt.DomainState to a HostState.Enum
func stateConvert (stateLibvirt libvirt.DomainState) (HostState.Enum) {
	switch stateLibvirt {
	case 1:
		return HostState.STARTED
	case 3, 5:
		return HostState.STOPPED
	case 4:
		return HostState.STOPPING
	default:
		return HostState.ERROR
	}
}

// TODO implement the 3 get*FromDomain functions
func getDescriptionV1FromDomain(domain *libvirt.Domain, libvirtService *libvirt.Connect) (*propsv1.HostDescription, error) {
	hostDescription := propsv1.NewHostDescription()

	//var Created time.Time
	//var Creator string
	//var Updated time.Time
	//var Purpose string

	//There is a creation and modification timestamp on disks but it'not the best way to get the vm creation / modification date

	return hostDescription, nil
}
func getSizingV1FromDomain(domain *libvirt.Domain, libvirtService *libvirt.Connect) (*propsv1.HostSizing, error) {
	hostSizing := propsv1.NewHostSizing()

	info, err := domain.GetInfo()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get infos from the domain : %s", err.Error()))
	}

	diskSize := 0
	volumes, err := getVolumesFromDomain(domain, libvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get volumes from the domain : %s", err.Error()))
	}
	for _, volume := range volumes {
		diskSize += int(volume.Capacity.Value/1024/1024/1024)
	}

	hostSizing.AllocatedSize.RAMSize = 	float32(info.MaxMem)
	hostSizing.AllocatedSize.Cores =	int(info.NrVirtCpu)
	hostSizing.AllocatedSize.DiskSize =	diskSize
	// TODO GPU not implemented
	hostSizing.AllocatedSize.GPUNumber = 0
	hostSizing.AllocatedSize.GPUType = ""

	//hostSizing.RequestedSize and hostSizing.RequestedSize are unknown by libvirt and are left unset

	return hostSizing, nil
}
func getNetworkV1FromDomain(domain *libvirt.Domain, libvirtService *libvirt.Connect) (*propsv1.HostNetwork, error) {
	hostNetwork := propsv1.NewHostNetwork()

	//interfaces, err := domain.ListAllInterfaceAddresses(0)
	//if err != nil{
	//	return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch network interfaces : %s", err.Error()))
	//}
	//for _, iface := range interfaces {
	//	for _, address := range iface.Addrs {
	//		if address.Type == 0 {
	//			//hostNetwork.IPv4Addresses[] = address.Addr
	//		} else if address.Type == 1 {
	//			//hostNetwork.IPv4Addresses[] = address.Addr
	//		} else {
	//			return nil, fmt.Errorf(fmt.Sprintf("Unknown adressType"))
	//		}
	//	}
	//}
	//
	//ifaces, _ :=libvirtService.ListAllInterfaces(0)
	//for _, iface := range ifaces{
	//	str, _ := iface.GetXMLDesc(0)
	//	fmt.Println("\n!\n", str)
	//}

	return hostNetwork, nil
}

// getHostFromDomain build a model.Host struct representing a Domain
func getHostFromDomain (domain *libvirt.Domain, libvirtService *libvirt.Connect) (*model.Host, error) {
	id, err := domain.GetUUIDString()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch id from domain : ", err.Error() ))
	}
	name, err := domain.GetName()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch name from domain : ", err.Error() ))
	}
	state, _, err := domain.GetState()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch state from domain : ", err.Error() ))
	}
	hpDescriptionV1, err := getDescriptionV1FromDomain(domain, libvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain description : ", err.Error() ))
	}
	hpSizingV1, err := getSizingV1FromDomain(domain, libvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain sizing : ", err.Error() ))
	}
	hostNetworkV1, err := getNetworkV1FromDomain(domain, libvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain networks: ", err.Error() ))
	}

	host := model.NewHost()

	host.ID = id
	host.Name = name
	host.PrivateKey = "Impossible to fetch them from the domain, the private key is unknown by the domain for security reasons"
	host.LastState = stateConvert(state)
	host.Properties.Set(HostProperty.DescriptionV1, hpDescriptionV1)
	host.Properties.Set(HostProperty.SizingV1, hpSizingV1)
	host.Properties.Set(HostProperty.NetworkV1, hostNetworkV1)

	return host, nil
}

// getHostAndDomainFromRef retrieve the host and the domain associated to an ref (id or name)
func getHostAndDomainFromRef(ref string, libvirtService *libvirt.Connect) (*model.Host, *libvirt.Domain, error){
	domain, err  := libvirtService.LookupDomainByUUIDString(ref)
	if err != nil {
		domain, err  = libvirtService.LookupDomainByName(ref)
		if err != nil {
			return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to fetch domain from ref : %s", err.Error()))
		}
	}

	host, err := getHostFromDomain(domain, libvirtService)

	fmt.Println(domain.GetXMLDesc(0))

	return host, domain, nil
}

// CreateHost creates an host satisfying request
func (client *Client) CreateHost(request model.HostRequest) (*model.Host, error) {
	resourceName	:= request.ResourceName
	hostName 		:= request.HostName
	networkIDs		:= request.NetworkIDs
	publicIP		:= request.PublicIP
	templateID		:= request.TemplateID
	imageID			:= request.ImageID
	keyPair 		:= request.KeyPair

	if resourceName == "" {
		return nil, fmt.Errorf("The ResourceName is mandatory", resourceName)
	}
	if hostName == "" {
		hostName = resourceName
	}

	networks := []*model.Network{}
	networksCommandString := ""
	if ( networkIDs == nil || len(networkIDs) == 0 ) && ! publicIP {
		return nil, fmt.Errorf("The host must be on a network or be public", resourceName)
	}
	if networkIDs != nil && len(networkIDs) != 0 {
		for _, networkID := range networkIDs {
			network, err := client.GetNetwork(networkID)
			if err != nil {
				return nil, fmt.Errorf("Failed to get network %s : %s", networkID, err.Error())
			}
			networks = append(networks, network)
			networksCommandString += (" --network network=" + network.Name)
		}
	}
	if publicIP {
		// TODO do it in a generig way (not enp2s0)
		networksCommandString += " --network type=direct,source=enp2s0"
	}

	if templateID == "" {
		return nil, fmt.Errorf("The TemplateID is mandatory", resourceName)
	}
	if imageID == "" {
		return nil, fmt.Errorf("The ImageID is mandatory", resourceName)
	}
	if keyPair == nil {
		var err error
		keyPair, err = client.CreateKeyPair(fmt.Sprintf("key_%s",resourceName))
		if err != nil {
			return nil, fmt.Errorf("KeyPair creation failed : ", err.Error())
		}
	}

	template, err	:= client.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("GetTemplate failed : ", err.Error())
	}
	imagePath, err	:= getImagePathFromID(imageID)
	if err != nil {
		return nil, fmt.Errorf("GetImageFromPath failled : ", err.Error())
	}

	host, _, err := getHostAndDomainFromRef(resourceName, client.LibvirtService)
	if err == nil && host != nil {
		return nil, fmt.Errorf("The Host %s already exists", resourceName)
	}



	// TODO gpu is ignored
	// TODO remove sudo rights
	// TODO use libvirt-go functions not bash commands
	command_setup	 	:= fmt.Sprintf("IMAGE_PATH=\"%s\" && IMAGE=\"`echo $IMAGE_PATH | rev | cut -d/ -f1 | rev`\" && EXT=\"`echo $IMAGE | grep -o '[^.]*$'`\" && LIBVIRT_STORAGE=\"%s\" && HOST_NAME=\"%s\" && VM_IMAGE=\"$LIBVIRT_STORAGE/$HOST_NAME.$EXT\"", imagePath, libvirtStorage, hostName)
	command_copy 		:= fmt.Sprintf("cd $LIBVIRT_STORAGE && cp $IMAGE_PATH .")
	command_resize 		:= fmt.Sprintf("truncate $VM_IMAGE -s %dG && sudo virt-resize --expand /dev/sda1 $IMAGE $VM_IMAGE && rm $IMAGE", template.DiskSize)
	command_sysprep		:= fmt.Sprintf("sudo virt-sysprep -a $VM_IMAGE --hostname %s --ssh-inject root:string:\"%s\" --operations all,-ssh-hostkeys", hostName, keyPair.PublicKey)
	command_virt_install:= fmt.Sprintf("sudo virt-install --name=%s --vcpus=%d --memory=%d --import --disk=$VM_IMAGE %s --noautoconsole", resourceName, template.Cores, int(template.RAMSize*1024), networksCommandString)
	command := strings.Join([]string{command_setup, command_copy, command_resize, command_sysprep, command_virt_install}, " && ")

	cmd := exec.Command("bash", "-c", command)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Commands failled : ", err.Error())
	}

	defer func() {
		if err != nil {
			client.DeleteHost(resourceName)
		}
	}()

	domain, err  := client.LibvirtService.LookupDomainByName(resourceName)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Can't find domain %s : %s", resourceName, err.Error()))
	}

	host, err = getHostFromDomain(domain, client.LibvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get host %s from domain : %s", resourceName, err.Error()))
	}

	err = metadata.SaveHost(client, host)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (client *Client) GetHost(hostParam interface{}) (*model.Host, error) {
	var ref string

	switch hostParam.(type) {
	case string:
		ref = hostParam.(string)
	case *model.Host:
		ref = hostParam.(*model.Host).ID
	default:
		panic("host must be a string or a *model.Host!")
	}

	host, _, err := getHostAndDomainFromRef(ref, client.LibvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	return host, nil
}

func (client *Client) GetHostByName(name string) (*model.Host, error) {
	return client.GetHost(name)
}

// DeleteHost deletes the host identified by id
func (client *Client) DeleteHost(id string) error {
	host, domain, err := getHostAndDomainFromRef(id, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Destroy()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to destroy the domain : %s", err.Error()))
	}
	domain.Undefine()

	volumes, err := getVolumesFromDomain(domain, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to get the volumes from the domain : %s", err.Error()))
	}

	for _, volume := range volumes {
		os.Remove(volume.Key)
	}

	metadata.RemoveHost(client, host)

	return nil
}

// ListHosts lists available hosts
func (client *Client) ListHosts() ([]*model.Host, error) {
	var hosts []*model.Host

	domains, err := client.LibvirtService.ListAllDomains(16383)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error listing domains : %s", err.Error()))
	}
	for _, domain := range domains {
		host, err := getHostFromDomain(&domain, client.LibvirtService)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Failed to get host from domain : %s", err.Error()))
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

// StopHost stops the host identified by id
func (client *Client) StopHost(id string) error {
	_, domain, err := getHostAndDomainFromRef(id, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Shutdown()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to shutdown the host : %s", err.Error()))
	}

	return nil
}

// StartHost starts the host identified by id
func (client *Client) StartHost(id string) error {
	_, domain, err := getHostAndDomainFromRef(id, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Create()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to launch the host : %s", err.Error()))
	}

	return nil
}

// RebootHost reboot the host identified by id
func (client *Client) RebootHost(id string) error {
	_, domain, err := getHostAndDomainFromRef(id, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Reboot(0)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reboot the host : %s", err.Error()))
	}

	return nil
}

// GetHostState returns the host identified by id
func (client *Client) GetHostState(hostParam interface{}) (HostState.Enum, error) {
	host, err := client.GetHost(hostParam)
	if err != nil {
		return HostState.ERROR, err
	}
	return host.LastState, nil
}

//-------------Provider Infos-------------------------------------------------------------------------------------------

// ListAvailabilityZones lists the usable AvailabilityZones
func (client *Client) ListAvailabilityZones(all bool) (map[string]bool, error) {
	return map[string]bool {"local":true}, nil
}






