/*
All the publish subscibe commincations with the edgecore are done through this package

*/

package helper

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brukt/experiments/coap-client-mapper/common/structs"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	DeviceETPrefix        = "$hw/events/device/"
	TwinETGetSuffix       = "/twin/get"
	TwinETGetResultSuffix = "/twin/get/result"
	TwinETUpdateSuffix    = "/twin/update"
	TwinETCloudSyncSuffix = "/twin/cloud_updated"
)

var TwinResult structs.DeviceTwinResult
var TokenClient Token
var Client MQTT.Client
var TwinAttributes []string
var ClientID string

//Token interface to validate the MQTT connection.
type Token interface {
	Wait() bool
	WaitTimeout(time.Duration) bool
	Error() error
}

// added to make the client ID related with the fog name
func init() {
	var err error
	ClientID, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
}

//MqttConnect function felicitates the MQTT connection
func MqttConnect(serverIP string) {
	server := "tcp://" + serverIP + ":1883"
	ClientOpts := MQTT.NewClientOptions().AddBroker(server).SetClientID(ClientID).SetCleanSession(true)
	Client = MQTT.NewClient(ClientOpts)
	if TokenClient = Client.Connect(); TokenClient.Wait() && TokenClient.Error() != nil {
		log.Printf("client.Connect() Error is %s", TokenClient.Error())
	}
}

//TwinSubscribe function subscribes  the device twin information through the MQTT broker
func TwinSubscribe(deviceID string, done chan bool) {
	getTwinResult := DeviceETPrefix + deviceID + TwinETGetResultSuffix
	// log.Println("Starting subscriber with topic", getTwinResult)
	TokenClient = Client.Subscribe(getTwinResult, 0, OnTwinMessageReceived)
	if TokenClient.Wait() && TokenClient.Error() != nil {
		log.Printf("subscribe() Error in device twin result get  is: %s", TokenClient.Error())
	}
	//log.Println("Subscriber start: Twin topic ", getTwinResult)
	for {
		//log.Println("IN subscriber loop: Twin Result", TwinResult)
		time.Sleep(1 * time.Second)
		if TwinResult.Twin != nil {
			for k := range TwinResult.Twin {
				TwinAttributes = append(TwinAttributes, k)
			}
			done <- true
			break
		}
	}
}

// OnTwinMessageReceived callback function which is called when message is received
// At the initiation stage of running the module
func OnTwinMessageReceived(client MQTT.Client, message MQTT.Message) {
	err := json.Unmarshal(message.Payload(), &TwinResult)
	if err != nil {
		log.Printf("Error in unmarshalling:  %s", err)
	}
}

// GetTwin function is used to get the device twin details from the edge
// First publish to the /get topic and then listen to /get/result topic until it gets reply
func GetTwin(updateMessage structs.DeviceTwinUpdate, deviceID string) {
	getTwin := DeviceETPrefix + deviceID + TwinETGetSuffix
	twinUpdateBody, err := json.Marshal(updateMessage)
	if err != nil {
		log.Printf("Error in marshalling: %s", err)
	}
	// log.Println("Published content", string(twinUpdateBody))
	TokenClient = Client.Publish(getTwin, 0, false, twinUpdateBody)
	if TokenClient.Wait() && TokenClient.Error() != nil {
		log.Printf("client.publish() Error in device twin get  is: %s ", TokenClient.Error())
	}
}

// Send the device state update to the edge node
// Specifically the temperature state
func UpdateTwinValue(deviceID string, updateMessage structs.DeviceTwinUpdate) error {
	twinUpdateBody, err := json.Marshal(updateMessage)
	if err != nil {
		log.Printf("Error in marshalling: %s", err)
	}
	deviceTwinUpdate := DeviceETPrefix + deviceID + TwinETUpdateSuffix
	TokenClient = Client.Publish(deviceTwinUpdate, 0, false, twinUpdateBody)
	if TokenClient.Wait() && TokenClient.Error() != nil {
		log.Printf("client.publish() Error in device twin update is %s", TokenClient.Error())
	}
	return nil
}

//SyncToCloud function syncs the updated device twin information to the cloud
func SyncToCloud(deviceID string, updateMessage structs.DeviceTwinUpdate) {
	deviceTwinResultUpdate := DeviceETPrefix + deviceID + TwinETCloudSyncSuffix
	twinUpdateBody, err := json.Marshal(updateMessage)
	if err != nil {
		log.Printf("Error in marshalling: %s", err)
	}
	TokenClient = Client.Publish(deviceTwinResultUpdate, 0, false, twinUpdateBody)
	if TokenClient.Wait() && TokenClient.Error() != nil {
		log.Printf("client.publish() Error in device twin update is: %s", TokenClient.Error())
	}
}

// Create the update information to send to the twin module
func CreateActualUpdateMessage(resourceState structs.EndPointState, mapperState structs.MapperState) structs.DeviceTwinUpdate {
	var deviceTwinUpdateMessage structs.DeviceTwinUpdate
	period := strconv.Itoa(mapperState.Period)
	actualMap := map[string]*structs.MsgTwin{"temperature": {Actual: &structs.TwinValue{Value: &resourceState.Temperature}, Metadata: &structs.TypeMetadata{Type: "Updated"}}, "address": {Actual: &structs.TwinValue{Value: &mapperState.Address}, Metadata: &structs.TypeMetadata{Type: "Updated"}}, "period": {Actual: &structs.TwinValue{Value: &period}, Metadata: &structs.TypeMetadata{Type: "Updated"}}}

	deviceTwinUpdateMessage.Twin = actualMap

	return deviceTwinUpdateMessage
}

func Disconnect() {
	Client.Disconnect(200)
}
