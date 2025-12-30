package core

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/database"
	"github.com/nanoteck137/authlab/types"
)

var _ App = (*BaseApp)(nil)

type BaseApp struct {
	db     *database.Database
	config *config.Config
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

	return nil
}

func NewBaseApp(config *config.Config) *BaseApp {
	return &BaseApp{
		config: config,
	}
}
