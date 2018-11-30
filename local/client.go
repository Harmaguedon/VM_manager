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

import (
	"fmt"
	"github.com/libvirt/libvirt-go"
	"github.com/minio/minio-go"

	"github.com/CS-SI/LocalDriver/api"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/CS-SI/LocalDriver/metadata"
	"github.com/CS-SI/LocalDriver/providers"
)

type Client struct {
	LibvirtService 	*libvirt.Connect
	MinioService 	*minio.Client

	Config			*CfgOptions
	AuthOptions		*AuthOptions
}

type AuthOptions struct {
}
type CfgOptions struct {
	// MetadataBucketName contains the name of the bucket storing metadata
	MetadataBucketName string
}

//Create and initialize a ClientAPI
func (client *Client) Build(params map[string]interface{}) (api.ClientAPI, error){
	clientAPI := &Client{
		Config: 		&CfgOptions{},
		AuthOptions:	&AuthOptions{},
	}

	libvirt, err := libvirt.NewConnect(params["uri"].(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to libvirt : %s", err.Error())
	}
	clientAPI.LibvirtService = libvirt
	fmt.Println("Libvirt Connected")

	minio, err := minio.New(params["minioEndpoint"].(string), params["minioAccessKeyID"].(string), params["minioSecretAccessKey"].(string), params["minioUseSSL"].(bool))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to minio : %s", err.Error())
	}
	clientAPI.MinioService = minio
	fmt.Println("Minio Connected")
	if clientAPI.Config.MetadataBucketName == "" {
		clientAPI.Config.MetadataBucketName = metadata.BuildMetadataBucketName("id")
	}

	if _, err = clientAPI.GetContainer(clientAPI.Config.MetadataBucketName); err != nil {
		err = providers.InitializeBucket(clientAPI)
		if err != nil {
			return nil, fmt.Errorf("Failed to intialize the metadata bucket : %s", err.Error())
		}
	}

	return clientAPI, nil
}



// GetAuthOpts returns authentification options as a Config
func (client *Client) GetAuthOpts() (model.Config, error){
	cfg := model.ConfigMap{}

	return cfg, nil
}
// GetCfgOpts returns configuration options as a Config
func (client *Client) GetCfgOpts() (model.Config, error){

	config := model.ConfigMap{}

	config.Set("MetadataBucket", client.Config.MetadataBucketName)

	return config, nil
}
