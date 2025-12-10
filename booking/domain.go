package booking

import (
	"context"

	"github.com/mwdev22/booking/booking/gen/bookingpb"
)

type BookingStatus int

const (
	StatusCreated BookingStatus = iota
	StatusPaymentRequired
	StatusPending
	StatusConfirmed
	StatusCancelled
)

func (bs BookingStatus) String() string {
	switch bs {
	case StatusCreated:
		return "CREATED"
	case StatusPaymentRequired:
		return "PAYMENT_REQUIRED"
	case StatusPending:
		return "PENDING"
	case StatusConfirmed:
		return "CONFIRMED"
	case StatusCancelled:
		return "CANCELLED"
	default:
		return ""
	}
}

func (bs BookingStatus) ToProto() bookingpb.BookingStatus {
	switch bs {
	case StatusCreated:
		return bookingpb.BookingStatus_BOOKING_STATUS_CREATED
	case StatusPaymentRequired:
		return bookingpb.BookingStatus_BOOKING_STATUS_PAYMENT_REQUIRED
	case StatusPending:
		return bookingpb.BookingStatus_BOOKING_STATUS_PENDING
	case StatusConfirmed:
		return bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED
	case StatusCancelled:
		return bookingpb.BookingStatus_BOOKING_STATUS_CANCELLED
	default:
		return bookingpb.BookingStatus_BOOKING_STATUS_UNSPECIFIED
	}
}

func StatusFromProto(bs bookingpb.BookingStatus) BookingStatus {
	switch bs {
	case bookingpb.BookingStatus_BOOKING_STATUS_CREATED:
		return StatusCreated
	case bookingpb.BookingStatus_BOOKING_STATUS_PAYMENT_REQUIRED:
		return StatusPaymentRequired
	case bookingpb.BookingStatus_BOOKING_STATUS_PENDING:
		return StatusPending
	case bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED:
		return StatusConfirmed
	case bookingpb.BookingStatus_BOOKING_STATUS_CANCELLED:
		return StatusCancelled
	default:
		return StatusCreated
	}
}

type Booking struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	PlaceID    string          `json:"place_id,omitempty"`
	Place      *Place          `json:"place,omitempty"`
	Details    *BookingDetails `json:"details"`
	Status     BookingStatus   `json:"status"`
	CreatedAt  string          `json:"created_at"`
	TotalPrice float32         `json:"total_price,omitempty"`
}

type BookingDetails struct {
	Description  string            `json:"description,omitempty"`
	Participants int               `json:"participants"`
	Extras       map[string]string `json:"extras,omitempty"`
}

type Place struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	Address       string   `json:"address,omitempty"`
	Description   string   `json:"description,omitempty"`
	Category      string   `json:"category,omitempty"`
	Facilities    []string `json:"facilities,omitempty"`
	BasePrice     float32  `json:"base_price,omitempty"`
	AvailableFrom string   `json:"available_from,omitempty"`
	AvailableTo   string   `json:"available_to,omitempty"`
}

type ListFilters struct {
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Status  string `json:"status,omitempty"`
	PlaceID string `json:"place_id,omitempty"`
}

type PlaceFilters struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Category string `json:"category,omitempty"`
}

type BookingStore interface {
	Create(ctx context.Context, b *Booking) (string, error)
	GetByID(ctx context.Context, id string) (*Booking, error)
	GetByUserID(ctx context.Context, userID string, filters *ListFilters) ([]*Booking, error)
	Cancel(ctx context.Context, bookingID string) error
	List(ctx context.Context, filters *ListFilters) ([]*Booking, int, error)
}

type PlaceStore interface {
	GetByID(ctx context.Context, id string) (*Place, error)
	Search(ctx context.Context, query, category string, limit int) ([]*Place, error)
}
