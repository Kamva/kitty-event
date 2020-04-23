package main

import (
	"fmt"
	"github.com/Kamva/gutil"
	"github.com/Kamva/hexa"
	"github.com/Kamva/hexa-event"
	"github.com/Kamva/hexa-event/pulsar"
	"github.com/Kamva/hexa/db/mgmadapter"
	"github.com/Kamva/hexa/hexalogger"
	"github.com/Kamva/hexa/hexatranslator"
	"github.com/apache/pulsar-client-go/pulsar"
	"time"
)

type HelloPayload struct {
	Hello string `json:"hello"`
}

var clientURL = "pulsar://localhost:6650"
var format = "%s"
var t = hexatranslator.NewEmptyDriver()
var l = hexalogger.NewPrinterDriver()
var userExporter = hexa.NewUserExporterImporter(mgmadapter.EmptyID)
var ctxExporterImporter = hexa.NewCtxExporterImporter(userExporter, l, t)

func main() {
	send()
	time.Sleep(2 * time.Second)
	receive()
}

func send() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: clientURL,
	})
	gutil.PanicErr(err)
	defer client.Close()

	var t = hexatranslator.NewEmptyDriver()
	var l = hexalogger.NewPrinterDriver()

	emitter, err := hexapulsar.NewEmitter(client, hexapulsar.EmitterOptions{
		ProducerGenerator: hexapulsar.DefaultProducerGenerator(format),
		CtxExporter:       ctxExporterImporter,
		Marshaller:        hevent.NewJsonMarshaller(),
	})
	gutil.PanicErr(err)

	defer func() {
		gutil.PanicErr(emitter.Close())
	}()

	event := &hevent.Event{
		Payload: HelloPayload{Hello: "from Hexa2:)"},
		Channel: "hexa-test",
		Key:     "test-key",
	}

	res, err := emitter.Emit(hexa.NewCtx(nil, "test-correlation-id", "en", hexa.NewGuest(), l, t), event)
	gutil.PanicErr(err)
	fmt.Println(res)
	fmt.Println("message sent :)")
}

func receive() {
	channel := hexapulsar.DefaultSubscriptionItemPack(format, "hexa-test", &HelloPayload{}, sayHello)

	// From here for all receivers is same.
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: clientURL,
	})
	gutil.PanicErr(err)

	receiver, err := hexapulsar.NewReceiver(client, hexapulsar.ReceiverOptions{
		CtxImporter:              ctxExporterImporter,
		ConsumerOptionsGenerator: hexapulsar.NewConsumerOptionsGenerator([]hexapulsar.ConsumerOptionsItem{channel.ConsumerOptions}),
	})
	gutil.PanicErr(err)

	defer func() {
		err := receiver.Close()
		gutil.PanicErr(err)
	}()

	err = receiver.SubscribeMulti(channel.SubscriptionItem.Channel, channel.SubscriptionItem.PayloadInstance, channel.SubscriptionItem.Handler)
	err = receiver.Start()
	gutil.PanicErr(err)
}

func sayHello(hc hevent.HandlerContext, c hexa.Context, m hevent.Message, err error) {
	gutil.PanicErr(err)
	fmt.Println("running hello handler.")
	fmt.Println(m.MessageHeader)
	p := m.Payload.(*HelloPayload)
	fmt.Println(p.Hello)
	fmt.Println(c.User().Type())
	hc.Ack()
}

var _ hevent.EventHandler = sayHello
