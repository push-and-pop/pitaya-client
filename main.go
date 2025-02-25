package main

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/conn/message"
	"github.com/topfreegames/pitaya/v3/pkg/client"
)

var port = flag.Int("port", 3250, "the port to listen")

func main() {
	flag.Parse()
	c := client.New(logrus.InfoLevel)
	err := c.ConnectTo(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logrus.Panicln(err)
		return
	}
	defer c.Disconnect()

	TestHandlerCallToFront(c)
}

func TestHandlerCallToFront(c *client.Client) {
	tables := []struct {
		req  string
		data []byte
		resp []byte
	}{
		{"connector.testsvc.testrequestonlysessionreturnsptr", []byte(``), []byte(`{"code":200,"msg":"hello"}`)},
		{"connector.testsvc.testrequestonlysessionreturnsptrnil", []byte(``), []byte(`{"code":"PIT-000","msg":"reply must not be null"}`)},
		{"connector.testsvc.testrequestonlysessionreturnsrawnil", []byte(``), []byte(`{"code":"PIT-000","msg":"reply must not be null"}`)},
		{"connector.testsvc.testrequestreturnsptr", []byte(`{"msg":"good"}`), []byte(`{"code":200,"msg":"good"}`)},
		{"connector.testsvc.testrequestreturnsraw", []byte(`{"msg":"good"}`), []byte(`good`)},
		{"connector.testsvc.testrequestreceivereturnsraw", []byte(`woow`), []byte(`woow`)},
		{"connector.testsvc.nonexistenthandler", []byte(`woow`), []byte(`{"code":"PIT-404","msg":"pitaya/handler: connector.testsvc.nonexistenthandler not found"}`)},
		{"connector.testsvc.testrequestreturnserror", []byte(`woow`), []byte(`{"code":"PIT-555","msg":"somerror"}`)},
	}

	for _, table := range tables {
		_, err := c.SendRequest(table.req, table.data)
		if err != nil {
			logrus.Panicln(err)
			return
		}

		msg := ShouldEventuallyReceive(c.IncomingMsgChan).(*message.Message)
		logrus.Infoln(msg)
	}

	for {
		time.Sleep(time.Second)
	}
}

func ShouldEventuallyReceive(c interface{}, timeouts ...time.Duration) interface{} {
	v := reflect.ValueOf(c)

	timeout := time.After(500 * time.Millisecond)

	if len(timeouts) > 0 {
		timeout = time.After(timeouts[0])
	}

	recvChan := make(chan reflect.Value)

	go func() {
		v, ok := v.Recv()
		if ok {
			recvChan <- v
		}
	}()

	select {
	case <-timeout:
		logrus.Panicln(errors.New("timed out waiting for channel to receive"))
	case a := <-recvChan:
		return a.Interface()
	}
	return nil
}
