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
	"github.com/CS-SI/LocalDriver/model/enums/HostProperty"
	propsv1 "github.com/CS-SI/LocalDriver/model/properties/v1"
	"github.com/CS-SI/LocalDriver/system"
	"github.com/urfave/cli"
)

// SSHCmd ssh command
var SSHCmd = cli.Command{
	Name:  "ssh",
	Usage: "ssh COMMAND",
	Subcommands: []cli.Command{
		sshConnect,
		//sshRun,
		//sshCopy,
		//sshTunnel,
		//sshClose,
	},
}

var sshConnect = cli.Command{
	Name:      "connect",
	Usage:     "Connect to the host with interactive shell",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		mHost, err := metadata.LoadHost(client, c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get host '%s' metadatas : %s", c.Args().First(), err.Error())
		}
		host := mHost.Get()

		sshConfig := system.SSHConfig{
			PrivateKey: host.PrivateKey,
			Port:       22,
			Host:       host.GetAccessIP(),
			User:       model.DefaultUser,
		}

		hostNetworkV1 := propsv1.NewHostNetwork()
		err = host.Properties.Get(HostProperty.NetworkV1, hostNetworkV1)
		if err != nil {
			return fmt.Errorf("Failed to get host '%s' network Properties : %s", c.Args().First(), err.Error())
		}
		if hostNetworkV1.DefaultGatewayID != "" {
			mGw, err := metadata.LoadHost(client, hostNetworkV1.DefaultGatewayID)
			if err != nil {
				return fmt.Errorf("Failed to get host '%s' gateway metadatas : %s", c.Args().First(), err.Error())
			}
			gw := mGw.Get()

			GatewayConfig := system.SSHConfig{
				PrivateKey: gw.PrivateKey,
				Port:       22,
				Host:       gw.GetAccessIP(),
				User:       model.DefaultUser,
			}
			sshConfig.GatewayConfig = &GatewayConfig
		}

		sshConfig.Enter()

		return nil
	},
}

