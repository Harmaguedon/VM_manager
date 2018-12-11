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

	"github.com/CS-SI/LocalDriver/model"

	"github.com/urfave/cli"
)

//VolumeCmd volume command
var VolumeCmd = cli.Command{
	Name:  "volume",
	Usage: "volume COMMAND",
	Subcommands: []cli.Command{
		volumeCreate,
		volumeDelete,
		volumeList,
		volumeInspect,
		// volumeAttach,
		// volumeDetach,
	},
}

var volumeCreate = cli.Command{
	Name:      "create",
	Aliases:   []string{"new"},
	Usage:     "Create a volume",
	ArgsUsage: "<Volume_name>",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "size",
			Value: 10,
			Usage: "Size of the volume (in Go)",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			return fmt.Errorf("Missing mandatory argument <Volume_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		volumeRequest := model.VolumeRequest{
			Name: c.Args().First(),
			Size: c.Int("size"),
		}

		volume, err := client.CreateVolume(volumeRequest)
		if err != nil {
			return fmt.Errorf("Failed to create volume %s", err.Error())
		}

		displayVolume(volume)

		return nil
	},
}

var volumeDelete = cli.Command{
	Name:      "delete",
	Aliases:   []string{"rm", "remove"},
	Usage:     "Delete volume",
	ArgsUsage: "<Volume_name|Volume_ID> [<Volume_name|Volume_ID>...]",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Volume_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		var volumeList []string
		volumeList = append(volumeList, c.Args().First())
		volumeList = append(volumeList, c.Args().Tail()...)

		for _, volume := range volumeList {
			err = client.DeleteVolume(volume)
			if err != nil {
				return fmt.Errorf("Failed to delete '%s' volume : %s", volume, err.Error())
			}
			fmt.Println(fmt.Sprintf("Volume '%s' sucessfully deleted", volume))
		}

		return nil
	},
}

var volumeList = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List available volumes",
	Flags:   []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		volumes, err := client.ListVolumes()
		if err != nil {
			return fmt.Errorf("Failed to list volumes %s", err.Error())
		}

		for _, volume := range volumes {
			displayVolume(&volume)
		}

		return nil
	},
}

var volumeInspect = cli.Command{
	Name:      "inspect",
	Aliases:   []string{"show"},
	Usage:     "Inspect volume",
	ArgsUsage: "<Volume_name|Volume_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing mandatory argument <Volume_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		volume, err := client.GetVolume(c.Args().First())
		if err != nil {
			return fmt.Errorf("Failed to get volume %s", err.Error())
		}

		displayVolume(volume)

		return nil
	},
}

// var volumeAttach = cli.Command{
// 	Name:      "attach",
// 	Usage:     "Attach a volume to an host",
// 	ArgsUsage: "<Volume_name|Volume_ID> <Host_name|Host_ID>",
// 	Flags: []cli.Flag{
// 		cli.StringFlag{
// 			Name:  "path",
// 			Value: model.DefaultVolumeMountPoint,
// 			Usage: "Mount point of the volume",
// 		},
// 		cli.StringFlag{
// 			Name:  "format",
// 			Value: "ext4",
// 			Usage: "Filesystem format",
// 		},
// 	},
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 2 {
// 			fmt.Println("Missing mandatory argument <Volume_name> and/or <Host_name>")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return clitools.ExitOnInvalidArgument()
// 		}
// 		def := pb.VolumeAttachment{
// 			Format:    c.String("format"),
// 			MountPath: c.String("path"),
// 			Host:      &pb.Reference{Name: c.Args().Get(1)},
// 			Volume:    &pb.Reference{Name: c.Args().Get(0)},
// 		}
// 		err := client.New().Volume.Attach(def, client.DefaultExecutionTimeout)
// 		if err != nil {
// 			return clitools.ExitOnRPC(utils.TitleFirst(client.DecorateError(err, "attach of volume", true).Error()))
// 		}
// 		fmt.Printf("Volume '%s' attached to host '%s'\n", c.Args().Get(0), c.Args().Get(1))
// 		return nil
// 	},
// }

// var volumeDetach = cli.Command{
// 	Name:      "detach",
// 	Usage:     "Detach a volume from an host",
// 	ArgsUsage: "<Volume_name|Volume_ID> <Host_name|Host_ID>",
// 	Action: func(c *cli.Context) error {
// 		if c.NArg() != 2 {
// 			fmt.Println("Missing mandatory argument <Volume_name> and/or <Host_name>")
// 			_ = cli.ShowSubcommandHelp(c)
// 			return clitools.ExitOnInvalidArgument()
// 		}
// 		err := client.New().Volume.Detach(c.Args().Get(0), c.Args().Get(1), client.DefaultExecutionTimeout)
// 		if err != nil {
// 			return clitools.ExitOnRPC(utils.TitleFirst(client.DecorateError(err, "unattach of volume", true).Error()))
// 		}

// 		fmt.Printf("Volume '%s' detached from host '%s'\n", c.Args().Get(0), c.Args().Get(1))
// 		return nil
// 	},
// }

// type volumeInfoDisplayable struct {
// 	ID        string
// 	Name      string
// 	Speed     string
// 	Size      int32
// 	Host      string
// 	MountPath string	app.Commands = append(app.Commands, cmd.VolumeCmd)
//	sort.Sort(cli.CommandsByName(cmd.VolumeCmd.Subcommands))
// 	Format    string
// 	Device    string
// }

// type volumeDisplayable struct {
// 	ID    string
// 	Name  string
// 	Speed string
// 	Size  int32
// }

// func toDisplaybleVolumeInfo(volumeInfo *pb.VolumeInfo) *volumeInfoDisplayable {
// 	return &volumeInfoDisplayable{
// 		volumeInfo.GetID(),
// 		volumeInfo.GetName(),
// 		pb.VolumeSpeed_name[int32(volumeInfo.GetSpeed())],
// 		volumeInfo.GetSize(),
// 		brokerutils.GetReference(volumeInfo.GetHost()),
// 		volumeInfo.GetMountPath(),
// 		volumeInfo.GetFormat(),
// 		volumeInfo.GetDevice(),
// 	}
// }

// func toDisplaybleVolume(volumeInfo *pb.Volume) *volumeDisplayable {
// 	return &volumeDisplayable{
// 		volumeInfo.GetID(),
// 		volumeInfo.GetName(),
// 		pb.VolumeSpeed_name[int32(volumeInfo.GetSpeed())],
// 		volumeInfo.GetSize(),
// 	}
// }

// func getAllowedSpeeds() string {
// 	speeds := ""
// 	i := 0
// 	for k := range pb.VolumeSpeed_value {
// 		if i > 0 {
// 			speeds += ", "
// 		}
// 		speeds += k
// 		i++

// 	}
// 	return speeds
// }

func displayVolume(volume *model.Volume) {
	fmt.Println(volume)
}
