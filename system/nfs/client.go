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

package nfs

import (
	"fmt"

	"github.com/CS-SI/LocalDriver/system"
)

// Client defines the structure of a Client object
type Client struct {
	SshConfig *system.SSHConfig
}

// NewNFSClient creates a new NFS client isntance
func NewNFSClient(sshconfig *system.SSHConfig) (*Client, error) {
	if sshconfig == nil {
		return nil, fmt.Errorf("invalid parameter: 'sshconfig' can't be nil")
	}

	client := &Client{
		SshConfig: sshconfig,
	}
	return client, nil
}

// Install installs NFS client on remote host
func (c *Client) Install() error {
	retcode, stdout, stderr, err := executeScript(*c.SshConfig, "nfs_client_install.sh", map[string]interface{}{})
	return handleExecuteScriptReturn(retcode, stdout, stderr, err, "Error executing script to install NFS client")
}

// Mount defines a mount of a remote share and mount it
func (c *Client) Mount(host string, share string, mountPoint string) error {
	data := map[string]interface{}{
		"Host":       host,
		"Share":      share,
		"MountPoint": mountPoint,
	}
	retcode, stdout, stderr, err := executeScript(*c.SshConfig, "nfs_client_share_mount.sh", data)
	return handleExecuteScriptReturn(retcode, stdout, stderr, err, "Error executing script to mount remote NFS share")
}

// Unmount a nfs share from NFS server
func (c *Client) Unmount(host string, share string) error {
	data := map[string]interface{}{
		"Host":  host,
		"Share": share,
	}
	retcode, stdout, stderr, err := executeScript(*c.SshConfig, "nfs_client_share_unmount.sh", data)
	return handleExecuteScriptReturn(retcode, stdout, stderr, err, "Error executing script to unmount remote NFS share")
}
