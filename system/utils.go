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

package system

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
)

//go:generate rice embed-go

// bashLibrayContent contains the content of the script bash_library.sh, that will be injected inside scripts through parameter {{.reserved_BashLibrary}}
var bashLibraryContent string

// GetBashLibrary generates the content of {{.reserved_BashLibrary}}
func GetBashLibrary() (string, error) {
	if bashLibraryContent == "" {
		box, err := rice.FindBox("../system/scripts")
		if err != nil {
			return "", err
		}

		// get file contents as string
		tmplContent, err := box.String("bash_library.sh")
		if err != nil {
			return "", err
		}
		// Prepare the template for execution
		tmplPrepared, err := template.New("bash_lbrary").Parse(tmplContent)
		if err != nil {
			return "", err
		}

		var buffer bytes.Buffer
		if err := tmplPrepared.Execute(&buffer, map[string]interface{}{}); err != nil {
			// TODO Use more explicit error
			return "", err
		}
		bashLibraryContent = buffer.String()
	}
	return bashLibraryContent, nil
}

// ExtractRetCode extracts info from the error
func ExtractRetCode(err error) (string, int, error) {
	retCode := -1
	msg := "__ NO MESSAGE __"
	if ee, ok := err.(*exec.ExitError); ok {
		//Try to get retCode
		if status, ok := ee.Sys().(syscall.WaitStatus); ok {
			retCode = status.ExitStatus()
		} else {
			return msg, retCode, fmt.Errorf("ExitError.Sys is not a 'syscall.WaitStatus'")
		}
		//Retrive error message
		msg = ee.Error()
		return msg, retCode, nil
	}
	return msg, retCode, fmt.Errorf("Error is not an 'ExitError'")
}
