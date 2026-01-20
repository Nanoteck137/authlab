package core

import (
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

func (app *BaseApp) AuthService() *service.AuthService {
	return app.authService
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

	app.authService = service.NewAuthService(app.db, app.config)
	// TODO(patrik): This should be a worker
	go app.authService.CleanRoutine()

	return nil
}

func NewBaseApp(config *config.Config) *BaseApp {
	return &BaseApp{
		config: config,
	}
}
