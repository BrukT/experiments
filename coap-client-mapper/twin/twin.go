/*
-- update the cloud state --
send the device state update to the EdgeNode
specifically publish to a topic formed by "coap-mapper"

the publishing period should be similar with the querying period
*/

package twin

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/brukt/experiments/coap-client-mapper/common/helper"
	"github.com/brukt/experiments/coap-client-mapper/common/structs"
	"github.com/brukt/experiments/coap-client-mapper/device"
)

var MqttIP string = "127.0.0.1" // Replace this IP with 172.16.4.147, 172.16.2.65
var DeviceID string = "coap-time-client"

func init() {
	helper.MqttConnect(MqttIP)
}

// Start runs the main thread of the twin module
func Start(stop chan bool, wg *sync.WaitGroup) {
	terminate := false
	var actualPeriod time.Duration = 1 // the period in unit of second
	var doNothing = false              // if the period is set to zero don't contact the coap server

	// Listen for new updates from the cloud about Mapperconfig and heater state
	// update the state of MapperConfig and Heater accordingly
	// check cloud configurations and configure yours
	log.Printf("Starting Ip Address: %s, Period: %d, Heater: %s", device.MapperConfig.Address, device.MapperConfig.Period, device.ResourceState.HeaterStatus)

	for {
		if device.MapperConfig.Period == 0 {
			actualPeriod = 1
			doNothing = true
		} else if device.MapperConfig.Period > 0 {
			actualPeriod = time.Duration(device.MapperConfig.Period)
			doNothing = false
		}

		// check cloud configurations and configure yours
		// needed to change from a sleeping state to a non sleeping state
		if err := GetAndSetState(); err != nil {
			log.Println(err)
		}

		if !doNothing {
			// Update the Resource copy inside the edgenode and also in k8s cloud indirectly
			updateMessage := helper.CreateActualUpdateMessage(device.ResourceState, device.MapperConfig)
			log.Printf("Syncing to edge")
			if err := helper.UpdateTwinValue(DeviceID, updateMessage); err != nil {
				log.Println(err)
			}
			log.Printf("Syncing to cloud")
			helper.SyncToCloud(DeviceID, updateMessage)
		}

		select {
		case <-stop:
			terminate = true
		case <-time.After(actualPeriod * time.Second):

		}
		if terminate {
			//helper.Disconnect()
			break
		}
	}
	log.Printf("Terminating twin module")
	wg.Done()
}

// GetAndSetState returns the current state of the resouce from the cloud and set the the mapper copy
// doesn't include the temperature state
func GetAndSetState() error {
	var updateMessage structs.DeviceTwinUpdate
	var Wg = make(chan bool, 1)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // put a timeout on the subscribet to stop incase of errors.
	defer cancel()
	go helper.TwinSubscribe(DeviceID, Wg)
	// if the subscription takes more than 5 second, stop for error
	helper.GetTwin(updateMessage, DeviceID)
	select {
	case <-Wg:
	case <-ctx.Done():
		log.Printf("TWIN subsription is impossible")
	}
	address := *(helper.TwinResult.Twin["address"]).Expected.Value
	period := *(helper.TwinResult.Twin["period"]).Expected.Value
	log.Printf("TWIN IP Addr: %s", address)
	log.Printf("TWIN Period : %s", period)

	device.SetAddress(address)

	return nil
}
