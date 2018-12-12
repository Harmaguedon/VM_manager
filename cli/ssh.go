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
	"os"
	"path/filepath"
	"strings"

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
		sshRun,
		sshCopy,
	},
}

var sshConnect = cli.Command{
	Name:      "connect",
	Usage:     "Connect to the host with interactive shell",
	ArgsUsage: "<Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			_ = cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		sshConfig, err := GetSSHConfigFromHostName(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get a sshConfig : %s", err.Error())
		}
		err = sshConfig.Enter()
		if err != nil {
			return fmt.Errorf("Failed to launch ssh connection : %s", err.Error())
		}

		return nil
	},
}

var sshRun = cli.Command{
	Name:      "run",
	Usage:     "Run a command on the host",
	ArgsUsage: "<Host_name|Host_ID>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "Command to execute",
		}},
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			_ = cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Missing mandatory argument <Host_name>")
		}

		sshConfig, err := GetSSHConfigFromHostName(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get a sshConfig : %s", err.Error())
		}

		retcode, stdout, stderr, err := SSHCommandRun(c.String("c"), sshConfig)
		if err != nil {
			return fmt.Errorf("Failed to run the command : %s", err.Error())
		}

		fmt.Println(stdout)
		fmt.Fprintln(os.Stderr, stderr)

		fmt.Println("\nRetCode : ", retcode)
		return nil
	},
}

var sshCopy = cli.Command{
	Name:      "copy",
	Usage:     "Copy a local file/directory to an host or copy from host to local",
	ArgsUsage: "from to  Ex: /my/local/file.txt host1:/remote/path/",
	Flags:     []cli.Flag{},
	Action: func(c *cli.Context) error {
		if c.NArg() != 2 {
			_ = cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Missing mandatory argument <from> <to>")
		}

		var isUpload bool
		var hostName string
		var localPath string
		var remotePath string
		from := normalizeFileName(c.Args().Get(0))
		to := normalizeFileName(c.Args().Get(1))
		if len(strings.Split(to, ":")) == 2 {
			if len(strings.Split(from, ":")) == 2 {
				return fmt.Errorf("Copy Between host is not supported")
			}
			isUpload = true
			hostName = strings.Split(to, ":")[0]
			localPath = from
			remotePath = strings.Split(to, ":")[1]
		} else {
			isUpload = false
			hostName = strings.Split(from, ":")[0]
			localPath = to
			remotePath = strings.Split(from, ":")[1]
		}

		sshConfig, err := GetSSHConfigFromHostName(hostName)
		if err != nil {
			return fmt.Errorf("Failed to get a sshConfig : %s", err.Error())
		}

		retCode, _, stderr, err := sshConfig.Copy(remotePath, localPath, isUpload)
		if err != nil {
			return fmt.Errorf("Failed to copy file from %s to %s : %s", from, to, err.Error())
		} else if retCode != 0 {
			return fmt.Errorf("Failed to copy file from %s to %s : %s", from, to, stderr)
		}

		fmt.Printf("Copy of '%s' to '%s' done\n", from, to)
		return nil
	},
}

func GetSSHConfigFromHostName(hostName string) (*system.SSHConfig, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to get a new client : %s", err.Error())
	}

	mHost, err := metadata.LoadHost(client, hostName)
	if err != nil {
		return nil, fmt.Errorf("Failed to get host '%s' metadatas : %s", hostName, err.Error())
	} else if mHost == nil {
		return nil, fmt.Errorf("Failed to get host '%s' metadatas", hostName)
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
		return nil, fmt.Errorf("Failed to get host '%s' network Properties : %s", hostName, err.Error())
	}
	if hostNetworkV1.DefaultGatewayID != "" {
		mGw, err := metadata.LoadHost(client, hostNetworkV1.DefaultGatewayID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get host '%s' gateway metadatas : %s", hostName, err.Error())
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
	return &sshConfig, nil
}

func SSHCommandRun(command string, sshConfig *system.SSHConfig) (int, string, string, error) {
	sshCommand, err := sshConfig.Command(command)
	if err != nil {
		return 0, "", "", fmt.Errorf("Failed to get sshCommand : %s", err.Error())
	}

	return sshCommand.Run()
}

func normalizeFileName(fileName string) string {
	absPath, _ := filepath.Abs(fileName)
	if _, err := os.Stat(absPath); err != nil {
		return fileName
	}
	return absPath
}

func DisplaySSHConfig(sshConfig *system.SSHConfig) {
	fmt.Println(sshConfig)
}
