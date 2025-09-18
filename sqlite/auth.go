package sqlite

import (
	"context"
	"errors"

	"maragu.dev/glue/sql"

	"app/model"
)

func (d *Database) GetUser(ctx context.Context, id model.UserID) (model.User, error) {
	var u model.User
	if err := d.H.Get(ctx, &u, `select * from users where id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, model.ErrorUserNotFound
		}
		return u, err
	}

	return u, nil
}

func (d *Database) IsUserActive(ctx context.Context, id model.UserID) (bool, error) {
	var active bool
	query := `select active from users where id = ?`
	if err := d.H.Get(ctx, &active, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, model.ErrorUserNotFound
		}
		return false, err
	}
	return active, nil
}

func (d *Database) GetPermissions(ctx context.Context, id model.UserID) ([]model.Permission, error) {
	var permissions []model.Permission
	query := `
		select distinct rp.permission
		from users_roles ur
			join roles_permissions rp on ur.role = rp.role
		where ur.user_id = ?
		`

	if err := d.H.Select(ctx, &permissions, query, id); err != nil {
		return nil, err
	}

	return permissions, nil
}
