package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/webhook"
	webhookSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/webhook"
	webhookStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/webhook"
)

var WebhookWireSet = wire.NewSet(
	webhook.NewController,
	wire.Bind(new(webhookSvc.Interface), new(*webhookSvc.Service)),
	webhookSvc.NewWebhookSvc,
	wire.Bind(new(webhookStore.Interface), new(*webhookStore.Store)),
	webhookStore.NewWebhookStore,
)
