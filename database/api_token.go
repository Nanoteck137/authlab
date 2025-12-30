package database

import (
	"context"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/nanoteck137/authlab/tools/utils"
	"github.com/nanoteck137/pyrin/ember"
)

type ApiToken struct {
	Id     string `db:"id"`
	UserId string `db:"user_id"`

	Name string `db:"name"`

	Created int64 `db:"created"`
	Updated int64 `db:"updated"`
}

func ApiTokenQuery() *goqu.SelectDataset {
	query := dialect.From("api_tokens").
		Select(
			"api_tokens.id",
			"api_tokens.user_id",

			"api_tokens.name",

			"api_tokens.updated",
			"api_tokens.created",
		).
		Prepared(true)

	return query
}

func (db DB) GetApiTokenById(ctx context.Context, id string) (ApiToken, error) {
	query := ApiTokenQuery().
		Where(goqu.I("api_tokens.id").Eq(id))

	return ember.Single[ApiToken](db.db, ctx, query)
}

func (db DB) GetAllApiTokensForUser(ctx context.Context, userId string) ([]ApiToken, error) {
	query := ApiTokenQuery().
		Where(goqu.I("api_tokens.user_id").Eq(userId))

	return ember.Multiple[ApiToken](db.db, ctx, query)
}

type CreateApiTokenParams struct {
	Id     string
	UserId string
	Name   string

	Created int64
	Updated int64
}

func (db DB) CreateApiToken(ctx context.Context, params CreateApiTokenParams) (ApiToken, error) {
	t := time.Now().UnixMilli()
	created := params.Created
	updated := params.Updated

	if created == 0 && updated == 0 {
		created = t
		updated = t
	}

	id := params.Id
	if id == "" {
		id = utils.CreateApiTokenId()
	}

	query := dialect.Insert("api_tokens").Rows(goqu.Record{
		"id":      id,
		"user_id": params.UserId,

		"name": params.Name,

		"created": created,
		"updated": updated,
	}).
		Returning(
			"api_tokens.id",
			"api_tokens.user_id",

			"api_tokens.name",

			"api_tokens.updated",
			"api_tokens.created",
		)

	return ember.Single[ApiToken](db.db, ctx, query)
}

func (db DB) DeleteApiToken(ctx context.Context, id string) error {
	query := dialect.Delete("api_tokens").
		Where(goqu.I("api_tokens.id").Eq(id))

	_, err := db.db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
