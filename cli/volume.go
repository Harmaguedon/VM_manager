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
	"strings"

	"github.com/CS-SI/LocalDriver/metadata"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/model/enums/HostProperty"
	"github.com/CS-SI/LocalDriver/model/enums/VolumeProperty"
	propsv1 "github.com/CS-SI/LocalDriver/model/properties/v1"
	"github.com/CS-SI/LocalDriver/system"
	"github.com/CS-SI/LocalDriver/system/nfs"

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
		volumeAttach,
		volumeDetach,
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

		err = metadata.SaveVolume(client, volume)
		if err != nil {
			return fmt.Errorf("Failed to save volume metadatas : %s", err.Error())
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

		for _, volumeName := range volumeList {
			mVolume, err := metadata.LoadVolume(client, volumeName)
			if err != nil {
				return fmt.Errorf("Failed to load volume '%s' from metadatas : %s", volumeName, err.Error())
			}
			volume := mVolume.Get()

			err = client.DeleteVolume(volumeName)
			if err != nil {
				return fmt.Errorf("Failed to delete '%s' volume : %s", volumeName, err.Error())
			}
			fmt.Println(fmt.Sprintf("Volume '%s' sucessfully deleted", volumeName))

			err = metadata.RemoveVolume(client, volume.ID)
			if err != nil {
				return fmt.Errorf("Failed to save volume metadatas : %s", err.Error())
			}
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

var volumeAttach = cli.Command{
	Name:      "attach",
	Usage:     "Attach a volume to an host",
	ArgsUsage: "<Volume_name|Volume_ID> <Host_name|Host_ID>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "path",
			Value: model.DefaultVolumeMountPoint,
			Usage: "Mount point of the volume",
		},
		cli.StringFlag{
			Name:  "format",
			Value: "ext4",
			Usage: "Filesystem format",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() != 2 {
			_ = cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Missing mandatory argument <Volume_name> and/or <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		mVolume, err := metadata.LoadVolume(client, c.Args().Get(0))
		if err != nil {
			return fmt.Errorf("Volume '%s' not found in metadatas", c.Args().Get(0))
		}
		volume := mVolume.Get()
		volumeAttachedV1 := propsv1.NewVolumeAttachments()
		err = volume.Properties.Get(VolumeProperty.AttachedV1, volumeAttachedV1)
		if err != nil {
			return fmt.Errorf("Failed to get volume propertie AttachedV1 : %s", err.Error())
		}

		mHost, err := metadata.LoadHost(client, c.Args().Get(1))
		if err != nil || mHost == nil {
			return fmt.Errorf("Host '%s' not found in metadatas", c.Args().Get(1))
		}
		host := mHost.Get()
		hostVolumesV1 := propsv1.NewHostVolumes()
		err = host.Properties.Get(HostProperty.VolumesV1, hostVolumesV1)
		if err != nil {
			return fmt.Errorf("Failed to get host propertie hostVolumesV1 : %s", err.Error())
		}
		hostMountsV1 := propsv1.NewHostMounts()
		err = host.Properties.Get(HostProperty.MountsV1, hostMountsV1)
		if err != nil {
			return fmt.Errorf("Failed to get volume propertie hostMountsV1 : %s", err.Error())
		}

		sshConfig, err := GetSSHConfigFromHostName(c.Args().Get(1))
		if err != nil {
			return fmt.Errorf("Failed get the sshConfig : %s", err.Error())
		}
		oldDiskSet, err := listAttachedDevices(sshConfig)
		if err != nil {
			return fmt.Errorf("Failed to get list of connected disks : %s", err.Error())
		}

		volumeAttachmentRequest := model.VolumeAttachmentRequest{
			Name:     "attachment-" + host.Name + "-" + volume.Name,
			VolumeID: volume.ID,
			HostID:   host.ID,
		}
		vaID, err := client.CreateVolumeAttachment(volumeAttachmentRequest)
		if err != nil {
			return fmt.Errorf("Failed to create an attachment between volume '%s' and host '%s' : %s", c.Args().Get(0), c.Args().Get(1), err.Error())
		}

		newDiskSet, err := listAttachedDevices(sshConfig)
		if err != nil {
			return fmt.Errorf("Failed to get list of connected disks : %s", err.Error())
		}

		diff := difference(oldDiskSet, newDiskSet)
		if len(diff) != 1 {
			return fmt.Errorf("Failed to create an attachment between volume '%s' and host '%s' : %s", c.Args().Get(0), c.Args().Get(1), err.Error())
		}
		newDisk := diff[0]
		diskName := "/dev/" + newDisk

		server, err := nfs.NewServer(sshConfig)
		if err != nil {
			return fmt.Errorf("Failed to creare the nfsServer : %s", err.Error())
		}
		err = server.MountBlockDevice(diskName, c.String("path"), c.String("format"))
		if err != nil {
			return fmt.Errorf("Failed to mount the block device : %s", err.Error())
		}

		volumeAttachedV1.Hosts[host.ID] = host.Name
		err = volume.Properties.Set(VolumeProperty.AttachedV1, volumeAttachedV1)
		if err != nil {
			return fmt.Errorf("Failed to set volume propertie AttachedV1 : %s", err.Error())
		}
		err = metadata.SaveVolume(client, volume)
		if err != nil {
			return fmt.Errorf("Failed to save volume metadatas : %s", err.Error())
		}

		hostVolumesV1.VolumesByID[volume.ID] = &propsv1.HostVolume{
			AttachID: vaID,
			Device:   diskName,
		}
		hostVolumesV1.VolumesByName[volume.Name] = volume.ID
		hostVolumesV1.VolumesByDevice[diskName] = volume.ID
		hostVolumesV1.DevicesByID[volume.ID] = diskName
		err = host.Properties.Set(HostProperty.VolumesV1, hostVolumesV1)
		if err != nil {
			return fmt.Errorf("Failed to set host propertie hostVolumesV1 : %s", err.Error())
		}
		hostMountsV1.LocalMountsByPath[c.String("path")] = &propsv1.HostLocalMount{
			Device:     diskName,
			Path:       c.String("path"),
			FileSystem: "nfs",
		}
		hostMountsV1.LocalMountsByDevice[diskName] = c.String("path")
		err = host.Properties.Set(HostProperty.MountsV1, hostMountsV1)
		if err != nil {
			return fmt.Errorf("Failed to set host propertie hostMountsV1 : %s", err.Error())
		}
		err = metadata.SaveHost(client, host)
		if err != nil {
			return fmt.Errorf("Failed to save host metadatas : %s", err.Error())
		}

		fmt.Printf("Volume '%s' attached to host '%s'\n", c.Args().Get(0), c.Args().Get(1))
		return nil
	},

	//TODO check if mout path is not duplicated!
}

var volumeDetach = cli.Command{
	Name:      "detach",
	Usage:     "Detach a volume from an host",
	ArgsUsage: "<Volume_name|Volume_ID> <Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		//TODO when detaching a volume the name of the other volumes evolves (detach all and re attach all execpt the chosen one?)
		if c.NArg() != 2 {
			_ = cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Missing mandatory argument <Volume_name> and/or <Host_name>")
		}

		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		mVolume, err := metadata.LoadVolume(client, c.Args().Get(0))
		if err != nil {
			return fmt.Errorf("Volume '%s' not found in metadatas", c.Args().Get(0))
		}
		volume := mVolume.Get()
		volumeAttachedV1 := propsv1.NewVolumeAttachments()
		err = volume.Properties.Get(VolumeProperty.AttachedV1, volumeAttachedV1)
		if err != nil {
			return fmt.Errorf("Failed to get volume propertie AttachedV1 : %s", err.Error())
		}

		mHost, err := metadata.LoadHost(client, c.Args().Get(1))
		if err != nil || mHost == nil {
			return fmt.Errorf("Host '%s' not found in metadatas", c.Args().Get(1))
		}
		host := mHost.Get()
		hostVolumesV1 := propsv1.NewHostVolumes()
		err = host.Properties.Get(HostProperty.VolumesV1, hostVolumesV1)
		if err != nil {
			return fmt.Errorf("Failed to get host propertie hostVolumesV1 : %s", err.Error())
		}
		hostMountsV1 := propsv1.NewHostMounts()
		err = host.Properties.Get(HostProperty.MountsV1, hostMountsV1)
		if err != nil {
			return fmt.Errorf("Failed to get volume propertie hostMountsV1 : %s", err.Error())
		}

		attachment, found := hostVolumesV1.VolumesByID[volume.ID]
		if !found {
			return fmt.Errorf("Fo attachments found, volume '%s' not attached to host '%s'", volume.Name, host.Name)
		}
		device := attachment.Device
		path := hostMountsV1.LocalMountsByDevice[device]
		mount := hostMountsV1.LocalMountsByPath[path]
		if mount == nil {
			return fmt.Errorf("metadata inconsistency: no mount corresponding to volume attachment")
		}

		sshConfig, err := GetSSHConfigFromHostName(c.Args().Get(1))
		if err != nil {
			return fmt.Errorf("Failed get the sshConfig : %s", err.Error())
		}
		server, err := nfs.NewServer(sshConfig)
		if err != nil {
			return fmt.Errorf("Failed to creare the nfsServer : %s", err.Error())
		}
		err = server.UnmountBlockDevice(attachment.Device)
		if err != nil {
			return fmt.Errorf("Failed to mount the block device : %s", err.Error())
		}
		err = client.DeleteVolumeAttachment(host.ID, attachment.AttachID)
		if err != nil {
			return fmt.Errorf("Failed to delete the volume attachment : %s", err.Error())
		}

		delete(hostVolumesV1.VolumesByID, volume.ID)
		delete(hostVolumesV1.VolumesByName, volume.Name)
		delete(hostVolumesV1.VolumesByDevice, attachment.Device)
		delete(hostVolumesV1.DevicesByID, volume.ID)
		err = host.Properties.Set(HostProperty.VolumesV1, hostVolumesV1)
		if err != nil {
			return fmt.Errorf("Failed to set host propertie hostVolumesV1 : %s", err.Error())
		}

		delete(hostMountsV1.LocalMountsByDevice, mount.Device)
		delete(hostMountsV1.LocalMountsByPath, mount.Path)
		err = host.Properties.Set(HostProperty.MountsV1, hostMountsV1)
		if err != nil {
			return fmt.Errorf("Failed to set host propertie hostMountsV1 : %s", err.Error())
		}

		err = metadata.SaveHost(client, host)
		if err != nil {
			return fmt.Errorf("Failed to save host metadatas : %s", err.Error())
		}

		delete(volumeAttachedV1.Hosts, host.ID)
		err = volume.Properties.Set(VolumeProperty.AttachedV1, volumeAttachedV1)
		if err != nil {
			return fmt.Errorf("Failed to set volume propertie volumeAttachedV1 : %s", err.Error())
		}

		err = metadata.SaveVolume(client, volume)
		if err != nil {
			return fmt.Errorf("Failed to save volume metadatas : %s", err.Error())
		}

		fmt.Printf("Volume '%s' detached from host '%s'\n", c.Args().Get(0), c.Args().Get(1))
		return nil
	},
}

func displayVolume(volume *model.Volume) {
	fmt.Println(volume)
}

func listAttachedDevices(sshConfig *system.SSHConfig) ([]string, error) {
	disks := []string{}
	command := "sudo lsblk -l -o NAME,TYPE | grep disk | cut -d' ' -f1"
	retcode, stdout, stderr, err := SSHCommandRun(command, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to run the command : %s", err.Error())
	} else if retcode != 0 {
		return nil, fmt.Errorf("Command did not finish properly : %s", stderr)
	}

	for _, disk := range strings.Split(stdout, "\n") {
		disks = append(disks, disk)
	}

	return disks, nil
}

func difference(slice1 []string, slice2 []string) []string {
	diffStr := []string{}
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			diffStr = append(diffStr, mKey)
		}
	}

	return diffStr
}
