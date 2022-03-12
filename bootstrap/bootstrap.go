package bootstrap

import (
	"context"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/server"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/kelseyhightower/envconfig"
	"os"
)

func Run() error {

	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return err
	}
	l := localizations.New(cfg.Language, "en")

	baseFilePath := cfg.BasePath
	if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
		err := os.Mkdir(baseFilePath, 0750)
		if err != nil {
			return err
		}
	}

	ctx, srv := server.NewServer(context.Background(), l, cfg)
	return srv.Run(ctx)
}
