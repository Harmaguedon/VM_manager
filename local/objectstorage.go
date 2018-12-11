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
	"bytes"
	"fmt"
	"github.com/CS-SI/LocalDriver/model"
	"github.com/minio/minio-go"
	"io"
	"strings"
)

const maxObjectSize = 1073741824 //bytes

//-------------CONTAINERS MANAGEMENT------------------------------------------------------------------------------------

// CreateContainer creates an object container
func (client *Client) CreateContainer(name string) error {
	err := client.MinioService.MakeBucket(name, "")
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to create the container %s : %s", name, err.Error()))
	}
	return nil
}

// DeleteContainer deletes an object container
func (client *Client) DeleteContainer(name string) error {
	err := client.MinioService.RemoveBucket(name)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to delete the container %s : %s", name, err.Error()))
	}
	return nil
}

// Getcontainer returns info of the container
func (client *Client) GetContainer(name string) (*model.Bucket, error){
	exists, err := client.MinioService.BucketExists(name)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Not Able to check the existance the container %s : %s", name, err.Error()))
	} else if !exists {
		return nil, fmt.Errorf(fmt.Sprintf("Container %s does not exists", name))
	}

	location, err := client.MinioService.GetBucketLocation(name)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Not Able to find the location of the container %s : %s", name, err.Error()))
	}
	//objectsNumber := 0
	//doneCh := make(chan struct{})
	//defer close(doneCh)
	//for range client.MinioService.ListObjects(name, "", true, doneCh){
	//	objectsNumber ++
	//}

	return &model.Bucket{
		Name:		name,
		Host:		"Not implemented yet",
		MountPoint:	location,
		//NbItems:	objectsNumber,
	}, nil
}

// ListContainers list object containers
func (client *Client) ListContainers() ([]string, error) {
	bucketNames := []string{}
	bucketInfos, err := client.MinioService.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Not Able to list the containers : %s", err.Error()))
	}
	for _, bucketInfo := range bucketInfos {
		bucketNames = append(bucketNames, bucketInfo.Name)
	}
	return bucketNames, nil
}



//-------------OBJECTS MANAGEMENT---------------------------------------------------------------------------------------

// Check if a given object exists in a given container of the object storage
func objectExists(containerName string, objectName string, minioService *minio.Client) (bool, error){
	object, err := minioService.GetObject(containerName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return false, err
	}
	if object == nil {
		return false, nil
	}

	_, err = object.Stat()

	if err != nil{
		if err.Error() == "The specified key does not exist."{
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// PutObject put an object into an object container
func (client *Client) PutObject(container string, obj model.Object) error {
	putOpts := minio.PutObjectOptions{
		ContentType: 	obj.ContentType,
		// TODO
		//UserMetadata:	obj.Metadata,
	}
	var objSize int64 = -1
	if obj.ContentLength != 0 {
		objSize = obj.ContentLength
	}

	_, err := client.MinioService.PutObject(container, obj.Name, obj.Content, objSize, putOpts)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to put %s file to container %s : %s", obj.Name, container, err.Error()))
	}

	return nil
}

// GetObject get  object content from an object container
func (client *Client) GetObject(container string, name string, ranges []model.Range) (*model.Object, error) {
	exists, err := objectExists(container, name, client.MinioService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to know if the object exists : %s", err.Error()))
	} else if !exists {
		return nil, fmt.Errorf(fmt.Sprintf("File %s does not exists", name))
	}

	object, err :=  client.MinioService.GetObject(container, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to get the object %s : %s", name, err.Error()))
	}
	info, err := object.Stat()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to get %s's info : %s", name, err.Error()))
	}

	if info.Size > maxObjectSize {
		return nil, fmt.Errorf(fmt.Sprintf("Object is to voluminous, the maximal size is %dGB", maxObjectSize/(1024*1024*1024)))
	}

	writer := bytes.NewBuffer([]byte{})
	io.CopyN(writer, object, info.Size)
	buffer := writer.Bytes()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to read data from object %s : %s", name, err.Error()))
	}

	if ranges != nil {
		subBuffers := [][]byte{}
		for _, rg := range ranges {
			if *rg.From < 0 {
				*rg.From = 0
			}
			if *rg.To > len(buffer)-1 {
				*rg.To = len(buffer)-1
			}
			if *rg.From > *rg.To {
				*rg.From, *rg.To = *rg.To, *rg.From
			}
			subBuffers = append(subBuffers, buffer[*rg.From:*rg.To+1])
		}
		buffer = []byte{}
		for _, subBuffer := range subBuffers {
			buffer = append(buffer, subBuffer...)
		}
	}

	metadataNew := map[string]string{}
	metadataOld := info.Metadata
	delete(metadataOld, "X-Minio-Deployment-Id")
	for key, value := range metadataOld {
		keySplit := strings.Split(key, "-")
		metadataNew[keySplit[len(keySplit)-1]] = value[0]
	}

	return &model.Object{
		Name:         name,
		Content:      bytes.NewReader(buffer),
		// TODO Metadata:     metadataNew,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
		ContentLength:info.Size,
	}, nil
}

