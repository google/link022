/* Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package syscmd contains methods that run external commands on device.
package syscmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	log "github.com/golang/glog"
)

// CommandRunner contains methods executing external commands.
type CommandRunner struct {
	// ExecCommand runs the given command with arguments.
	// It returns the content of stdout or stderr, and error if command failed.
	ExecCommand func(wait bool, cmd string, args ...string) (string, error)
}

// Runner executes external commands in the real environment.
func Runner() *CommandRunner {
	return &CommandRunner{
		ExecCommand: execute,
	}
}

func execute(wait bool, cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	var output []byte
	var err error

	if wait {
		output, err = command.CombinedOutput()
	} else {
		command.Stdout = os.Stdout
		err = command.Start()
	}

	outputString := string(output)
	if err != nil {
		log.Errorf("Command (%v %v) failed. Error: %v.\nOutput:\n%v", cmd, args, err, outputString)
	} else {
		log.V(2).Infof("Command (%v %v) succeeded.\nOutput:\n%v", cmd, args, outputString)
	}

	return outputString, err
}

func vlanINTFName(intfName string, vlanID int) string {
	return fmt.Sprintf("%s.%d", intfName, vlanID)
}

// SaveToFile saves the input string into a file in the file system.
// It creates the file and all parent folder if not exist.
func SaveToFile(folderPath, fileName, content string) error {
	fileFullPath := path.Join(folderPath, fileName)
	log.Infof("Saving content to file %v...", fileFullPath)

	fileInfo, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Config folder does not exist, create one.
		if err = os.MkdirAll(folderPath, 0755); err != nil {
			log.Errorf("Creating folder %v failed. Error: %v.", folderPath, err)
			return err
		} else {
			log.Infof("Created folder %v.", folderPath)
		}
	} else if !fileInfo.Mode().IsDir() {
		log.Errorf("Unable to save content to %v, since %v points an existing file.", fileFullPath, folderPath)
		return fmt.Errorf("%s points an existing file", folderPath)
	}

	if err := ioutil.WriteFile(fileFullPath, []byte(content), 0600); err != nil {
		log.Errorf("Saving content to file %v failed. Error: %v.", fileFullPath, err)
		return err
	}
	log.Infof("Saved content to file %v.", fileFullPath)
	return nil
}
