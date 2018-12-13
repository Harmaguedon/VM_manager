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

// ImageCmd command
var ImageCmd = cli.Command{
	Name:  "image",
	Usage: "image COMMAND",
	Subcommands: []cli.Command{
		imageList,
	},
}

var imageList = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List available images",
	Flags:   []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := NewClient()
		if err != nil {
			return fmt.Errorf("Failed to get a new client : %s", err.Error())
		}

		images, err := client.ListImages(true)
		if err != nil {
			return fmt.Errorf("Failed to list images : %s", err.Error())
		}

		for _, image := range images {
			displayImage(&image)
		}
		return nil
	},
}

func displayImage(image *model.Image) {
	fmt.Println("\nTemplate :", image.Name)
	fmt.Println("	ID	:", image.ID)
}
