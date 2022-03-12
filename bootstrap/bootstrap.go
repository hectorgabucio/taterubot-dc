package bootstrap

import (
	"context"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/infrastructure"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/kelseyhightower/envconfig"
)

func Run() error {

	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return err
	}
	l := localizations.New(cfg.Language, "en")

	ctx, srv := infrastructure.NewServer(context.Background(), l, cfg)
	return srv.Run(ctx)
}
