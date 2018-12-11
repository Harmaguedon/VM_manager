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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/HostProperty"
	"github.com/CS-SI/LocalDriver/model/enums/HostState"
	propsv1 "github.com/CS-SI/LocalDriver/model/properties/v1"
	"github.com/CS-SI/LocalDriver/userdata"
	"github.com/CS-SI/LocalDriver/utils/retry"
	libvirt "github.com/libvirt/libvirt-go"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

const imagesJsonPath string = "/home/armand/Iso/images.json"
const templatesJsonPath string = "/home/armand/Iso/templates.json"
const libvirtStorage string = "/home/armand/LibvirtStorage"
const startupPath string = "/home/armand/go/src/github.com/CS-SI/LocalDriver/startup.sh"

//-------------IMAGES---------------------------------------------------------------------------------------------------

// ListImages lists available OS images
func (client *Client) ListImages(all bool) ([]model.Image, error) {
	if !all {
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
		if imageID, ok := imageJson.(map[string]interface{})["imageID"]; ok && imageID == id {
			return &model.Image{
				imageJson.(map[string]interface{})["imageID"].(string),
				imageJson.(map[string]interface{})["imageName"].(string),
			}, nil
		}
		if imageName, ok := imageJson.(map[string]interface{})["imageName"]; ok && imageName == id {
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
	if !all {
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
			HostTemplate: &propsv1.HostTemplate{
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
				HostTemplate: &propsv1.HostTemplate{
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
func getImagePathFromID(id string) (string, error) {
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
	if err != nil {
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
func stateConvert(stateLibvirt libvirt.DomainState) HostState.Enum {
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

// TODO implement the getDescriptionV1FromDomain
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
		diskSize += int(volume.Capacity.Value / 1024 / 1024 / 1024)
	}

	hostSizing.AllocatedSize.RAMSize = float32(info.MaxMem)
	hostSizing.AllocatedSize.Cores = int(info.NrVirtCpu)
	hostSizing.AllocatedSize.DiskSize = diskSize
	// TODO GPU not implemented
	hostSizing.AllocatedSize.GPUNumber = 0
	hostSizing.AllocatedSize.GPUType = ""

	//hostSizing.RequestedSize and hostSizing.Template are unknown by libvirt and are left unset

	return hostSizing, nil
}
func (client *Client) getNetworkV1FromDomain(domain *libvirt.Domain) (*propsv1.HostNetwork, error) {
	hostNetwork := propsv1.NewHostNetwork()

	domainXML, err := domain.GetXMLDesc(0)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed get xml description of a domain : %s", err.Error()))
	}
	domainDescription := &libvirtxml.Domain{}
	err = xml.Unmarshal([]byte(domainXML), domainDescription)

	networks, err := client.LibvirtService.ListAllNetworks(3)
	if err != nil {
		return nil, fmt.Errorf("Failed to list all networks : %s", err.Error())
	}

	for _, iface := range domainDescription.Devices.Interfaces {
		if iface.Source.Network != nil {
			err = retry.WhileUnsuccessfulDelay5Seconds(
				func() error {
					for _, network := range networks {
						name, err := network.GetName()
						if err != nil {
							return fmt.Errorf("Failed to get network name : %s", err.Error())
						}
						if name == iface.Source.Network.Network {
							dhcpLeases, err := network.GetDHCPLeases()
							if err != nil {
								return fmt.Errorf("Failed to get network dhcpLeases : %s", err.Error())
							}
							for _, dhcpLease := range dhcpLeases {
								if dhcpLease.Mac == iface.MAC.Address {
									net, err := client.GetNetwork(iface.Source.Network.Network)
									if err != nil {
										return fmt.Errorf("Unknown Network %s", iface.Source.Network.Network)
									}
									if len(strings.Split(dhcpLease.IPaddr, ".")) == 4 {
										hostNetwork.IPv4Addresses[net.ID] = dhcpLease.IPaddr
									} else if len(strings.Split(dhcpLease.IPaddr, ":")) == 8 {
										hostNetwork.IPv6Addresses[net.ID] = dhcpLease.IPaddr
									} else {
										return fmt.Errorf("Unknown adressType")
									}
									hostNetwork.NetworksByID[net.ID] = net.Name
									hostNetwork.NetworksByName[net.Name] = net.ID
									return nil
								}
							}
						}
					}
					return fmt.Errorf("No local IP matching inteface %s found", iface.Alias)
				},
				5*time.Minute,
			)

		}
		if iface.Source.Direct != nil {
			var ip string
			err = retry.WhileUnsuccessfulDelay5Seconds(
				func() error {
					//TODO dynamic ip range
					cmd := exec.Command("bash", "-c", "sleep 30 && nmap -T5 -sP --host-timeout 1 172.26.128.0/24 > /dev/null && arp | grep "+iface.MAC.Address+" | cut -f1 -d\\ ")
					cmdOutput := &bytes.Buffer{}
					cmd.Stdout = cmdOutput
					err = cmd.Run()
					if err != nil {
						return fmt.Errorf("Commands failled : ", err.Error())
					}
					ip = strings.Trim(fmt.Sprintf("%s", cmdOutput), " \n")
					if len(strings.Split(ip, ".")) == 4 {
						hostNetwork.PublicIPv4 = ip
					} else if len(strings.Split(ip, ":")) == 8 {
						hostNetwork.PublicIPv6 = ip
					} else {
						return fmt.Errorf("Unknown adressType")
					}
					return nil
				},
				5*time.Minute,
			)
		}
	}
	return hostNetwork, nil
}

//TODO propose a version without the properties(time consuming)
// getHostFromDomain build a model.Host struct representing a Domain
func (client *Client) getHostFromDomain(domain *libvirt.Domain) (*model.Host, error) {
	id, err := domain.GetUUIDString()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch id from domain : %s", err.Error()))
	}
	name, err := domain.GetName()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch name from domain : %s", err.Error()))
	}
	state, _, err := domain.GetState()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to fetch state from domain : %s", err.Error()))
	}
	hostDescriptionV1, err := getDescriptionV1FromDomain(domain, client.LibvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain description : %s", err.Error()))
	}
	hostSizingV1, err := getSizingV1FromDomain(domain, client.LibvirtService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain sizing : %s", err.Error()))
	}
	hostNetworkV1, err := client.getNetworkV1FromDomain(domain)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get domain networks: %s", err.Error()))
	}

	host := model.NewHost()

	host.ID = id
	host.Name = name
	host.PrivateKey = "Impossible to fetch them from the domain, the private key is unknown by the domain for security reasons"
	host.LastState = stateConvert(state)
	host.Properties.Set(HostProperty.DescriptionV1, hostDescriptionV1)
	host.Properties.Set(HostProperty.SizingV1, hostSizingV1)
	host.Properties.Set(HostProperty.NetworkV1, hostNetworkV1)

	return host, nil
}

// getHostAndDomainFromRef retrieve the host and the domain associated to an ref (id or name)
func (client *Client) getHostAndDomainFromRef(ref string) (*model.Host, *libvirt.Domain, error) {
	domain, err := client.LibvirtService.LookupDomainByUUIDString(ref)
	if err != nil {
		domain, err = client.LibvirtService.LookupDomainByName(ref)
		if err != nil {
			return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to fetch domain from ref : %s", err.Error()))
		}
	}

	host, err := client.getHostFromDomain(domain)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get host from domain : %s", err.Error())
	}

	return host, domain, nil
}

// CreateHost creates an host satisfying request
func (client *Client) CreateHost(request model.HostRequest) (*model.Host, error) {
	resourceName := request.ResourceName
	hostName := request.HostName
	networks := request.Networks
	publicIP := request.PublicIP
	templateID := request.TemplateID
	imageID := request.ImageID
	keyPair := request.KeyPair
	defaultGateway := request.DefaultGateway

	//----Check Inputs----
	if resourceName == "" {
		return nil, fmt.Errorf("The ResourceName is mandatory ")
	}
	if hostName == "" {
		hostName = resourceName
	}
	if networks == nil || len(networks) == 0 {
		return nil, fmt.Errorf("The host %s must be on at least one network (even if public)", resourceName)
	}
	if defaultGateway == nil && !publicIP {
		return nil, fmt.Errorf("The host %s must have a gateway or be public", resourceName)
	}
	if templateID == "" {
		return nil, fmt.Errorf("The TemplateID is mandatory", resourceName)
	}
	if imageID == "" {
		return nil, fmt.Errorf("The ImageID is mandatory", resourceName)
	}
	host, _, err := client.getHostAndDomainFromRef(resourceName)
	if err == nil && host != nil {
		return nil, fmt.Errorf("The Host %s already exists", resourceName)
	}

	//----Initialize----
	if keyPair == nil {
		var err error
		keyPair, err = client.CreateKeyPair(fmt.Sprintf("key_%s", resourceName))
		if err != nil {
			return nil, fmt.Errorf("KeyPair creation failed : ", err.Error())
		}
	}
	template, err := client.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("GetTemplate failed : ", err.Error())
	}
	imagePath, err := getImagePathFromID(imageID)
	if err != nil {
		return nil, fmt.Errorf("GetImageFromPath failled : ", err.Error())
	}

	userData, err := userdata.Prepare(client, request, keyPair, networks[0].CIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare user data content: %+v", err)
	}
	err = ioutil.WriteFile(libvirtStorage+"/"+resourceName+"_userdata.sh", userData, 0644)
	if err != nil {
		return nil, fmt.Errorf("Failed to write userData in %s_userdata.sh file : %s", resourceName, err.Error())
	}

	//----Commands----
	networksCommandString := ""
	for _, network := range networks {
		networksCommandString += fmt.Sprintf(" --network network=%s", network.Name)
	}
	if publicIP {
		networksCommandString += fmt.Sprintf(" --network type=direct,source=%s,source_mode=bridge", client.Config.LanInterface)
	}

	// without sudo rights /boot/vmlinuz/`uname -r` have to be readable by the user to execute virt-resize / virt-sysprep
	// TODO gpu is ignored
	// TODO use libvirt-go functions not bash commands
	command_setup := fmt.Sprintf("IMAGE_PATH=\"%s\" && IMAGE=\"`echo $IMAGE_PATH | rev | cut -d/ -f1 | rev`\" && EXT=\"`echo $IMAGE | grep -o '[^.]*$'`\" && LIBVIRT_STORAGE=\"%s\" && HOST_NAME=\"%s\" && VM_IMAGE=\"$LIBVIRT_STORAGE/$HOST_NAME.$EXT\"", imagePath, libvirtStorage, resourceName)
	command_copy := fmt.Sprintf("cd $LIBVIRT_STORAGE && cp $IMAGE_PATH . && chmod 666 $IMAGE")
	command_resize := fmt.Sprintf("truncate $VM_IMAGE -s %dG && virt-resize --expand /dev/sda1 $IMAGE $VM_IMAGE && rm $IMAGE", template.DiskSize)
	command_sysprep := fmt.Sprintf("virt-sysprep -a $VM_IMAGE --hostname %s --operations all,-ssh-hostkeys --firstboot %s_userdata.sh && rm %s_userdata.sh", hostName, resourceName, resourceName)
	command_virt_install := fmt.Sprintf("virt-install --name=%s --vcpus=%d --memory=%d --import --disk=$VM_IMAGE %s --noautoconsole", resourceName, template.Cores, int(template.RAMSize*1024), networksCommandString)
	command := strings.Join([]string{command_setup, command_copy, command_resize, command_sysprep, command_virt_install}, " && ")

	cmd := exec.Command("bash", "-c", command)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Commands failled : \n", command, "\n", err.Error())
	}

	defer func() {
		if err != nil {
			client.DeleteHost(resourceName)
		}
	}()

	//----Generate model.Host----
	domain, err := client.LibvirtService.LookupDomainByName(resourceName)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Can't find domain %s : %s", resourceName, err.Error()))
	}

	host, err = client.getHostFromDomain(domain)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to get host %s from domain : %s", resourceName, err.Error()))
	}

	host.PrivateKey = keyPair.PrivateKey

	hostNetworkV1 := propsv1.NewHostNetwork()
	host.Properties.Get(HostProperty.NetworkV1, hostNetworkV1)

	hostNetworkV1.DefaultNetworkID = request.Networks[0].ID
	hostNetworkV1.IsGateway = request.DefaultGateway == nil && request.Networks[0].Name != model.SingleHostNetworkName
	if request.DefaultGateway != nil {
		hostNetworkV1.DefaultGatewayID = request.DefaultGateway.ID

		gateway, err := client.GetHost(request.DefaultGateway)
		if err != nil {
			return nil, fmt.Errorf("Failed to get gateway host : %s", err.Error())
		}

		hostNetworkV1.DefaultGatewayPrivateIP = gateway.GetPrivateIP()
	}

	hostSizingV1 := propsv1.NewHostSizing()
	host.Properties.Get(HostProperty.SizingV1, hostSizingV1)

	hostSizingV1.RequestedSize.RAMSize = float32(template.RAMSize * 1024)
	hostSizingV1.RequestedSize.Cores = template.Cores
	hostSizingV1.RequestedSize.DiskSize = template.DiskSize
	// TODO GPU not implemented
	hostSizingV1.RequestedSize.GPUNumber = template.GPUNumber
	hostSizingV1.RequestedSize.GPUType = template.GPUType

	host.Properties.Set(HostProperty.NetworkV1, hostNetworkV1)
	host.Properties.Set(HostProperty.SizingV1, hostSizingV1)

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

	host, _, err := client.getHostAndDomainFromRef(ref)
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
	_, domain, err := client.getHostAndDomainFromRef(id)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	volumes, err := getVolumesFromDomain(domain, client.LibvirtService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to get the volumes from the domain : %s", err.Error()))
	}

	err = domain.Destroy()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to destroy the domain : %s", err.Error()))
	}
	err = domain.Undefine()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to undefine the domain : %s", err.Error()))
	}

	for _, volume := range volumes {
		volumePath := volume.Key
		pathSplitted := strings.Split(volumePath, "/")
		volumeName := strings.Split(pathSplitted[len(pathSplitted)-1], ".")[0]
		domainName, err := domain.GetName()
		if err != nil {
			return fmt.Errorf("Failed to get domain name : %s", err.Error())
		}
		if domainName == volumeName {
			err := os.Remove(volumePath)
			if err != nil {
				return fmt.Errorf("Failed to delete volume %s : %s", volumeName, err.Error())
			}
		}
	}

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
		host, err := client.getHostFromDomain(&domain)
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Failed to get host from domain : %s", err.Error()))
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}

// StopHost stops the host identified by id
func (client *Client) StopHost(id string) error {
	_, domain, err := client.getHostAndDomainFromRef(id)
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
	_, domain, err := client.getHostAndDomainFromRef(id)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Create()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to launch the host : %s", err.Error()))
	}

	//TODO wait domain to be fully operational?

	return nil
}

// RebootHost reboot the host identified by id
func (client *Client) RebootHost(id string) error {
	_, domain, err := client.getHostAndDomainFromRef(id)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("getHostAndDomainFromRef failed : %s", err.Error()))
	}

	err = domain.Reboot(0)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to reboot the host : %s", err.Error()))
	}

	//TODO wait domain to be fully operational?

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
	return map[string]bool{"local": true}, nil
}
