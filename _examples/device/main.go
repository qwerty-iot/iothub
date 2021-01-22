package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/qwerty-iot/iothub/iotdevice"
	iotmqtt "github.com/qwerty-iot/iothub/iotdevice/transport/mqtt"
)

func main() {
	c, err := iotdevice.NewFromConnectionString(
		iotmqtt.New(), "HostName=tartabit-dev.azure-devices.net;DeviceId=testclient;SharedAccessKey=IHEnjn0ODkNCrLx2bP0DhCCl9EZSKbw4dsZH3E6RpfM=",
	)
	if err != nil {
		log.Fatal(err)
	}

	// connect to the iothub
	if err = c.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}

	// send a device-to-cloud message
	if err = c.SendEvent(context.Background(), []byte("hello")); err != nil {
		log.Fatal(err)
	}

	fu, err := c.FileUpload(context.Background(), "test.text")
	if err != nil {
		log.Fatal("sas error: " + err.Error())
	}

	fil, _ := ioutil.ReadFile("_examples/device/test.text")

	err = fu.Upload(fil)
	if err != nil {
		log.Fatal("upload error: " + err.Error())
	}
	err = fu.Complete()
	if err != nil {
		log.Fatal("upload error: " + err.Error())
	}
}
