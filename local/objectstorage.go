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
	"github.com/CS-SI/SafeScale/providers/model"
)

// CreateContainer creates an object container
func (client *Client) CreateContainer(name string) error {
	return nil
}

// DeleteContainer deletes an object container
func (client *Client) DeleteContainer(name string) error {
	return nil
}

// UpdateContainer updates an object container
func (client *Client) UpdateContainer(name string, meta map[string]string) error {
	return nil
}

// GetContainerMetadata get an object container metadata
func (client *Client) GetContainerMetadata(name string) (map[string]string, error) {
	return nil, nil
}

// Getcontainer returns info of the container
func (client *Client) GetContainer(name string) (*model.ContainerInfo, error){
	return nil, nil
}
// ListContainers list object containers
func (client *Client) ListContainers() ([]string, error) {
	return nil, nil
}

// PutObject put an object into an object container
func (client *Client) PutObject(container string, obj model.Object) error {
	return nil
}

// UpdateObjectMetadata update an object into an object container
func (client *Client) UpdateObjectMetadata(container string, obj model.Object) error {
	return nil
}

// GetObject get  object content from an object container
func (client *Client) GetObject(container string, name string, ranges []model.Range) (*model.Object, error) {
	return nil, nil
}

// GetObjectMetadata get  object metadata from an object container
func (client *Client) GetObjectMetadata(container string, name string) (*model.Object, error) {
	return nil, nil
}

// ListObjects list objects of a container
func (client *Client) ListObjects(container string, filter model.ObjectFilter) ([]string, error) {
	return nil, nil
}

// CopyObject copies an object
func (client *Client) CopyObject(containerSrc, objectSrc, objectDst string) error {
	return nil
}

// DeleteObject deleta an object from a container
func (client *Client) DeleteObject(container, object string) error {
	return nil
}

