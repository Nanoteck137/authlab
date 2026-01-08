package database

import (
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/nanoteck137/pyrin/ember"
)

type UserIdentity struct {
	Provider   string `db:"provider"`
	ProviderId string `db:"provider_id"`

	UserId string `db:"user_id"`

	Created int64 `db:"created"`
	Updated int64 `db:"updated"`
}

func UserIdentityQuery() *goqu.SelectDataset {
	query := dialect.From("user_identities").
		Select(
			"user_identities.provider",
			"user_identities.provider_id",

			"user_identities.user_id",

			"user_identities.created",
			"user_identities.updated",
		)

	return query
}

// func (db DB) GetAllUserIdentitys(ctx context.Context) ([]UserIdentity, error) {
// 	query := UserIdentityQuery()
//
// 	return ember.Multiple[UserIdentity](db.db, ctx, query)
// }

func (db DB) GetUserIdentity(ctx context.Context, provider, providerId string) (UserIdentity, error) {
	query := UserIdentityQuery().
		Where(
			goqu.I("user_identities.provider").Eq(provider),
			goqu.I("user_identities.provider_id").Eq(providerId),
		)

	return ember.Single[UserIdentity](db.db, ctx, query)
}

type CreateUserIdentityParams struct {
	Provider   string
	ProviderId string

	UserId string

	Created int64
	Updated int64
}

func (db DB) CreateUserIdentity(ctx context.Context, params CreateUserIdentityParams) error {
	t := time.Now().UnixMilli()
	created := params.Created
	updated := params.Updated

	if created == 0 && updated == 0 {
		created = t
		updated = t
	}

	query := dialect.
		Insert("user_identities").
		Rows(goqu.Record{
			"provider":    params.Provider,
			"provider_id": params.ProviderId,

			"user_id": params.UserId,

			"created": created,
			"updated": updated,
		})

	_, err := db.db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
