package main

import (
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/RafaelRochaS/xapp-kpm-go/cmd/handlers"
)

type App struct {
	Stats      map[string]xapp.Counter
	NBsHandler handlers.NBsHandler
}

func main() {
	metrics := []xapp.CounterOpts{
		{
			Name: "RICIndicationRx",
			Help: "Total number of RIC Indication message received",
		},
	}

	nbsHandler := handlers.NewNBsImpl()

	app := App{
		Stats:      xapp.Metric.RegisterCounterGroup(metrics, "kpm_app"),
		NBsHandler: nbsHandler,
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

func (e *App) handleRICIndication(ranName string, r *xapp.RMRParams) {
	xapp.Logger.Info("handleRICIndication", ranName, "\tparams: ", *r)
	e.Stats["RICIndicationRx"].Inc()
}

func (e *App) Consume(msg *xapp.RMRParams) (err error) {
	id := xapp.Rmr.GetRicMessageName(msg.Mtype)

	xapp.Logger.Info("Message received: name=%s meid=%s subId=%d txid=%s len=%d", id, msg.Meid.RanName, msg.SubId, msg.Xid, msg.PayloadLen)

	switch id {
	case "RIC_HEALTH_CHECK_REQ":
		xapp.Logger.Info("Received health check request")

	case "RIC_INDICATION":
		xapp.Logger.Info("Received RIC Indication message")
		e.handleRICIndication(msg.Meid.RanName, msg)

	default:
		xapp.Logger.Info("Unknown message type '%d', discarding", msg.Mtype)
	}

	defer func() {
		xapp.Rmr.Free(msg.Mbuf)
		msg.Mbuf = nil
	}()
	return
}
