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

package main

import (
	"fmt"
	"os"
	"sort"

	cliL "github.com/CS-SI/LocalDriver/cli"
	cli "github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "virt"
	app.Usage = "virt COMMAND"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "CS-SI",
			Email: "safescale@c-s.fr",
		},
	}

	app.EnableBashCompletion = true

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "Print program version",
	}

	app.Flags = []cli.Flag{}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Commands = append(app.Commands, cliL.HostCmd)
	sort.Sort(cli.CommandsByName(cliL.HostCmd.Subcommands))

	app.Commands = append(app.Commands, cliL.NetworkCmd)
	sort.Sort(cli.CommandsByName(cliL.NetworkCmd.Subcommands))

	// app.Commands = append(app.Commands, cmd.TenantCmd)
	// sort.Sort(cli.CommandsByName(cmd.TenantCmd.Subcommands))

	app.Commands = append(app.Commands, cliL.VolumeCmd)
	sort.Sort(cli.CommandsByName(cliL.VolumeCmd.Subcommands))

	app.Commands = append(app.Commands, cliL.SSHCmd)
	sort.Sort(cli.CommandsByName(cliL.SSHCmd.Subcommands))

	// app.Commands = append(app.Commands, cmd.BucketCmd)
	// sort.Sort(cli.CommandsByName(cmd.BucketCmd.Subcommands))

	// app.Commands = append(app.Commands, cmd.ShareCmd)
	// sort.Sort(cli.CommandsByName(cmd.ShareCmd.Subcommands))

	//app.Commands = append(app.Commands, cliL.ImageCmd)
	//sort.Sort(cli.CommandsByName(cliL.ImageCmd.Subcommands))

	// useless
	// app.Commands = append(app.Commands, cliL.TemplateCmd)
	// sort.Sort(cli.CommandsByName(cliL.TemplateCmd.Subcommands))

	sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
