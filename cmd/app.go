package main

import (
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/RafaelRochaS/xapp-kpm-go/cmd/handlers"
)

type App struct {
	NBsHandler      handlers.NBsHandler
	MessagesHandler handlers.MessagesHandler
}

func main() {

	nbsHandler := handlers.NewNBsImpl()
	messagesHandler := handlers.NewMessagesHandler()

	app := App{
		NBsHandler:      nbsHandler,
		MessagesHandler: messagesHandler,
	}

	app.Run()
}

func (e *App) Run() {

	xapp.Logger.SetMdc("App", "0.0.1")
	xapp.AddConfigChangeListener(e.ConfigChangeHandler)
	xapp.SetReadyCB(e.xAppStartCB, true)
	waitForSdl := xapp.Config.GetBool("db.waitForSdl")

	xapp.RunWithParams(e, waitForSdl)
}

func (e *App) ConfigChangeHandler(f string) {
	xapp.Logger.Info("Config file changed", f)
}

func (e *App) xAppStartCB(d interface{}) {
	xapp.Logger.Info("xApp ready call back received", d)

	err := e.NBsHandler.SubscribeToAll()

	if err != nil {
		xapp.Logger.Error("Failed to subscribe to all NBs: %v", err)
	}
}

func (e *App) Consume(msg *xapp.RMRParams) (err error) {
	id := xapp.Rmr.GetRicMessageName(msg.Mtype)

	xapp.Logger.Info("Message received: name=%s meid=%s subId=%d txid=%s len=%d", id, msg.Meid.RanName, msg.SubId, msg.Xid, msg.PayloadLen)

	switch id {
	case "RIC_HEALTH_CHECK_REQ":
		xapp.Logger.Info("Received health check request")

	case "RIC_INDICATION":
		xapp.Logger.Info("Received RIC Indication message")
		e.MessagesHandler.HandleRICIndication(msg)

	case "RIC_SUBSCRIPTION_RESPONSE":
		xapp.Logger.Info("Received RIC Subscription Response message")
		e.MessagesHandler.HandleSubscriptionResponse(msg)

	default:
		xapp.Logger.Info("Unknown message type '%d', discarding", msg.Mtype)
	}

	defer func() {
		xapp.Rmr.Free(msg.Mbuf)
		msg.Mbuf = nil
	}()
	return
}
