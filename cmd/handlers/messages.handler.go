package handlers

import "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

const RicIndicationRx = "RICIndicationRx"

type MessagesHandler interface {
	HandleRICIndication(msg *xapp.RMRParams)
	HandleSubscriptionResponse(msg *xapp.RMRParams)
	GetStats() map[string]xapp.Counter
}

type MessagesImpl struct {
	Stats map[string]xapp.Counter
}

func NewMessagesHandler() *MessagesImpl {
	metrics := []xapp.CounterOpts{
		{
			Name: RicIndicationRx,
			Help: "Total number of RIC Indication message received",
		},
	}

	m := MessagesImpl{Stats: xapp.Metric.RegisterCounterGroup(metrics, "kpm_app")}

	return &m
}

func (m *MessagesImpl) HandleRICIndication(msg *xapp.RMRParams) {
	xapp.Logger.Info("HandleRICIndication", msg.Meid.RanName, "\tparams: ", *msg)
	m.Stats[RicIndicationRx].Inc()
}

func (m *MessagesImpl) HandleSubscriptionResponse(msg *xapp.RMRParams) {
	xapp.Logger.Info("HandleSubscriptionResponse", msg.Meid.RanName, "\tparams: ", *msg)
}

func (m *MessagesImpl) GetStats() map[string]xapp.Counter {
	return m.Stats
}
