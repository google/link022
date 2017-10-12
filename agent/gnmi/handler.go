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

package gnmi

import (
    "errors"
    "fmt"
    "time"

    "github.com/google/link022/agent/context"
    "github.com/google/link022/agent/util/ocutil"
    "github.com/google/link022/agent/service"
    "github.com/google/link022/agent/syscmd"
    "github.com/google/link022/generated/ocstruct"
    "github.com/openconfig/ygot/ygot"

    pb "github.com/openconfig/gnmi/proto/gnmi"
    log "github.com/golang/glog"
)

var (
    cmdRunner = syscmd.Runner()
)

// handleSet is the callback function of the GNMI SET call.
// It is triggered by the GNMI server.
func handleSet(op pb.UpdateResult_Operation, updatedPath *pb.Path, updatedConfig interface{},
               existingConfig ygot.ValidatedGoStruct) error {
    // TODO: Handle delta change. Currently the GNMI server only supports replacing root.
    officeConfig, ok := updatedConfig.(*ocstruct.Office)
    if !ok {
        return errors.New("new configuration has invalid type")
    }

    configString, err := ygot.EmitJSON(officeConfig, &ygot.EmitJSONConfig{
        Format: ygot.RFC7951,
        Indent: "  ",
        RFC7951Config: &ygot.RFC7951JSONConfig{
            AppendModuleName: false,
        },
    })
    if err != nil {
        return err
    }
    log.Infof("Received a new configuration:\n%v\n", configString)

    // TODO: Validate the OpenConfig module.
    deviceConfig := context.GetDeviceConfig()

    // Check and clean up the existing configuration.
    var changedVLANIDs []int
    existingVLANIDs, err := cmdRunner.VLANOnIntf(deviceConfig.ETHINTFName)
    if err != nil {
        return fmt.Errorf("unable to fetch the existing VLAN with error (%v), may need to reboot the device.", err)
    }

    resetIntf := false
    newVLANIDs := ocutil.VLANIDs(officeConfig)
    if ocutil.VLANChanged(existingVLANIDs, newVLANIDs) {
        log.Infof("VLAN changes (%v -> %v) on interface %s.", existingVLANIDs, newVLANIDs ,deviceConfig.ETHINTFName)
        changedVLANIDs = existingVLANIDs
        resetIntf = true
    } else {
        log.Infof("No VLAN change on interface %s.", deviceConfig.ETHINTFName)
    }

    // Clean up the existing configuration.
    service.CleanupConfig(deviceConfig.ETHINTFName, changedVLANIDs)
    
    // Wait for link to be available again.
    time.Sleep(5 * time.Second)

    // Process the incoming configuraiton.
    if err = service.ApplyConfig(officeConfig, resetIntf, deviceConfig.Hostname, deviceConfig.ETHINTFName,
                                 deviceConfig.WLANINTFName); err != nil {
        return err
    }
    log.Info("Device configuration succeeded.")

    // Save the succeeded config file.
    if err := syscmd.SaveToFile(runFolder, apConfigFileName, configString); err != nil {
        return err
    }
    log.Info("Saved the configuration to file.")
    return nil
}

