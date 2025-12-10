package store

import (
	"context"
	"time"

	"github.com/mwdev22/booking/booking"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoBookingRepository struct {
	c *mongo.Collection
}

func NewMongoBookingRepository(db *mongo.Database, collectionName string) *MongoBookingRepository {
	return &MongoBookingRepository{
		c: db.Collection(collectionName),
	}
}

func (br *MongoBookingRepository) Collection() *mongo.Collection {
	return br.c
}

func (br *MongoBookingRepository) Create(ctx context.Context, booking *booking.Booking) (string, error) {
	res, err := br.c.InsertOne(ctx, bson.M{
		"id":         booking.ID,
		"user_id":    booking.UserID,
		"place_id":   booking.PlaceID,
		"details":    booking.Details,
		"status":     booking.Status,
		"created_at": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	return res.InsertedID.(string), nil
}

func (br *MongoBookingRepository) GetByID(ctx context.Context, id string) (*booking.Booking, error) {
	var booking booking.Booking
	err := br.c.FindOne(ctx, bson.M{"id": id}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (br *MongoBookingRepository) GetByUserID(ctx context.Context, userID string, filters *booking.ListFilters) ([]*booking.Booking, int, error) {
	query := bson.M{"user_id": userID}
	total := 0
	opts := options.Find()
	if filters != nil {
		if filters.Status != "" {
			query["status"] = filters.Status
		}
		if filters.Limit > 0 {
			opts.SetLimit(int64(filters.Limit))
		}
		if filters.Offset > 0 {
			opts.SetSkip(int64(filters.Offset))
		}
	}

	cursor, err := br.c.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var bookings []*booking.Booking
	for cursor.Next(ctx) {
		var b booking.Booking
		if err := cursor.Decode(&b); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, &b)
	}
	total = len(bookings) + cursor.RemainingBatchLength()
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}

func (br *MongoBookingRepository) List(ctx context.Context, filters *booking.ListFilters) ([]*booking.Booking, int, error) {
	query := bson.M{}
	total := 0
	opts := options.Find()
	if filters != nil {
		if filters.Status != "" {
			query["status"] = filters.Status
		}
		if filters.PlaceID != "" {
			query["place_id"] = filters.PlaceID
		}
		if filters.Limit > 0 {
			opts.SetLimit(int64(filters.Limit))
		}
		if filters.Offset > 0 {
			opts.SetSkip(int64(filters.Offset))
		}
	}

	cursor, err := br.c.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var bookings []*booking.Booking
	for cursor.Next(ctx) {
		var b booking.Booking
		if err := cursor.Decode(&b); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, &b)
	}
	total = len(bookings) + cursor.RemainingBatchLength()
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}

func (br *MongoBookingRepository) Cancel(ctx context.Context, bookingID string) error {
	_, err := br.c.UpdateOne(ctx, bson.M{"id": bookingID}, bson.M{
		"$set": bson.M{"status": booking.StatusCancelled}},
	)
	return err
}
