package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/mwdev22/booking/booking/gen/bookingpb"
	"github.com/mwdev22/grpclib/grpcclient"
)

var (
	ErrInvalidDateFormat = fmt.Errorf("invalid date format")
	ErrCouldNotCreate    = fmt.Errorf("could not create booking")
	ErrNotFound          = fmt.Errorf("booking not found")
	ErrCancelFailed      = fmt.Errorf("could not cancel booking")
)

type Client struct {
	c          bookingpb.BookingServiceClient
	grpcClient *grpcclient.Client
}

func NewClient(ctx context.Context, addr string, opts ...grpcclient.Option) (*Client, error) {
	c, err := grpcclient.New(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial booking service: %w", err)
	}
	pb := bookingpb.NewBookingServiceClient(c.Conn())
	return &Client{
		c:          pb,
		grpcClient: c,
	}, nil
}

func (bc *Client) Close() error {
	return bc.grpcClient.Close()
}

func (bc *Client) Create(ctx context.Context, userID, placeID string, details *BookingDetails, participantIDs []string) (string, error) {
	pbDetails := &bookingpb.BookingDetails{
		Description: details.Description,
		Extras:      details.Extras,
	}

	req := &bookingpb.CreateBookingRequest{
		UserId:  userID,
		PlaceId: placeID,
		Details: pbDetails,
	}

	resp, err := bc.c.CreateBooking(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCouldNotCreate, err)
	}
	return resp.Id, nil
}

func (bc *Client) GetByID(ctx context.Context, id string) (*Booking, error) {
	resp, err := bc.c.GetBooking(ctx, &bookingpb.GetBookingRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return protoToBooking(resp.Booking), nil
}

func (bc *Client) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Booking, error) {
	resp, err := bc.c.GetBookingsByUserID(ctx, &bookingpb.GetBookingsByUserIdRequest{UserId: userID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	return protoBookingsToDomain(resp.Bookings), nil
}

func (bc *Client) List(ctx context.Context, filters *ListFilters) ([]*Booking, int32, error) {
	req := &bookingpb.ListBookingsRequest{
		Status:  filters.Status,
		Limit:   int32(filters.Limit),
		Offset:  int32(filters.Offset),
		PlaceId: filters.PlaceID,
	}
	resp, err := bc.c.ListBookings(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return protoBookingsToDomain(resp.Bookings), resp.Total, nil
}

func (bc *Client) Cancel(ctx context.Context, bookingID string) error {
	resp, err := bc.c.CancelBooking(ctx, &bookingpb.CancelBookingRequest{Id: bookingID})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCancelFailed, err)
	}
	if !resp.Success {
		return fmt.Errorf("cancel failed: %s", resp.Message)
	}
	return nil
}

func (bc *Client) GetPlace(ctx context.Context, id string) (*Place, error) {
	resp, err := bc.c.GetPlace(ctx, &bookingpb.GetPlaceRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToPlace(resp.Place), nil
}

func (bc *Client) ListPlaces(ctx context.Context, filters *PlaceFilters) ([]*Place, int32, error) {
	req := &bookingpb.ListPlacesRequest{
		Limit:    int32(filters.Limit),
		Offset:   int32(filters.Offset),
		Category: filters.Category,
	}
	resp, err := bc.c.ListPlaces(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	places := make([]*Place, len(resp.Places))
	for i, pbPlace := range resp.Places {
		places[i] = protoToPlace(pbPlace)
	}
	return places, resp.Total, nil
}

func (bc *Client) SearchPlaces(ctx context.Context, query, category string, limit int) ([]*Place, error) {
	resp, err := bc.c.SearchPlaces(ctx, &bookingpb.SearchPlacesRequest{
		Query:    query,
		Category: category,
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}
	places := make([]*Place, len(resp.Places))
	for i, pbPlace := range resp.Places {
		places[i] = protoToPlace(pbPlace)
	}
	return places, nil
}

func protoToBooking(pb *bookingpb.Booking) *Booking {
	return &Booking{
		ID:      pb.Id,
		UserID:  pb.UserId,
		PlaceID: pb.PlaceId,
		Place:   protoToPlace(pb.Place),
		Details: &BookingDetails{
			Description: pb.Details.Description,
			Extras:      pb.Details.Extras,
		},
		Status:     StatusFromProto(pb.Status),
		CreatedAt:  pb.CreatedAt.AsTime().Format(time.RFC3339),
		TotalPrice: pb.TotalPrice,
	}
}

func protoBookingsToDomain(pbs []*bookingpb.Booking) []*Booking {
	bookings := make([]*Booking, len(pbs))
	for i, pb := range pbs {
		bookings[i] = protoToBooking(pb)
	}
	return bookings
}

func protoToPlace(pb *bookingpb.Place) *Place {
	return &Place{
		ID:            pb.Id,
		Name:          pb.Name,
		Address:       pb.Address,
		Description:   pb.Description,
		Category:      pb.Category,
		Facilities:    pb.Facilities,
		BasePrice:     pb.BasePrice,
		AvailableFrom: pb.AvailableFrom.AsTime().Format(time.RFC3339),
		AvailableTo:   pb.AvailableTo.AsTime().Format(time.RFC3339),
	}
}
