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

package providers

import (
	"fmt"

	"github.com/CS-SI/LocalDriver/api"
)

// InitializeBucket creates the Object Storage Container/Bucket that will store the metadata
// id contains a unique identifier of the tenant (something coming from the providers, not the tenant name)
func InitializeBucket(svc api.ClientAPI) error {
	cfg, err := svc.GetCfgOpts()
	if err != nil {
		fmt.Printf("failed to get client options: %s\n", err.Error())
	}
	anon, found := cfg.Get("MetadataBucket")
	if !found || anon.(string) == "" {
		return fmt.Errorf("failed to get value of option 'MetadataBucket'")
	}
	return svc.CreateContainer(anon.(string))
}
