package service

import (
	"context"
	"time"

	"github.com/mwdev22/booking/booking"
	"github.com/mwdev22/booking/booking/gen/bookingpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrInvalidDateFormat = booking.ErrInvalidDateFormat
	ErrCouldNotCreate    = booking.ErrCouldNotCreate
	ErrNotFound          = booking.ErrNotFound
	ErrCancelFailed      = booking.ErrCancelFailed
)

type BookingService struct {
	bookingpb.UnimplementedBookingServiceServer
	bookings booking.BookingStore
	places   booking.PlaceStore
}

func NewBookingService(bookings booking.BookingStore, places booking.PlaceStore) *BookingService {
	return &BookingService{
		bookings: bookings,
		places:   places,
	}
}

func (bs *BookingService) Create(ctx context.Context, req *bookingpb.CreateBookingRequest) (bookingpb.CreateBookingResponse, error) {

	id, err := bs.bookings.Create(ctx, &booking.Booking{
		UserID:  req.UserId,
		PlaceID: req.PlaceId,
		Details: &booking.BookingDetails{
			Description: req.Details.Description,
			Extras:      req.Details.Extras,
		},
	})

	return bookingpb.CreateBookingResponse{Id: id}, err
}

func (bs *BookingService) GetByID(ctx context.Context, req *bookingpb.GetBookingRequest) (*bookingpb.GetBookingResponse, error) {
	b, err := bs.bookings.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	created, err := parseDate(b.CreatedAt)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	pbBooking := &bookingpb.Booking{
		Id:      b.ID,
		UserId:  b.UserID,
		PlaceId: b.PlaceID,
		Details: &bookingpb.BookingDetails{
			Description: b.Details.Description,
			Extras:      b.Details.Extras,
		},
		Status:     b.Status.ToProto(),
		CreatedAt:  timestamppb.New(created),
		TotalPrice: b.TotalPrice,
	}

	return &bookingpb.GetBookingResponse{Booking: pbBooking}, nil
}

func (bs *BookingService) GetByUserID(ctx context.Context, req *bookingpb.GetBookingsByUserIdRequest) (*bookingpb.GetBookingsByUserIdResponse, error) {
	bookings, err := bs.bookings.GetByUserID(ctx, req.UserId, &booking.ListFilters{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	})
	if err != nil {
		return nil, err
	}

	pbBookings := make([]*bookingpb.Booking, len(bookings))
	for i, b := range bookings {
		pbBookings[i] = &bookingpb.Booking{
			Id:      b.ID,
			UserId:  b.UserID,
			PlaceId: b.PlaceID,
			Details: &bookingpb.BookingDetails{
				Description: b.Details.Description,
				Extras:      b.Details.Extras,
			},
			Status:     b.Status.ToProto(),
			CreatedAt:  timestamppb.Now(),
			TotalPrice: b.TotalPrice,
		}
	}

	return &bookingpb.GetBookingsByUserIdResponse{Bookings: pbBookings}, nil
}

func (bs *BookingService) ListBookings(ctx context.Context, req *bookingpb.ListBookingsRequest) (*bookingpb.ListBookingsResponse, error) {
	filters := &booking.ListFilters{
		Limit:   int(req.Limit),
		Offset:  int(req.Offset),
		Status:  req.Status,
		PlaceID: req.PlaceId,
	}
	bookings, total, err := bs.bookings.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	pbBookings := make([]*bookingpb.Booking, len(bookings))
	for i, b := range bookings {
		created, err := parseDate(b.CreatedAt)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		pbBookings[i] = &bookingpb.Booking{
			Id:      b.ID,
			UserId:  b.UserID,
			PlaceId: b.PlaceID,
			Details: &bookingpb.BookingDetails{
				Description: b.Details.Description,
				Extras:      b.Details.Extras,
			},
			Status:     b.Status.ToProto(),
			CreatedAt:  timestamppb.New(created),
			TotalPrice: b.TotalPrice,
		}
	}

	return &bookingpb.ListBookingsResponse{Bookings: pbBookings, Total: int32(total)}, nil
}

func (bs *BookingService) CancelBooking(ctx context.Context, req *bookingpb.CancelBookingRequest) (*bookingpb.CancelBookingResponse, error) {
	err := bs.bookings.Cancel(ctx, req.Id)
	return &bookingpb.CancelBookingResponse{}, err
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}
