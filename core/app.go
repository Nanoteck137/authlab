package core

import (
	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/database"
	"github.com/nanoteck137/authlab/service"
	"github.com/nanoteck137/authlab/types"
)

// Inspiration from Pocketbase: https://github.com/pocketbase/pocketbase
// File: https://github.com/pocketbase/pocketbase/blob/master/core/app.go
type App interface {
	DB() *database.Database
	Config() *config.Config

	AuthService() (*service.AuthService, error)

	WorkDir() types.WorkDir

	Bootstrap() error
}
