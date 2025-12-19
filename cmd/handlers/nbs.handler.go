package handlers

import (
	"encoding/json"

	"gerrit.o-ran-sc.org/r/ric-plt/alarm-go.git/alarm"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type NBsHandler interface {
	SubscribeToAll() error
}

type NBsImpl struct {
	Nbs []*xapp.RNIBNbIdentity
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

func NewNBsImpl() *NBsImpl {
	nbsImpl := &NBsImpl{}

	nbsImpl.Nbs = nbsImpl.getNBs()

	return nbsImpl
}

func (n *NBsImpl) getNBs() []*xapp.RNIBNbIdentity {
	var nbs []*xapp.RNIBNbIdentity

	if enbs, err := n.getEnbList(); err == nil {
		nbs = append(nbs, enbs...)
	}

	if gnbs, err := n.getGnbList(); err == nil {
		nbs = append(nbs, gnbs...)
	}

	return nbs
}

func (n *NBsImpl) getEnbList() ([]*xapp.RNIBNbIdentity, error) {
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

func (n *NBsImpl) getGnbList() ([]*xapp.RNIBNbIdentity, error) {
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

func (n *NBsImpl) SubscribeToAll() error {
	var err error

	for _, nb := range n.Nbs {
		err = n.sendSubscription(nb.InventoryName)

		if err != nil {
			return err
		}
	}
	return nil
}

func (n *NBsImpl) sendSubscription(meid string) error {
	xapp.Logger.Info("sending subscription request for meid : %s", meid)

	subscriptionParams := clientmodel.SubscriptionParams{
		ClientEndpoint: &clientEndpoint,
		Meid:           &meid,
		RANFunctionID:  &funId,
		SubscriptionDetails: clientmodel.SubscriptionDetailsList([]*clientmodel.SubscriptionDetail{
			{
				ActionToBeSetupList: clientmodel.ActionsToBeSetup{
					&clientmodel.ActionToBeSetup{
						ActionDefinition: []int64{1, 2, 3, 4},
						ActionID:         &actionId,
						ActionType:       &actionType,
						SubsequentAction: &clientmodel.SubsequentAction{
							SubsequentActionType: &subsequestActioType,
							TimeToWait:           &timeToWait,
						},
					},
				},
				EventTriggers:       []int64{1, 2, 3, 4},
				XappEventInstanceID: &xappEventInstanceID,
			},
		}),
	}

	b, err := json.MarshalIndent(subscriptionParams, "", "  ")

	if err != nil {
		xapp.Logger.Error("Json marshaling failed : %s", err)
		return err
	}

	xapp.Logger.Info("*****body: %s ", string(b))

	resp, err := xapp.Subscription.Subscribe(&subscriptionParams)

	if err != nil {
		xapp.Logger.Error("subscription failed (%s) with error: %s", meid, err)

		err := xapp.Alarm.Raise(8086, alarm.SeverityCritical, meid, "subscriptionFailed")
		if err != nil {
			xapp.Logger.Error("Raising alarm failed with error %v", err)
		}

		return err
	}
	xapp.Logger.Info("Subscription successful (%s), subscription id : %s", meid, *resp.SubscriptionID)

	return nil
}
