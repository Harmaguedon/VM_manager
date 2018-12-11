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

package cmd

import (
	"fmt"

	"github.com/CS-SI/LocalDriver/metadata"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/IPVersion"

	"github.com/urfave/cli"
)

// NetworkCmd command
var NetworkCmd = cli.Command{
	Name:  "network",
	Usage: "network COMMAND",
	Subcommands: []cli.Command{
		networkCreate,
		networkDelete,
		networkList,
		networkInspect,
	},
}

//TODO use metadata to get info on nets

var networkCreate = cli.Command{
	Name:      "create",
	Aliases:   []string{"new"},
	Usage:     "create a network",
	ArgsUsage: "<network_name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "cidr",
			Value: "192.168.0.0/24",
			Usage: "cidr of the network",
		},
		cli.IntFlag{
			Name:  "cpu",
			Value: 1,
			Usage: "Number of CPU for the gateway",
		},
		cli.Float64Flag{
			Name:  "ram",
			Value: 1,
			Usage: "RAM for the gateway",
		},
		cli.IntFlag{
			Name:  "disk",
			Value: 16,
			Usage: "Disk space for the gateway",
		},
		cli.StringFlag{
			Name:  "os",
			Value: "Ubuntu 16.04",
			Usage: "Image name for the gateway",
		},
		cli.StringFlag{
			Name:  "gwname",
			Value: "",
			Usage: "Name for the gateway. Default to 'gw-<network_name>'",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Network_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		networkRequest := model.NetworkRequest{
			Name:       c.Args().First(),
			IPVersion:  IPVersion.IPv4,
			CIDR:       c.String("cidr"),
			DNSServers: []string{},
		}
		network, err := client.CreateNetwork(networkRequest)
		if err != nil {
			return fmt.Errorf("Create network failed : %s", err.Error())

		}

		displayNetwork(network)

		image, err := client.GetImage(c.String("os"))
		if err != nil {
			return fmt.Errorf("Failed to get the image : %s", err.Error())
		}
		templates, err := client.ListTemplates(true)
		if err != nil {
			return fmt.Errorf("Failed to get the templates : %s", err.Error())
		}
		sizingRequirements := model.SizingRequirements{
			MinCores:    c.Int("cpu"),
			MinRAMSize:  float32(c.Float64("ram")),
			MinDiskSize: c.Int("disk"),
		}
		template, err := SelectTemplateBySize(sizingRequirements, templates)
		if err != nil {
			return fmt.Errorf("Failed to select template by size : %s", err.Error())
		}

		gwRequest := model.GatewayRequest{
			Network:    network,
			CIDR:       c.String("cidr"),
			TemplateID: template.ID,
			ImageID:    image.ID,
			KeyPair:    nil,
			Name:       "",
		}

		gw, err := client.CreateGateway(gwRequest)
		if err != nil {
			return fmt.Errorf("Failed to create Gateway : %s", err.Error())
		}

		// TODO test if SSH connection available

		err = metadata.SaveHost(client, gw)
		if err != nil {
			return fmt.Errorf("Failed to save gateway metadata into object storage : %s", err.Error())
		}

		network.GatewayID = gw.ID

		metadata.SaveNetwork(client, network)

		return nil
	},
}

var networkDelete = cli.Command{
	Name:      "delete",
	Aliases:   []string{"rm", "remove"},
	Usage:     "delete Network",
	ArgsUsage: "<Network_name> [<Network_name>...]",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Network_name>")
		}

		var networkList []string
		networkList = append(networkList, c.Args().First())
		networkList = append(networkList, c.Args().Tail()...)

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		for _, networkName := range networkList {
			mNetwork, err := metadata.LoadNetwork(client, networkName)
			if err != nil {
				return fmt.Errorf("Network '%s' not found in metadatas %s", networkName, err.Error())
			}
			network := mNetwork.Get()
			mGW, err := metadata.LoadHost(client, network.GatewayID)
			if err != nil {
				return fmt.Errorf("Network '%s' not found in metadatas %s", networkName, err.Error())
			}
			gw := mGW.Get()

			err = client.DeleteHost(gw.ID)
			if err != nil {
				return fmt.Errorf("Failed to delete '%s' gateway : %s", gw.Name, err.Error())
			}
			fmt.Println(fmt.Sprintf("Gateway '%s' sucessfully deleted", networkName))
			err = metadata.RemoveHost(client, gw)
			if err != nil {
				return fmt.Errorf("Failed to remove gateway '%s' from metadatas : %s", gw.Name, err.Error())
			}

			err = client.DeleteNetwork(networkName)
			if err != nil {
				return fmt.Errorf("Failed to delete '%s' network : %s", networkName, err.Error())
			}
			fmt.Println(fmt.Sprintf("Network '%s' sucessfully deleted", networkName))
			err = metadata.RemoveNetwork(client, network)
			if err != nil {
				return fmt.Errorf("Failed to remove network '%s' from metadatas : %s", networkName, err.Error())
			}
		}

		//TODO Check that the network is empty

		return nil
	},
}

var networkList = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List existing Networks",
	Flags:   []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		networks, err := client.ListNetworks()
		if err != nil {
			return fmt.Errorf("Failed to list networks : %s", err.Error())
		}

		for _, network := range networks {
			displayNetwork(network)
		}

		return nil
	},
}

var networkInspect = cli.Command{
	Name:      "inspect",
	Aliases:   []string{"show"},
	Usage:     "inspect NETWORK",
	ArgsUsage: "<network_name>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Network_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		network, err := client.GetNetwork(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to list networks : %s", err.Error())
		}

		displayNetwork(network)

		return nil
	},
}

func displayNetwork(network *model.Network) {
	fmt.Println(network)
}
