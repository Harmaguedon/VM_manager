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

package metadata

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/CS-SI/LocalDriver/api"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/NetworkProperty"
	propsv1 "github.com/CS-SI/LocalDriver/model/properties/v1"
	"github.com/CS-SI/LocalDriver/utils/metadata"
)

const (
	//NetworkFolderName is the technical name of the container used to store networks info
	networksFolderName = "networks"
	// //GatewayObjectName is the name of the object containing the id of the host acting as a default gateway for a network
	// gatewayObjectName = "gw"
)

// Network links Object Storage folder and Network
type Network struct {
	item *metadata.Item
	//inside *metadata.Folder
	name *string
	id   *string
}

// NewNetwork creates an instance of network.Metadata
func NewNetwork(svc api.ClientAPI) *Network {
	return &Network{
		item: metadata.NewItem(svc, networksFolderName),
	}
}

// GetPath returns the path in Object Storage where the item is stored
func (m *Network) GetPath() string {
	if m.item == nil {
		panic("m.item is nil!")
	}
	return m.item.GetPath()
}

// Carry links a Network instance to the Metadata instance
func (m *Network) Carry(network *model.Network) *Network {
	if network == nil {
		panic("network parameter is nil!")
	}
	if m.item == nil {
		panic("m.item is nil!")
	}
	if network.Properties != nil {
		network.Properties = model.NewExtensions()
	}
	m.item.Carry(network)
	m.id = &network.ID
	m.name = &network.Name
	//m.inside = metadata.NewFolder(m.item.GetService(), strings.Trim(m.item.GetPath()+"/"+*m.id, "/"))
	return m
}

// Get returns the model.Network instance linked to metadata
func (m *Network) Get() *model.Network {
	if m.item == nil {
		panic("m.item is nil!")
	}
	return m.item.Get().(*model.Network)
}

// Write updates the metadata corresponding to the network in the Object Storage
func (m *Network) Write() error {
	if m.item == nil {
		panic("m.item is nil!")
	}
	err := m.item.WriteInto(ByIDFolderName, *m.id)
	if err != nil {
		return err
	}
	return m.item.WriteInto(ByNameFolderName, *m.name)
}

// Reload reloads the content of the Object Storage, overriding what is in the metadata instance
func (m *Network) Reload() error {
	found, err := m.ReadByID(*m.id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("the metadata of Network '%s' vanished", *m.name)
	}
	return nil
}

// ReadByID reads the metadata of a network identified by ID from Object Storage
func (m *Network) ReadByID(id string) (bool, error) {
	if m.item == nil {
		panic("m.item is nil!")
	}

	var network model.Network
	found, err := m.item.ReadFrom(ByIDFolderName, id, func(buf []byte) (model.Serializable, error) {
		err := (&network).Deserialize(buf)
		if err != nil {
			return nil, err
		}
		return &network, nil
	})
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	m.id = &network.ID
	m.name = &network.Name
	// m.inside = metadata.NewFolder(m.item.GetService(), strings.Trim(m.item.GetPath()+"/"+id, "/"))
	return true, nil
}

// ReadByName reads the metadata of a network identified by name
func (m *Network) ReadByName(name string) (bool, error) {
	if m.item == nil {
		panic("m.item is nil!")
	}

	var network model.Network
	found, err := m.item.ReadFrom(ByNameFolderName, name, func(buf []byte) (model.Serializable, error) {
		err := (&network).Deserialize(buf)
		if err != nil {
			return nil, err
		}
		return &network, nil
	})
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	m.name = &network.Name
	m.id = &network.ID
	//	m.inside = metadata.NewFolder(m.item.GetService(), strings.Trim(m.item.GetPath()+"/"+*m.id, "/"))
	return true, nil
}

// Delete deletes the metadata corresponding to the network
func (m *Network) Delete() error {
	if m.item == nil {
		panic("m.item is nil!")
	}

	// Delete the entry in 'ByIDFolderName' folder
	err := m.item.DeleteFrom(ByIDFolderName, *m.id)
	if err != nil {
		return err
	}
	// Delete the entry in 'ByNameFolderName' folder
	return m.item.DeleteFrom(ByNameFolderName, *m.name)
}