// var sshRun = cli.Command{
// 	Name:      "run",
// 	Usage:     "Run a command on the host",
// 	ArgsUsage: "<Host_name|Host_ID>",
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "c",
// 			Usage: "Command to execute",
// 		},
// 		cli.StringFlag{
// 			Name:  "timeout",
// 			Value: "5",
// 			Usage: "timeout in minutes",
// 		}},
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 1 {
// 			fmt.Println("Missing mandatory argument <Host_name>")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return clitools.ExitOnInvalidArgument()
// 		}
// 		timeout := brokerutils.TimeoutCtxHost
// 		if c.IsSet("timeout") {
// 			timeout = time.Duration(c.Float64("timeout")) * time.Minute
// 		}
// 		retcode, stdout, stderr, err := client.New().Ssh.Run(c.Args().Get(0), c.String("c"), client.DefaultConnectionTimeout, timeout)
// 		if err != nil {
// 			return clitools.ExitOnRPC(utils.TitleFirst(client.DecorateError(err, "ssh run", false).Error()))
// 		}

// 		fmt.Println(stdout)
// 		fmt.Fprintln(os.Stderr, stderr)

// 		if retcode != 0 {
// 			return cli.NewExitError("", retcode)
// 		}
// 		return nil
// 	},
// }

// func normalizeFileName(fileName string) string {
// 	absPath, _ := filepath.Abs(fileName)
// 	if _, err := os.Stat(absPath); err != nil {
// 		return fileName
// 	}
// 	return absPath
// }

// var sshCopy = cli.Command{
// 	Name:      "copy",
// 	Usage:     "Copy a local file/directory to an host or copy from host to local",
// 	ArgsUsage: "from to  Ex: /my/local/file.txt host1:/remote/path/",
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "timeout",
// 			Value: "5",
// 			Usage: "timeout in minutes",
// 		}},
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 2 {
// 			fmt.Println("2 arguments (from and to) are required")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return clitools.ExitOnInvalidArgument()
// 		}
// 		timeout := brokerutils.TimeoutCtxHost
// 		if c.IsSet("timeout") {
// 			timeout = time.Duration(c.Float64("timeout")) * time.Minute
// 		}
// 		_, _, _, err := client.New().Ssh.Copy(normalizeFileName(c.Args().Get(0)), normalizeFileName(c.Args().Get(1)), client.DefaultConnectionTimeout, timeout)
// 		if err != nil {
// 			return clitools.ExitOnRPC(utils.TitleFirst(client.DecorateError(err, "ssh copy", true).Error()))
// 		}
// 		fmt.Printf("Copy of '%s' to '%s' done\n", c.Args().Get(0), c.Args().Get(1))
// 		return nil
// 	},
// }

// var sshTunnel = cli.Command{
// 	Name:      "tunnel",
// 	Usage:     "Create a ssh tunnel between admin host and a host in the cloud",
// 	ArgsUsage: "<Host_name|Host_ID --local local_port  --remote remote_port>",
// 	Flags: []cli.Flag{
// 		cli.IntFlag{
// 			Name:  "local",
// 			Value: 8080,
// 			Usage: "local tunnel's port, if not set all",
// 		},
// 		cli.IntFlag{
// 			Name:  "remote",
// 			Value: 8080,
// 			Usage: "remote tunnel's port, if not set all",
// 		},
// 		cli.StringFlag{
// 			Name:  "timeout",
// 			Value: "1",
// 			Usage: "timeout in minutes",
// 		},
// 	},
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 1 {
// 			fmt.Println("Missing mandatory argument")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return fmt.Errorf("Missing arguments")
// 		}

// 		localPort := c.Int("local")
// 		if 0 > localPort || localPort > 65535 {
// 			fmt.Printf("%d is not a valid port\n", localPort)
// 			_ = cli.ShowSubcommandHelp(c)
// 			return fmt.Errorf("wrong value of localport")
// 		}

// 		remotePort := c.Int("remote")
// 		if 0 > localPort || localPort > 65535 {
// 			fmt.Printf("%d is not a valid port\n", remotePort)
// 			_ = cli.ShowSubcommandHelp(c)
// 			return fmt.Errorf("wrong value of remoteport")
// 		}

// 		timeout := time.Duration(c.Float64("timeout")) * time.Minute

// 		//c.GlobalInt("port") is the grpc port aka. 50051
// 		err := client.New().Ssh.CreateTunnel(c.Args().Get(0), localPort, remotePort, timeout)
// 		if err != nil {
// 			err = client.DecorateError(err, "ssh tunnel", false)
// 		}
// 		return err
// 	},
// }

// var sshClose = cli.Command{
// 	Name:      "close",
// 	Usage:     "Close one or several ssh tunnel",
// 	ArgsUsage: "<Host_name|Host_ID> --local local_port --remote remote_port",
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "local",
// 			Value: ".*",
// 			Usage: "local tunnel's port, if not set all",
// 		},
// 		cli.StringFlag{
// 			Name:  "remote",
// 			Value: ".*",
// 			Usage: "remote tunnel's port, if not set all",
// 		},
// 		cli.StringFlag{
// 			Name:  "timeout",
// 			Value: "1",
// 			Usage: "timeout in minutes",
// 		},
// 	},
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 1 {
// 			fmt.Println("Missing mandatory argument")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return fmt.Errorf("Missing arguments")
// 		}

// 		strLocalPort := c.String("local")
// 		if c.IsSet("local") {
// 			localPort, err := strconv.Atoi(strLocalPort)
// 			if err != nil || 0 > localPort || localPort > 65535 {
// 				fmt.Printf("%d is not a valid port\n", localPort)
// 				_ = cli.ShowSubcommandHelp(c)
// 				return fmt.Errorf("wrong value of localport")
// 			}
// 		}

// 		strRemotePort := c.String("remote")
// 		if c.IsSet("remote") {
// 			remotePort, err := strconv.Atoi(strRemotePort)
// 			if err != nil || 0 > remotePort || remotePort > 65535 {
// 				fmt.Printf("%d is not a valid port\n", remotePort)
// 				_ = cli.ShowSubcommandHelp(c)
// 				return fmt.Errorf("wrong value of remoteport")
// 			}
// 		}

// 		timeout := time.Duration(c.Float64("timeout")) * time.Minute

// 		//c.GlobalInt("port") is the grpc port aka. 50051
// 		err := client.New().Ssh.CloseTunnels(c.Args().Get(0), strLocalPort, strRemotePort, timeout)
// 		if err != nil {
// 			err = client.DecorateError(err, "ssh close", false)
// 		}
// 		return err
// 	},
// }
