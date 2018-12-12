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

// HostCmd command
var HostCmd = cli.Command{
	Name:  "host",
	Usage: "host COMMAND",
	Subcommands: []cli.Command{
		hostCreate,
		hostDelete,
		hostList,
		hostInspect,
		hostStart,
		hostStop,
		hostReboot,
		hostStatus,
		hostSsh,
	},
}

var hostCreate = cli.Command{
	Name:      "create",
	Aliases:   []string{"new"},
	Usage:     "create a new host",
	ArgsUsage: "<Host_name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "net,network",
			Value: "",
			Usage: "network name or network id",
		},
		cli.IntFlag{
			Name:  "cpu",
			Value: 1,
			Usage: "Number of CPU for the host",
		},
		cli.Float64Flag{
			Name:  "ram",
			Value: 1,
			Usage: "RAM for the host (GB)",
		},
		cli.IntFlag{
			Name:  "disk",
			Value: 16,
			Usage: "Disk space for the host (GB)",
		},
		cli.StringFlag{
			Name:  "os",
			Value: "Ubuntu 16.04",
			Usage: "Image name for the host",
		},
		cli.BoolFlag{
			Name:  "public",
			Usage: "Create with public IP",
		},
		cli.IntFlag{
			Name:  "gpu",
			Value: 0,
			Usage: "Number of GPU for the host",
		},
		cli.Float64Flag{
			Name:  "cpu-freq, cpufreq",
			Value: 0,
			Usage: "Minimum cpu frequency required for the host (GHz)",
		},
		cli.BoolFlag{
			Name:  "f, force",
			Usage: "Force creation even if the host doesn't meet the GPU and CPU freq requirements",
		},
	},
	Action: func(c *cli.Context) error {

		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

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
			MinGPU:      c.Int("gpu"),
			MinFreq:     float32(c.Float64("cpufreq")),
		}
		template, err := SelectTemplateBySize(sizingRequirements, templates)
		if err != nil {
			return fmt.Errorf("Failed to select template by size : %s", err.Error())
		}

		networkName := c.String("network")
		if networkName == "" {
			net, err := client.GetNetwork("net-safescale")
			if err != nil || net == nil {
				networkRequest := model.NetworkRequest{
					Name:       "net-safescale",
					IPVersion:  IPVersion.IPv4,
					CIDR:       "10.0.0.0/24",
					DNSServers: []string{},
				}
				net, err = client.CreateNetwork(networkRequest)
				if err != nil {
					return fmt.Errorf("Failed to crete default network : %s", err.Error())
				}
			}
			networkName = net.Name
		}
		mNetwork, err := metadata.LoadNetwork(client, networkName)
		if err != nil || mNetwork == nil {
			return fmt.Errorf("Failed to load network '%s' metadatas", networkName)
		}
		network := mNetwork.Get()

		mGw, err := metadata.LoadHost(client, network.GatewayID)
		if err != nil || mGw == nil {
			return fmt.Errorf("Failed to load host '%s'  gateway metadatas", networkName)
		}
		gw := mGw.Get()

		hostRequest := model.HostRequest{
			ResourceName:   c.Args().First(),
			PublicIP:       c.Bool("public"),
			Networks:       []*model.Network{network},
			DefaultGateway: gw,
			TemplateID:     template.ID,
			ImageID:        image.ID,
		}

		host, err := client.CreateHost(hostRequest)
		if err != nil {
			return fmt.Errorf("Failed to create Host : %s", err.Error())
		}

		// TODO test if SSH connection available

		err = metadata.SaveHost(client, host)
		if err != nil {
			return fmt.Errorf("Failed to save host metadata into object storage : %s", err.Error())
		}

		return nil
	},
}

var hostDelete = cli.Command{
	Name:      "delete",
	Aliases:   []string{"rm", "remove"},
	Usage:     "Delete host",
	ArgsUsage: "<Host_name|Host_ID> [<Host_name|Host_ID>...]",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		var hostList []string
		hostList = append(hostList, c.Args().First())
		hostList = append(hostList, c.Args().Tail()...)

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		for _, hostName := range hostList {
			mHost, err := metadata.LoadHost(client, hostName)
			if err != nil || mHost == nil {
				return fmt.Errorf("Host '%s' not found in metadatas", hostName)
			}

			err = client.DeleteHost(hostName)
			if err != nil {
				return fmt.Errorf("Failed to delete '%s' host : %s", hostName, err.Error())
			}
			fmt.Println(fmt.Sprintf("Host '%s' sucessfully deleted", hostName))

			err = metadata.RemoveHost(client, mHost.Get())
			if err != nil {
				return fmt.Errorf("Failed to remove host '%s' from metadatas : %s", hostName, err.Error())
			}
		}

		// TODO check if a host is ready to be destroyed (no volumes ...)

		return nil
	},
}

var hostList = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List available hosts",
	Flags:   []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		hosts, err := client.ListHosts()
		if err != nil {
			return fmt.Errorf("Failed to list hosts : %s", err.Error())
		}

		for _, host := range hosts {
			displayHost(host)
		}

		return nil
	},
}

var hostInspect = cli.Command{
	Name:      "inspect",
	Aliases:   []string{"show"},
	Usage:     "inspect Host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		host, err := client.GetHost(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to inspect host '%s' : %s", c.Args().First(), err.Error())
		}
		displayHost(host)

		return nil
	},
}

var hostStart = cli.Command{
	Name:      "start",
	Usage:     "start Host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}
		err = client.StartHost(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to start the host : %s", err.Error())
		}

		fmt.Printf("Host '%s' successfully started.\n", c.Args().First())
		return nil
	},
}

var hostStop = cli.Command{
	Name:      "stop",
	Usage:     "stop Host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}
		err = client.StopHost(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to stop the host : %s", err.Error())
		}

		fmt.Printf("Host '%s' successfully stopped.\n", c.Args().First())
		return nil
	},
}

var hostReboot = cli.Command{
	Name:      "reboot",
	Usage:     "reboot Host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}
		err = client.RebootHost(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to reboot the host : %s", err.Error())
		}

		fmt.Printf("Host '%s' successfully rebooted.\n", c.Args().First())
		return nil
	},
}

var hostStatus = cli.Command{
	Name:      "status",
	Usage:     "status Host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		hostState, err := client.GetHostState(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get host '%s' state : %s", c.Args().First(), err.Error())
		}
		fmt.Println(fmt.Sprintf("Host '%s' is in state : %s", c.Args().First(), hostState))

		return nil
	},
}

var hostSsh = cli.Command{
	Name:      "ssh",
	Usage:     "Get ssh config to connect to host",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		sshConfig, err := GetSSHConfigFromHostName(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get a sshConfig : %s", err.Error())
		}

		DisplaySSHConfig(sshConfig)

		return nil
	},
}

func displayHost(host *model.Host) {
	fmt.Println(host)
}