// Browse walks through all the metadata objects in network
func (m *Network) Browse(callback func(*model.Network) error) error {
	if m.item == nil {
		panic("m.item is nil!")
	}

	return m.item.BrowseInto(ByIDFolderName, func(buf []byte) error {
		network := model.Network{}
		err := (&network).Deserialize(buf)
		if err != nil {
			return err
		}
		return callback(&network)
	})
}

// // attachGateway register an host metadata as the Gateway for the network
// func (m *Network) attachGateway(mh *Host) error {
// 	if m.inside == nil {
// 		panic("m.inside is nil!")
// 	}
// 	data, err := mh.Get().Serialize()
// 	if err != nil {
// 		return err
// 	}
// 	return m.inside.Write(".", gatewayObjectName, data)
// }

// // getGateway returns the host acting as a gateway for the network
// func (m *Network) getGateway() (bool, *model.Host, error) {
// 	if m.inside == nil {
// 		panic("m.inside is nil!")
// 	}
// 	var host model.Host
// 	found, err := m.inside.Read(".", gatewayObjectName, func(buf []byte) error {
// 		return (&host).Deserialize(buf)
// 	})
// 	if err != nil {
// 		return false, nil, err
// 	}
// 	if !found {
// 		return false, nil, fmt.Errorf("failed to find gateway metadata")
// 	}
// 	return true, &host, nil
// }

// // detachGateway detaches the host used as gateway of the network
// func (m *Network) detachGateway() error {
// 	if m.inside == nil {
// 		panic("m.inside is nil")
// 	}

// 	err := m.inside.Delete(".", gatewayObjectName)
// 	if err != nil {
// 		log.Errorf("Error detaching gateway: deleting host folder: %+v", err)
// 	}

// 	return err
// }

// AttachHost links host ID to the network
func (m *Network) AttachHost(host *model.Host) error {
	network := m.Get()
	networkHostsV1 := propsv1.NewNetworkHosts()
	err := network.Properties.Get(NetworkProperty.HostsV1, networkHostsV1)
	if err != nil {
		return err
	}
	networkHostsV1.ByID[host.ID] = host.Name
	networkHostsV1.ByName[host.Name] = host.ID
	return network.Properties.Set(NetworkProperty.HostsV1, networkHostsV1)
}

// DetachHost unlinks host ID to network
func (m *Network) DetachHost(hostID string) error {
	network := m.Get()
	networkHostsV1 := propsv1.NewNetworkHosts()
	err := network.Properties.Get(NetworkProperty.HostsV1, networkHostsV1)
	if err != nil {
		return nil
	}
	hostName, found := networkHostsV1.ByID[hostID]
	if found {
		delete(networkHostsV1.ByName, hostName)
		delete(networkHostsV1.ByID, hostID)
	}
	return network.Properties.Set(NetworkProperty.HostsV1, networkHostsV1)
}

// ListHosts returns the list of model.Host attached to the network (not including gateway)
func (m *Network) ListHosts() ([]*model.Host, error) {
	network := m.Get()
	networkHostsV1 := propsv1.NewNetworkHosts()
	err := network.Properties.Get(NetworkProperty.HostsV1, networkHostsV1)
	if err != nil {
		return nil, err
	}
	var list []*model.Host
	for id := range networkHostsV1.ByID {
		mh, err := LoadHost(m.item.GetService(), id)
		if err != nil {
			return nil, err
		}
		list = append(list, mh.Get())
	}
	if err != nil {
		log.Errorf("Error listing hosts: %+v", err)
	}
	return list, nil
}

// Acquire waits until the write lock is available, then locks the metadata
func (m *Network) Acquire() {
	m.item.Acquire()
}

// Release unlocks the metadata
func (m *Network) Release() {
	m.item.Release()
}

// SaveNetwork saves the Network definition in Object Storage
func SaveNetwork(svc api.ClientAPI, net *model.Network) error {
	return NewNetwork(svc).Carry(net).Write()
}

// RemoveNetwork removes the Network definition from Object Storage
func RemoveNetwork(svc api.ClientAPI, net *model.Network) error {
	return NewNetwork(svc).Carry(net).Delete()
}

