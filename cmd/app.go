package main

import (
	"encoding/json"

	"gerrit.o-ran-sc.org/r/ric-plt/alarm-go.git/alarm"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type App struct {
	Stats map[string]xapp.Counter
}

var (
	funId               = int64(1)
	actionId            = int64(1)
	actionType          = "report"
	subsequestActioType = "continue"
	timeToWait          = "w10ms"
	xappEventInstanceID = int64(1234)
	hPort               = int64(8080)
	rPort               = int64(4560)
	clientEndpoint      = clientmodel.SubscriptionParamsClientEndpoint{Host: "service-ricxapp-kpm-rmr.ricxapp", HTTPPort: &hPort, RMRPort: &rPort}
)

func main() {
	metrics := []xapp.CounterOpts{
		{
			Name: "RICIndicationRx",
			Help: "Total number of RIC Indication message received",
		},
	}

	app := App{
		Stats: xapp.Metric.RegisterCounterGroup(metrics, "kpm_app")}

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

	nbList := e.getnbList()

	for _, nb := range nbList {
		e.sendSubscription(nb.InventoryName)
	}
}

func (e *App) handleRICIndication(ranName string, r *xapp.RMRParams) {
	xapp.Logger.Info("handleRICIndication", ranName, "\tparams: ", *r)
	e.Stats["RICIndicationRx"].Inc()
}

func (e *App) sendSubscription(meid string) {

	xapp.Logger.Info("sending subscription request for meid : %s", meid)

	subscriptionParams := clientmodel.SubscriptionParams{
		ClientEndpoint: &clientEndpoint,
		Meid:           &meid,
		RANFunctionID:  &funId,
		SubscriptionDetails: clientmodel.SubscriptionDetailsList([]*clientmodel.SubscriptionDetail{
			{
				ActionToBeSetupList: clientmodel.ActionsToBeSetup{
					&clientmodel.ActionToBeSetup{
						ActionDefinition: clientmodel.ActionDefinition([]int64{1, 2, 3, 4}),
						ActionID:         &actionId,
						ActionType:       &actionType,
						SubsequentAction: &clientmodel.SubsequentAction{
							SubsequentActionType: &subsequestActioType,
							TimeToWait:           &timeToWait,
						},
					},
				},
				EventTriggers:       clientmodel.EventTriggerDefinition([]int64{1, 2, 3, 4}),
				XappEventInstanceID: &xappEventInstanceID,
			},
		}),
	}

	b, err := json.MarshalIndent(subscriptionParams, "", "  ")

	if err != nil {
		xapp.Logger.Error("Json marshaling failed : %s", err)
		return
	}

	xapp.Logger.Info("*****body: %s ", string(b))

	resp, err := xapp.Subscription.Subscribe(&subscriptionParams)

	if err != nil {
		xapp.Logger.Error("subscription failed (%s) with error: %s", meid, err)

		// subscription failed, raise alarm
		err := xapp.Alarm.Raise(8086, alarm.SeverityCritical, meid, "subscriptionFailed")
		if err != nil {
			xapp.Logger.Error("Raising alarm failed with error %v", err)
		}

		return
	}
	xapp.Logger.Info("Successfully subcription done (%s), subscription id : %s", meid, *resp.SubscriptionID)
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

func (e *App) getnbList() []*xapp.RNIBNbIdentity {
	var nbs []*xapp.RNIBNbIdentity

	if enbs, err := e.getEnbList(); err == nil {
		nbs = append(nbs, enbs...)
	}

	if gnbs, err := e.getGnbList(); err == nil {
		nbs = append(nbs, gnbs...)
	}

	return nbs
}

func (e *App) getEnbList() ([]*xapp.RNIBNbIdentity, error) {
	enbs, err := xapp.Rnib.GetListEnbIds()

	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}

	xapp.Logger.Info("List of connected eNBs :")
	for index, enb := range enbs {
		xapp.Logger.Info("%d. enbid: %s", index+1, enb.InventoryName)
	}
	return enbs, nil
}

func (e *App) getGnbList() ([]*xapp.RNIBNbIdentity, error) {
	gnbs, err := xapp.Rnib.GetListGnbIds()

	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}

	xapp.Logger.Info("List of connected gNBs :")
	for index, gnb := range gnbs {
		xapp.Logger.Info("%d. gnbid : %s", index+1, gnb.InventoryName)
	}
	return gnbs, nil
}
