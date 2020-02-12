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
	if err := r.Db.Get(&me, "SELECT email, name, stage FROM users WHERE id=?", userID); err != nil {
		return nil, err
	}

	return &me, nil
}

func (r *queryResolver) SaveInformation(ctx context.Context, information models.InformationInput) (*models.Me, error) {
	if err := validateInformation(information); err != nil {
		return nil, err
	}

	var (
		building models.Building
		address  string
	)
	if information.OnCampus {
		building = models.BuildingLookup[information.Building]
		address = ""
	} else {
		building = 0
		address = information.Address
	}

	userID := ctx.Value("user_id").(string)

	if _, err := r.Db.Exec("UPDATE users SET name=?, onCampus=?, building=?, address=?, onCampusFuture=?, stage=? WHERE id=?", information.Name, information.OnCampus, building, address, information.OnCampusFuture, models.StageDefault, userID); err != nil {
		return nil, err
	}

	me := models.Me{}
	if err := r.Db.Get(&me, "SELECT email, name, stage FROM users WHERE id=?", userID); err != nil {
		return nil, err
	}

	return &me, nil
}
