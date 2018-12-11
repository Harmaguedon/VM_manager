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
	"sort"

	"github.com/CS-SI/LocalDriver/api"
	"github.com/CS-SI/LocalDriver/local"
	"github.com/CS-SI/LocalDriver/model"
)

var tenantLocal = map[string]interface{}{
	"uri":                  "qemu:///system",
	"lanInterface":         "enp2s0",
	"minioEndpoint":        "localhost:9000",
	"minioAccessKeyID":     "accesKey",
	"minioSecretAccessKey": "secretKey",
	"minioUseSSL":          false,
}

const (
	//CoreDRFWeight is the Dominant Resource Fairness weight of a core
	CoreDRFWeight float32 = 1.0
	//RAMDRFWeight is the Dominant Resource Fairness weight of 1 GB of RAM
	RAMDRFWeight float32 = 1.0 / 8.0
	//DiskDRFWeight is the Dominant Resource Fairness weight of 1 GB of Disk
	DiskDRFWeight float32 = 1.0 / 16.0
)

// RankDRF computes the Dominant Resource Fairness Rank of an host template
func RankDRF(t *model.HostTemplate) float32 {
	fc := float32(t.Cores)
	fr := t.RAMSize
	fd := float32(t.DiskSize)
	return fc*CoreDRFWeight + fr*RAMDRFWeight + fd*DiskDRFWeight
}

// ByRankDRF implements sort.Interface for []HostTemplate based on
// the Dominant Resource Fairness
type ByRankDRF []*model.HostTemplate

func (a ByRankDRF) Len() int           { return len(a) }
func (a ByRankDRF) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRankDRF) Less(i, j int) bool { return RankDRF(a[i]) < RankDRF(a[j]) }

func NewClient() (api.ClientAPI, error) {
	client, err := (&local.Client{}).Build(tenantLocal)
	if err != nil {
		return nil, fmt.Errorf("Build failed : %s", err.Error())
	}
	return client, nil
}

func SelectTemplateBySize(sizing model.SizingRequirements, templates []model.HostTemplate) (*model.HostTemplate, error) {
	var selectedTpls []*model.HostTemplate
	for _, template := range templates {
		if template.Cores >= sizing.MinCores && (template.DiskSize == 0 || template.DiskSize >= sizing.MinDiskSize) && template.RAMSize >= sizing.MinRAMSize {
			selectedTpls = append(selectedTpls, &template)
		}
	}

	if len(selectedTpls) == 0 {
		return nil, fmt.Errorf("No template matching requested size")
	}

	sort.Sort(ByRankDRF(selectedTpls))

	return selectedTpls[0], nil
}
