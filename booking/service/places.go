package service

import (
	"context"

	"github.com/mwdev22/booking/booking/gen/bookingpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (bs *BookingService) SearchPlaces(ctx context.Context, req *bookingpb.SearchPlacesRequest) (*bookingpb.SearchPlacesResponse, error) {
	places, err := bs.places.Search(ctx, req.Query, req.Category, int(req.Limit))
	if err != nil {
		return nil, err
	}

	pbPlaces := make([]*bookingpb.Place, len(places))
	for i, place := range places {
		from, err := parseDate(place.AvailableFrom)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		to, err := parseDate(place.AvailableTo)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		pbPlaces[i] = &bookingpb.Place{
			Id:            place.ID,
			Name:          place.Name,
			Address:       place.Address,
			Description:   place.Description,
			Category:      place.Category,
			Facilities:    place.Facilities,
			BasePrice:     place.BasePrice,
			AvailableFrom: timestamppb.New(from),
			AvailableTo:   timestamppb.New(to),
		}
	}

	return &bookingpb.SearchPlacesResponse{Places: pbPlaces}, nil
}

func (bs *BookingService) GetPlace(ctx context.Context, req *bookingpb.GetPlaceRequest) (*bookingpb.GetPlaceResponse, error) {
	place, err := bs.places.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	from, err := parseDate(place.AvailableFrom)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	to, err := parseDate(place.AvailableTo)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	pbPlace := &bookingpb.Place{
		Id:            place.ID,
		Name:          place.Name,
		Address:       place.Address,
		Description:   place.Description,
		Category:      place.Category,
		Facilities:    place.Facilities,
		BasePrice:     place.BasePrice,
		AvailableFrom: timestamppb.New(from),
		AvailableTo:   timestamppb.New(to),
	}

	return &bookingpb.GetPlaceResponse{Place: pbPlace}, nil
}
