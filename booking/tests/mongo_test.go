//go:build integration

package tests

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/mwdev22/booking/booking"
	"github.com/mwdev22/booking/booking/store"
	mongo "github.com/mwdev22/database/mongo"
	"github.com/testcontainers/testcontainers-go"
)

var db *mongo.MongoDB

var bookingRepo *store.MongoBookingRepository

func generateUserID() uuid.UUID {
	return uuid.New()
}

func generateBooking(userID string) *booking.Booking {
	return &booking.Booking{
		UserID:  userID,
		PlaceID: "place1",
		Details: &booking.BookingDetails{
			Description: fmt.Sprintf("desc %s", userID),
			Extras:      map[string]string{"extra": "field"},
		},
		Status: booking.StatusCreated,
	}
}

func generateBookings(count int) []*booking.Booking {
	bookings := make([]*booking.Booking, 0, count)
	for i := 0; i < count; i++ {
		bookings = append(bookings, generateBooking(generateUserID().String()))
	}
	return bookings
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	port := "27017/tcp"

	c, err := testcontainers.Run(ctx,
		"mongo:latest",
		testcontainers.WithExposedPorts(port),
	)
	if err != nil {
		log.Fatalf("couldn't start mongo image: %s", err)
	}

	defer c.Terminate(ctx)
	host, err := c.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := c.MappedPort(ctx, nat.Port(port))
	if err != nil {
		log.Fatal(err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port())
	db = mongo.New(uri)
	if err := db.Connect("booking"); err != nil {
		log.Fatal(err)
	}

	bookingRepo = store.NewMongoBookingRepository(db.Db, "bookings")

	m.Run()

}

func TestMongoBookingRepository_Create(t *testing.T) {
	ctx := context.Background()
	defer bookingRepo.Collection().Drop(ctx)

	tests := []struct {
		name          string
		booking       *booking.Booking
		expectedError bool
	}{
		{
			name: "#1 - OK - create booking",
			booking: &booking.Booking{
				UserID:  "user1",
				PlaceID: "place1",
				Details: &booking.BookingDetails{
					Description: "Test booking",
					Extras:      map[string]string{"wifi": "true"},
				},
				Status: booking.StatusCreated,
			},
			expectedError: false,
		},
		{
			name: "#2 - OK - create booking with minimal data",
			booking: &booking.Booking{
				UserID:  "user2",
				PlaceID: "place2",
				Details: &booking.BookingDetails{},
				Status:  booking.StatusPending,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := bookingRepo.Create(ctx, tt.booking)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if id == "" {
					t.Error("expected non-empty ID")
				}
			}
		})
	}
}

func TestMongoBookingRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	defer bookingRepo.Collection().Drop(ctx)

	// existing booking
	testBooking := &booking.Booking{
		UserID:  "user1",
		PlaceID: "place1",
		Details: &booking.BookingDetails{
			Description: "Test booking",
			Extras:      map[string]string{"wifi": "true"},
		},
		Status: booking.StatusConfirmed,
	}
	id, err := bookingRepo.Create(ctx, testBooking)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	testBooking.ID = id

	tests := []struct {
		name          string
		bookingID     string
		expectedError bool
		validate      func(*testing.T, *booking.Booking)
	}{
		{
			name:          "#1 - OK - get existing booking",
			bookingID:     testBooking.ID,
			expectedError: false,
			validate: func(t *testing.T, b *booking.Booking) {
				if b.ID != testBooking.ID {
					t.Errorf("expected ID %s, got %s", testBooking.ID, b.ID)
				}
				if b.UserID != testBooking.UserID {
					t.Errorf("expected UserID %s, got %s", testBooking.UserID, b.UserID)
				}
				if b.Status != testBooking.Status {
					t.Errorf("expected Status %v, got %v", testBooking.Status, b.Status)
				}
			},
		},
		{
			name:          "#2 - FAIL - booking not found",
			bookingID:     "nonexistent-id",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bookingRepo.GetByID(ctx, tt.bookingID)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestMongoBookingRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	defer bookingRepo.Collection().Drop(ctx)

	// multiple bookings for a specific user
	// one for different user
	userID := "user-test"
	bookings := []*booking.Booking{
		{
			UserID:  userID,
			PlaceID: "place1",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusConfirmed,
		},
		{
			UserID:  userID,
			PlaceID: "place2",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusPending,
		},
		{
			UserID:  userID,
			PlaceID: "place3",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusConfirmed,
		},
		{
			UserID:  "different-user",
			PlaceID: "place4",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusCreated,
		},
	}

	for _, b := range bookings {
		if _, err := bookingRepo.Create(ctx, b); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	tests := []struct {
		name          string
		userID        string
		filters       *booking.ListFilters
		expectedCount int
		expectedError bool
	}{
		{
			name:          "#1 - OK - get all bookings for user",
			userID:        userID,
			filters:       &booking.ListFilters{},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name:   "#2 - OK - get bookings with limit",
			userID: userID,
			filters: &booking.ListFilters{
				Limit: 2,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:   "#3 - OK - get bookings with offset",
			userID: userID,
			filters: &booking.ListFilters{
				Limit:  2,
				Offset: 1,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:   "#4 - OK - filter by status",
			userID: userID,
			filters: &booking.ListFilters{
				Status: booking.StatusConfirmed.String(),
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:          "#5 - OK - no bookings for user",
			userID:        "nonexistent-user",
			filters:       &booking.ListFilters{},
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := bookingRepo.GetByUserID(ctx, tt.userID, tt.filters)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("expected %d bookings, got %d", tt.expectedCount, len(result))
				}
				if total < tt.expectedCount {
					t.Errorf("expected total >= %d, got %d", tt.expectedCount, total)
				}
			}
		})
	}
}

func TestMongoBookingRepository_List(t *testing.T) {
	ctx := context.Background()
	defer bookingRepo.Collection().Drop(ctx)

	// multiple bookings
	bookings := []*booking.Booking{
		{
			UserID:  "user1",
			PlaceID: "place1",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusConfirmed,
		},
		{
			UserID:  "user2",
			PlaceID: "place1",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusPending,
		},
		{
			UserID:  "user3",
			PlaceID: "place2",
			Details: &booking.BookingDetails{},
			Status:  booking.StatusConfirmed,
		},
	}

	for _, b := range bookings {
		if _, err := bookingRepo.Create(ctx, b); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}

	tests := []struct {
		name          string
		filters       *booking.ListFilters
		expectedCount int
		expectedError bool
	}{
		{
			name:          "#1 - OK - list all bookings",
			filters:       &booking.ListFilters{},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "#2 - OK - list with limit",
			filters: &booking.ListFilters{
				Limit: 2,
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "#3 - OK - filter by place",
			filters: &booking.ListFilters{
				PlaceID: "place1",
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "#4 - OK - filter by status",
			filters: &booking.ListFilters{
				Status: booking.StatusConfirmed.String(),
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "#5 - OK - combine filters",
			filters: &booking.ListFilters{
				PlaceID: "place1",
				Status:  booking.StatusConfirmed.String(),
			},
			expectedCount: 1,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := bookingRepo.List(ctx, tt.filters)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("expected %d bookings, got %d", tt.expectedCount, len(result))
				}
				if total < tt.expectedCount {
					t.Errorf("expected total >= %d, got %d", tt.expectedCount, total)
				}
			}
		})
	}
}

func TestMongoBookingRepository_Cancel(t *testing.T) {
	ctx := context.Background()
	defer bookingRepo.Collection().Drop(ctx)

	testBooking := &booking.Booking{
		UserID:  "user1",
		PlaceID: "place1",
		Details: &booking.BookingDetails{},
		Status:  booking.StatusConfirmed,
	}
	id, err := bookingRepo.Create(ctx, testBooking)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	testBooking.ID = id

	tests := []struct {
		name          string
		bookingID     string
		expectedError bool
		validate      func(*testing.T)
	}{
		{
			name:          "#1 - OK - cancel existing booking",
			bookingID:     testBooking.ID,
			expectedError: false,
			validate: func(t *testing.T) {
				result, err := bookingRepo.GetByID(ctx, testBooking.ID)
				if err != nil {
					t.Errorf("failed to get booking: %v", err)
					return
				}
				if result.Status != booking.StatusCancelled {
					t.Errorf("expected status %v, got %v", booking.StatusCancelled, result.Status)
				}
			},
		},
		{
			name:          "#2 - FAIL - cancel with invalid ID",
			bookingID:     "nonexistent-id",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bookingRepo.Cancel(ctx, tt.bookingID)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t)
				}
			}
		})
	}
}
