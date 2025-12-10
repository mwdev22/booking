package service

import (
	"context"

	"github.com/mwdev22/booking/booking"
	"github.com/mwdev22/booking/booking/gen/bookingpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookingService struct {
	bookingpb.UnimplementedBookingServiceServer
	bookings booking.BookingRepository
	places   booking.PlaceRepository
}

func NewBookingService(bookings booking.BookingRepository, places booking.PlaceRepository) *BookingService {
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
			Description:  req.Details.Description,
			Participants: int(req.Details.Participants),
			Extras:       req.Details.Extras,
		},
		ParticipantIDs: req.ParticipantIds,
	})

	return bookingpb.CreateBookingResponse{Id: id}, err
}

func (bs *BookingService) GetByID(ctx context.Context, req *bookingpb.GetBookingRequest) (*bookingpb.GetBookingResponse, error) {
	b, err := bs.bookings.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	pbBooking := &bookingpb.Booking{
		Id:      b.ID,
		UserId:  b.UserID,
		PlaceId: b.PlaceID,
		Details: &bookingpb.BookingDetails{
			Description:  b.Details.Description,
			Participants: int32(b.Details.Participants),
			Extras:       b.Details.Extras,
		},
		Status:         b.Status.ToProto(),
		CreatedAt:      timestamppb.Now(),
		TotalPrice:     b.TotalPrice,
		ParticipantIds: b.ParticipantIDs,
	}

	return &bookingpb.GetBookingResponse{Booking: pbBooking}, nil
}

func (bs *BookingService) GetByUserID(ctx context.Context, req *bookingpb.GetBookingsByUserIdRequest) (*bookingpb.GetBookingsByUserIdResponse, error) {
	bookings, err := bs.bookings.GetByUserID(ctx, req.UserId, &booking.ListFilters{})
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
				Description:  b.Details.Description,
				Participants: int32(b.Details.Participants),
				Extras:       b.Details.Extras,
			},
			Status:         b.Status.ToProto(),
			CreatedAt:      timestamppb.Now(),
			TotalPrice:     b.TotalPrice,
			ParticipantIds: b.ParticipantIDs,
		}
	}

	return &bookingpb.GetBookingsByUserIdResponse{Bookings: pbBookings}, nil
}

func (bs *BookingService) CancelBooking(ctx context.Context, req *bookingpb.CancelBookingRequest) (*bookingpb.CancelBookingResponse, error) {
	err := bs.bookings.Cancel(ctx, req.Id)
	return &bookingpb.CancelBookingResponse{}, err
}

func (bs *BookingService) GetPlace(ctx context.Context, req *bookingpb.GetPlaceRequest) (*bookingpb.GetPlaceResponse, error) {
	place, err := bs.places.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	pbPlace := &bookingpb.Place{
		Id:            place.ID,
		Name:          place.Name,
		Address:       place.Address,
		Description:   place.Description,
		Category:      place.Category,
		Facilities:    place.Facilities,
		BasePrice:     place.BasePrice,
		AvailableFrom: timestamppb.Now(),
		AvailableTo:   timestamppb.Now(),
	}

	return &bookingpb.GetPlaceResponse{Place: pbPlace}, nil
}

func (bs *BookingService) SearchPlaces(ctx context.Context, req *bookingpb.SearchPlacesRequest) (*bookingpb.SearchPlacesResponse, error) {
	places, err := bs.places.Search(ctx, req.Query, req.Category, int(req.Limit))
	if err != nil {
		return nil, err
	}

	pbPlaces := make([]*bookingpb.Place, len(places))
	for i, place := range places {
		pbPlaces[i] = &bookingpb.Place{
			Id:            place.ID,
			Name:          place.Name,
			Address:       place.Address,
			Description:   place.Description,
			Category:      place.Category,
			Facilities:    place.Facilities,
			BasePrice:     place.BasePrice,
			AvailableFrom: timestamppb.Now(),
			AvailableTo:   timestamppb.Now(),
		}
	}

	return &bookingpb.SearchPlacesResponse{Places: pbPlaces}, nil
}
