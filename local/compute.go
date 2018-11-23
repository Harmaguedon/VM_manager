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
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"os/exec"

	propsv1 "github.com/CS-SI/SafeScale/providers/model/properties/v1"
	"github.com/CS-SI/LocalDriver/model"

	"golang.org/x/crypto/ssh"
)

const imagesJsonPath string = "/home/armand/Iso/images.json"
const templatesJsonPath string = "/home/armand/Iso/templates.json"

// ListImages lists available OS images
func (client *Client) ListImages(all bool) ([]model.Image, error) {
	//TODO
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

// ListTemplates overload OpenStack ListTemplate method to filter wind and flex instance and add GPU configuration
func (client *Client) ListTemplates(all bool) ([]model.HostTemplate, error) {
	//TODO
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

// CreateHost creates an host satisfying request
func (client *Client) CreateHost(request model.HostRequest) (*model.Host, error) {
	resourceName	:= request.ResourceName
	hostName 		:= request.HostName
	networkID		:= request.NetworkIDs
	publicIP		:= request.PublicIP
	templateID		:= request.TemplateID
	imageID			:= request.ImageID
	keyPair 		:= request.KeyPair

	//TODO check if ressource name is already used

	if resourceName == "" {
		return nil, fmt.Errorf("The ResourceName is mandatory", resourceName)
	}
	if hostName == "" {
		hostName = resourceName
	}
	if ( networkID == nil || len(networkID) == 0 ) && ! publicIP {
		return nil, fmt.Errorf("The host must be on a network or be public", resourceName)
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
	image, err		:= client.GetImage(imageID)
	if err != nil {
		return nil, fmt.Errorf("GetImage failled : ", err.Error())
	}
	imagePath, err	:= getImagePathFromID(imageID)
	if err != nil {
		return nil, fmt.Errorf("GetImageFromPath failled : ", err.Error())
	}

	// TODO gpu is ignored
	command := fmt.Sprintf("virt-install --name=%s_%s --vcpus=%d --memory=%d --disk size=%d --cdrom=%s  --print-xml", resourceName, image.Name, template.Cores, int(template.RAMSize*1024), template.DiskSize, imagePath)

	fmt.Println(command)

	cmd := exec.Command("bash", "-c", command)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command failled : ", err.Error())
	}
	fmt.Print(string(cmdOutput.Bytes()))

	//domain, err := client.conn.DomainCreateXML(xmldescription, 0)
	//if err != nil {
	//	return nil, fmt.Errorf("Failed to create host : %s", err.Error())
	//}

	return nil, nil
}

//// WaitHostReady waits an host achieve ready state
//func (client *Client) WaitHostReady(hostID string, timeout time.Duration) (*model.Host, error) {
//	return nil, nil
//}
//
//// UpdateHost returns the host identified by id
//func (client *Client) UpdateHost(host *model.Host) error {
//	return nil
//}
//
//// GetHostState returns the host identified by id
//func (client *Client) GetHostState(hostParam interface{}) (HostState.Enum, error) {
//	return nil, nil
//}
//
//// ListHosts lists available hosts
//func (client *Client) ListHosts(all bool) ([]*model.Host, error) {
//	return nil, nil
//}
//
//// DeleteHost deletes the host identified by id
//func (client *Client) DeleteHost(ref string) error {
//	return nil
//}
//
//// StopHost stops the host identified by id
//func (client *Client) StopHost(id string) error {
//	return nil
//}
//
//// RebootHost ...
//func (client *Client) RebootHost(id string) error {
//	log.Println("Received reboot petition OVH")
//	return nil
//}
//
//// StartHost starts the host identified by id
//func (client *Client) StartHost(id string) error {
//	return nil
//}

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