// LoadNetworkByID gets the Network definition from Object Storage
func LoadNetworkByID(svc api.ClientAPI, networkID string) (*Network, error) {
	m := NewNetwork(svc)
	found, err := m.ReadByID(networkID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return m, nil
}

// LoadNetworkByName gets the Network definition from Object Storage
func LoadNetworkByName(svc api.ClientAPI, networkname string) (*Network, error) {
	m := NewNetwork(svc)
	found, err := m.ReadByName(networkname)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return m, nil
}

// LoadNetwork gets the Network definition from Object Storage
func LoadNetwork(svc api.ClientAPI, ref string) (*Network, error) {
	m, err := LoadNetworkByID(svc, ref)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}

	m, err = LoadNetworkByName(svc, ref)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}
	return nil, nil
}

// Gateway links Object Storage folder and Network
type Gateway struct {
	host      *Host
	network   *Network
	networkID string
}

// NewGateway creates an instance of metadata.Gateway
func NewGateway(svc api.ClientAPI, networkID string) (*Gateway, error) {
	network := NewNetwork(svc)
	found, err := network.ReadByID(networkID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("failed to find metadata of network using gateway")
	}
	return &Gateway{
		host:      NewHost(svc),
		network:   network,
		networkID: networkID,
	}, nil
}

// Carry links a Network instance to the Metadata instance
func (mg *Gateway) Carry(host *model.Host) *Gateway {
	mg.host.Carry(host)
	return mg
}

// Get returns the *model.Host linked to the metadata
func (mg *Gateway) Get() *model.Host {
	if mg.host == nil {
		panic("mg.host is nil!")
	}
	return mg.host.Get()
}

// Write updates the metadata corresponding to the network in the Object Storage
// A Gateway is a particular host : we want it listed in hosts, but not listed as attached to the network
func (mg *Gateway) Write() error {
	if mg.host == nil {
		panic("m.item is nil!")
	}
	if mg.network == nil {
		panic("m.network is nil!")
	}
	return mg.host.Write()
}

// Read reads the metadata of a gateway of a network identified by ID from Object Storage
func (mg *Gateway) Read() (bool, error) {
	if mg.network == nil {
		panic("mg.network is nil!")
	}
	err := mg.network.Reload()
	if err != nil {
		return false, err
	}
	found := false
	found, err = mg.host.ReadByID(mg.network.Get().GatewayID)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return true, nil
}

// Reload reloads the content of the Object Storage, overriding what is in the metadata instance
func (mg *Gateway) Reload() error {
	found, err := mg.Read()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("metadata about the gateway of network '%s' doesn't exist anymore", mg.networkID)
	}
	return nil
}

// Delete updates the metadata of the network concerning the gateway
func (mg *Gateway) Delete() error {
	if mg.network == nil {
		panic("mg.network is nil!")
	}
	if mg.host == nil {
		panic("mg.host is nil!")
	}

	mg.network.Get().GatewayID = ""
	err := mg.network.Write()
	if err != nil {
		return err
	}
	return mg.host.Delete()
}

// Acquire waits until the write lock is available, then locks the metadata
func (mg *Gateway) Acquire() {
	mg.host.Acquire()
}

// Release unlocks the metadata
func (mg *Gateway) Release() {
	mg.host.Release()
}

// LoadGateway returns the metadata of the Gateway of a network
func LoadGateway(svc api.ClientAPI, networkID string) (*Gateway, error) {
	mg, err := NewGateway(svc, networkID)
	if err != nil {
		return nil, err
	}
	found, err := mg.Read()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return mg, nil
}

// SaveGateway saves the metadata of a gateway
func SaveGateway(svc api.ClientAPI, host *model.Host, networkID string) error {
	mg, err := NewGateway(svc, networkID)
	if err != nil {
		return err
	}

	// Update network
	mn := NewNetwork(svc)
	ok, err := mn.ReadByID(networkID)
	if !ok || err != nil {
		return fmt.Errorf("metadata about the network '%s' doesn't exist anymore", networkID)
	}
	network := mn.Get()
	network.GatewayID = host.ID
	err = mn.Write()
	if err != nil {
		return err
	}

	return mg.Carry(host).Write()
}
