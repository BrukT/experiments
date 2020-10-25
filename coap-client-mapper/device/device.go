/*
-- get time from endpoint --
*/

package device

import (
	"bytes"
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/brukt/experiments/coap-client-mapper/common/structs"
	"github.com/plgd-dev/go-coap/v2/udp"
)

var MinimumPeriod int = 1
var Locker sync.Mutex

var ResourceState structs.EndPointState
var MapperConfig structs.MapperState

var PORT = ":5683" //coap udp
var Path string = "/time"

// GetResourceState gets the resouce state from the EndPoint device
func GetResourceState() (string, error) {
	co, err := udp.Dial(MapperConfig.Address + PORT)
	if err != nil {
		return "", err
	}
	defer co.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	resp, err := co.Get(ctx, Path)
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)

	if _, err = io.Copy(buffer, resp.Body()); err != nil {
		log.Println(err)
	}
	return buffer.String(), nil
}

func SetAddress(addr string) {
	MapperConfig.Address = addr
}
