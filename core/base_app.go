package core

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/database"
	"github.com/nanoteck137/authlab/service"
	"github.com/nanoteck137/authlab/types"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	db     *database.Database
	config *config.Config

	authService *service.AuthService
}

func (app *BaseApp) AuthService() (*service.AuthService, error) {
	err := app.authService.Init(context.Background(), app.config)
	if err != nil {
		return nil, err
	}

	return app.authService, nil
}

func (app *BaseApp) DB() *database.Database {
	return app.db
}

func (app *BaseApp) Config() *config.Config {
	return app.config
}

func (app *BaseApp) WorkDir() types.WorkDir {
	return app.config.WorkDir()
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (app *BaseApp) Bootstrap() error {
	var err error

	workDir := app.config.WorkDir()

	// dirs := []string{}
	//
	// for _, dir := range dirs {
	// 	err = os.Mkdir(dir, 0755)
	// 	if err != nil && !os.IsExist(err) {
	// 		return err
	// 	}
	// }

	app.db, err = database.Open(workDir.DatabaseFile())
	if err != nil {
		return err
	}

	if app.config.RunMigrations {
		err = app.db.RunMigrateUp()
		if err != nil {
			return err
		}
	}

	app.authService = service.NewAuthService(app.db, app.config.JwtSecret)

	return nil
}

func NewBaseApp(config *config.Config) *BaseApp {
	return &BaseApp{
		config: config,
	}
}
