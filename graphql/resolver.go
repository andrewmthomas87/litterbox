package graphql

import (
	"context"
	"github.com/jmoiron/sqlx"

	"github.com/andrewmthomas87/litterbox/graphql/models"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Db *sqlx.DB
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Me(ctx context.Context) (*models.Me, error) {
	userID := ctx.Value("user_id").(string)
	me := models.Me{}
	if err := r.Db.Get(&me, "SELECT email, name FROM users WHERE id=?", userID); err != nil {
		return nil, err
	}

	return &me, nil
}
