package graphql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"

	"github.com/andrewmthomas87/litterbox/graphql/models"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Db *sqlx.DB
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }

func (r *queryResolver) Me(ctx context.Context) (*models.Me, error) {
	userID := ctx.Value("user_id").(string)
	me := models.Me{}
	if err := r.Db.GetContext(ctx, &me, "SELECT email, name, stage FROM users WHERE id=?", userID); err != nil {
		return nil, err
	}

	return &me, nil
}

func (r *queryResolver) StorageItems(ctx context.Context) ([]*models.StorageItem, error) {
	var storageItems []*models.StorageItem
	if err := r.Db.SelectContext(ctx, &storageItems, "SELECT id, name, price, description FROM storageItems ORDER BY name ASC"); err != nil {
		return nil, err
	}

	return storageItems, nil
}

func (r *queryResolver) MyStorageItemQuantities(ctx context.Context) ([]*models.StorageItemQuantity, error) {
	userID := ctx.Value("user_id").(string)

	var storageItemQuantities []*models.StorageItemQuantity
	if err := r.Db.SelectContext(ctx, &storageItemQuantities, "SELECT itemID, quantity FROM storageItemQuantities WHERE userID=?", userID); err != nil {
		return nil, err
	}

	return storageItemQuantities, nil
}

func (r *queryResolver) PickupTimeSlots(ctx context.Context) ([]*models.TimeSlot, error) {
	var timeSlots []*models.TimeSlot
	if err := r.Db.SelectContext(ctx, &timeSlots, "SELECT id, date, DATE_FORMAT(startTime, '%H:%i') AS startTime, DATE_FORMAT(endTime, '%H:%i') AS endTime, capacity, count FROM pickupTimeSlots ORDER BY date ASC, startTime ASC, endTime ASC"); err != nil {
		return nil, err
	}

	return timeSlots, nil
}

func (r *queryResolver) MyPickupTimeSlot(ctx context.Context) (*models.TimeSlot, error) {
	userID := ctx.Value("user_id").(string)

	var timeSlotID int
	if err := r.Db.GetContext(ctx, &timeSlotID, "SELECT timeSlotID FROM pickupTimeSlotSelections WHERE userID=?", userID); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var timeSlot models.TimeSlot
	if err := r.Db.GetContext(ctx, &timeSlot, "SELECT id, date, DATE_FORMAT(startTime, '%H:%i') AS startTime, DATE_FORMAT(endTime, '%H:%i') AS endTime, capacity, count FROM pickupTimeSlots WHERE id=?", timeSlotID); err != nil {
		return nil, err
	}

	return &timeSlot, nil
}

func (r *mutationResolver) SaveInformation(ctx context.Context, information models.InformationInput) (*models.Me, error) {
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

	if _, err := r.Db.ExecContext(ctx, "UPDATE users SET name=?, onCampus=?, building=?, address=?, onCampusFuture=?, stage=? WHERE id=?", information.Name, information.OnCampus, building, address, information.OnCampusFuture, models.StageDefault, userID); err != nil {
		return nil, err
	}

	me := models.Me{}
	if err := r.Db.GetContext(ctx, &me, "SELECT email, name, stage FROM users WHERE id=?", userID); err != nil {
		return nil, err
	}

	return &me, nil
}

func (r *mutationResolver) GenerateReservationSession(ctx context.Context) (string, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Name:        stripe.String("Reservation fee"),
				Description: stripe.String("TODO"),
				Amount:      stripe.Int64(2500),
				Currency:    stripe.String(string(stripe.CurrencyUSD)),
				Quantity:    stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(viper.GetString("stripe.successURL")),
		CancelURL:  stripe.String(viper.GetString("stripe.cancelURL")),
	}

	session, err := session.New(params)
	if err != nil {
		return "", err
	}

	return session.ID, nil
}

func (r *mutationResolver) SetStorageItemQuantities(ctx context.Context, quantities models.StorageItemQuantitiesInput) ([]*models.StorageItemQuantity, error) {
	userID := ctx.Value("user_id").(string)

	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM storageItemQuantities WHERE userID=?", userID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	for _, quantity := range quantities.Quantities {
		if _, err := tx.ExecContext(ctx, "INSERT INTO storageItemQuantities (userID, itemID, quantity) VALUES (?, ?, ?)", userID, quantity.ItemID, quantity.Quantity); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	var storageItemQuantities []*models.StorageItemQuantity
	if err := tx.SelectContext(ctx, &storageItemQuantities, "SELECT itemID, quantity FROM storageItemQuantities WHERE userID=?", userID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return storageItemQuantities, nil
}

func (r *mutationResolver) SelectPickupTimeSlot(ctx context.Context, id int) (*models.TimeSlot, error) {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var timeSlot models.TimeSlot
	if err := tx.GetContext(ctx, &timeSlot, "SELECT capacity, count FROM pickupTimeSlots WHERE id=?", id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if timeSlot.Count >= timeSlot.Capacity {
		_ = tx.Rollback()
		return nil, errors.New("time slot at capacity")
	}

	if _, err := tx.ExecContext(ctx, "UPDATE pickupTimeSlots SET count=? WHERE id=?", timeSlot.Count+1, id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	userID := ctx.Value("user_id").(string)

	var currentTimeSlotID int
	if err := tx.GetContext(ctx, &currentTimeSlotID, "SELECT timeSlotID FROM pickupTimeSlotSelections WHERE userID=?", userID); err != nil && err != sql.ErrNoRows {
		_ = tx.Rollback()
		return nil, err
	}

	if currentTimeSlotID > 0 {
		if err := tx.GetContext(ctx, &timeSlot, "SELECT capacity, count FROM pickupTimeSlots WHERE id=?", currentTimeSlotID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, "UPDATE pickupTimeSlots SET count=? WHERE id=?", timeSlot.Count-1, currentTimeSlotID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	res, err := tx.ExecContext(ctx, "UPDATE pickupTimeSlotSelections SET timeSlotID=? WHERE userID=?", id, userID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if affected == 0 {
		if _, err := tx.ExecContext(ctx, "INSERT INTO pickupTimeSlotSelections (userID, timeSlotID) VALUES (?, ?)", userID, id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	var updatedTimeSlot models.TimeSlot
	if err := r.Db.GetContext(ctx, &updatedTimeSlot, "SELECT id, date, DATE_FORMAT(startTime, '%H:%i') AS startTime, DATE_FORMAT(endTime, '%H:%i') AS endTime, capacity, count FROM pickupTimeSlots WHERE id=?", id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &updatedTimeSlot, nil
}
