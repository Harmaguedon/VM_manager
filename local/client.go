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

import(
	"fmt"

	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/api"
	"github.com/libvirt/libvirt-go"
)

type Client struct {
	conn *libvirt.Connect
	MetadataBucketName string
}

//Create and initialize a ClientAPI
func (client *Client) Build(params map[string]interface{}) (api.ClientAPI, error){
	clientAPI := &Client{}
	uri, _ := params["Uri"].(string)

	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to libvirt : %s", err.Error())
	}

	clientAPI.conn = conn

	return clientAPI, nil
}

// GetAuthOpts returns authentification options as a Config
func (client *Client) GetAuthOpts() (model.Config, error){
	return nil, nil
}
// GetCfgOpts returns configuration options as a Config
func (client *Client) GetCfgOpts() (model.Config, error){
	return nil, nil
}