// DeleteObject deleta an object from a container
func (client *Client) DeleteObject(container string , object string) error {
	exists, err := objectExists(container, object, client.MinioService)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Unable to know if the object exists : %s", err.Error()))
	} else if !exists {
		return fmt.Errorf(fmt.Sprintf("File %s does not exists", object))
	}

	err = client.MinioService.RemoveObject(container, object)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to remove object %s : %s", object, err.Error()))
	}

	return nil
}

// ListObjects list objects of a container
func (client *Client) ListObjects(container string, filter model.ObjectFilter) ([]string, error) {
	objectNames := []string{}
	if filter.Path != "" &&  filter.Path[len(filter.Path)-1:] != "/" {
		filter.Path += "/"
	}
	doneCh := make(chan struct{})
	defer close(doneCh)
	for object := range client.MinioService.ListObjects(container, filter.Path+filter.Prefix, true, doneCh){
		if object.Err == nil {
			objectNames = append(objectNames, object.Key)
		}
	}
	return objectNames, nil
}

// CopyObject copies an object
func (client *Client) CopyObject(containerSrc, objectSrc, objectDst string) error {
	sourceInfo 				:= minio.NewSourceInfo(containerSrc, objectSrc, nil)
	destinationInfo, err	:= minio.NewDestinationInfo(containerSrc, objectDst, nil, nil)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to create destination Infos while copying %s to %s in container %s : %s", objectSrc, objectDst, containerSrc, err.Error()))
	}
	err = client.MinioService.CopyObject(destinationInfo, sourceInfo)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to copy %s to %s in container %s : %s", objectSrc, objectDst, containerSrc, err.Error()))
	}
	return nil
}



//-------------METADATAS MANAGEMENT-------------------------------------------------------------------------------------

// GetObjectMetadata get object metadata from an object container
func (client *Client) GetObjectMetadata(container string, name string) (*model.Object, error) {
	exists, err := objectExists(container, name, client.MinioService)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to know if the object exists : %s", err.Error()))
	} else if !exists {
		return nil, fmt.Errorf(fmt.Sprintf("File %s does not exists", name))
	}

	object, err :=  client.MinioService.GetObject(container, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to get the object %s : %s", name, err.Error()))
	}
	info, err := object.Stat()
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to get %s's info : %s", name, err.Error()))
	}

	metadataNew := map[string]string{}
	metadataOld := info.Metadata
	delete(metadataOld, "X-Minio-Deployment-Id")
	for key, value := range metadataOld {
		keySplit := strings.Split(key, "-")
		metadataNew[keySplit[len(keySplit)-1]] = value[0]
	}

	return &model.Object{
		Name:         name,
		//TODO Metadata:     metadataNew,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
		ContentLength:info.Size,
	}, nil
}

// UpdateObjectMetadata update an object into an object container
func (client *Client) UpdateObjectMetadata(container string, obj model.Object) error {
	objectOld, err := client.GetObject(container, obj.Name, nil)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to GET the object %s of the container %s : %s", obj.Name, container, err.Error()))
	}

	//TODO objectOld.Metadata = obj.Metadata

	err = client.PutObject(container, *objectOld)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to PUT the object %s on the container %s : %s", obj.Name, container, err.Error()))
	}
	return nil
}




