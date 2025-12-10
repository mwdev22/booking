package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mwdev22/booking/booking"
	"github.com/mwdev22/booking/booking/gen/bookingpb"
	"github.com/mwdev22/booking/booking/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookingService_Create(t *testing.T) {
	tests := []struct {
		name          string
		req           *bookingpb.CreateBookingRequest
		bookingStore  func(*mocks.MockBookingStore)
		expectedID    string
		expectedError bool
	}{
		{
			name: "#1 - OK",
			req: &bookingpb.CreateBookingRequest{
				UserId:  "user123",
				PlaceId: "place456",
				Details: &bookingpb.BookingDetails{
					Description: "Test booking",
					Extras:      map[string]string{"parking": "true"},
				},
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("Create", mock.Anything, mock.MatchedBy(func(b *booking.Booking) bool {
					return b.UserID == "user123" && b.PlaceID == "place456"
				})).Return("booking789", nil)
			},
			expectedID:    "booking789",
			expectedError: false,
		},
		{
			name: "#2 - FAIL - store internal",
			req: &bookingpb.CreateBookingRequest{
				UserId:  "user123",
				PlaceId: "place456",
				Details: &bookingpb.BookingDetails{
					Description: "Test booking",
				},
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("Create", mock.Anything, mock.Anything).Return("", errors.New("creation failed"))
			},
			expectedID:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBookingStore(t)
			tt.bookingStore(s)

			service := NewBookingService(s, mocks.NewMockPlaceStore(t))
			resp, err := service.Create(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, resp.Id)
			}
		})
	}
}

func TestBookingService_GetByID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)

	tests := []struct {
		name          string
		req           *bookingpb.GetBookingRequest
		bookingStore  func(*mocks.MockBookingStore)
		expectedError bool
		validate      func(*testing.T, *bookingpb.GetBookingResponse)
	}{
		{
			name: "#1 - OK",
			req: &bookingpb.GetBookingRequest{
				Id: "booking123",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByID", mock.Anything, "booking123").Return(&booking.Booking{
					ID:        "booking123",
					UserID:    "user456",
					PlaceID:   "place789",
					Status:    booking.StatusConfirmed,
					CreatedAt: now,
					Details: &booking.BookingDetails{
						Description: "Test booking",
						Extras:      map[string]string{"wifi": "true"},
					},
					TotalPrice: 150.50,
				}, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *bookingpb.GetBookingResponse) {
				assert.NotNil(t, resp.Booking)
				assert.Equal(t, "booking123", resp.Booking.Id)
				assert.Equal(t, "user456", resp.Booking.UserId)
				assert.Equal(t, "place789", resp.Booking.PlaceId)
				assert.Equal(t, bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED, resp.Booking.Status)
				assert.Equal(t, float32(150.50), resp.Booking.TotalPrice)
			},
		},
		{
			name: "#2 - FAIL - not found",
			req: &bookingpb.GetBookingRequest{
				Id: "nonexistent",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
		{
			name: "#3 - FAIL - invalid date format",
			req: &bookingpb.GetBookingRequest{
				Id: "booking123",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByID", mock.Anything, "booking123").Return(&booking.Booking{
					ID:        "booking123",
					UserID:    "user456",
					PlaceID:   "place789",
					Status:    booking.StatusConfirmed,
					CreatedAt: "invalid-date",
					Details: &booking.BookingDetails{
						Description: "Test booking",
					},
				}, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBookingStore(t)
			tt.bookingStore(s)

			service := NewBookingService(s, mocks.NewMockPlaceStore(t))
			resp, err := service.GetByID(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}

func TestBookingService_GetByUserID(t *testing.T) {
	tests := []struct {
		name          string
		req           *bookingpb.GetBookingsByUserIdRequest
		bookingStore  func(*mocks.MockBookingStore)
		expectedError bool
		validate      func(*testing.T, *bookingpb.GetBookingsByUserIdResponse)
	}{
		{
			name: "#1 - OK - with pagination",
			req: &bookingpb.GetBookingsByUserIdRequest{
				UserId: "user123",
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByUserID", mock.Anything, "user123", &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return([]*booking.Booking{
					{
						ID:      "booking1",
						UserID:  "user123",
						PlaceID: "place1",
						Status:  booking.StatusConfirmed,
						Details: &booking.BookingDetails{
							Description: "Booking 1",
						},
						TotalPrice: 100.0,
					},
					{
						ID:      "booking2",
						UserID:  "user123",
						PlaceID: "place2",
						Status:  booking.StatusPending,
						Details: &booking.BookingDetails{
							Description: "Booking 2",
						},
						TotalPrice: 200.0,
					},
				}, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *bookingpb.GetBookingsByUserIdResponse) {
				assert.Len(t, resp.Bookings, 2)
				assert.Equal(t, "booking1", resp.Bookings[0].Id)
				assert.Equal(t, "booking2", resp.Bookings[1].Id)
				assert.Equal(t, bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED, resp.Bookings[0].Status)
				assert.Equal(t, bookingpb.BookingStatus_BOOKING_STATUS_PENDING, resp.Bookings[1].Status)
			},
		},
		{
			name: "#2 - OK - empty result",
			req: &bookingpb.GetBookingsByUserIdRequest{
				UserId: "user999",
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByUserID", mock.Anything, "user999", &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return([]*booking.Booking{}, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *bookingpb.GetBookingsByUserIdResponse) {
				assert.Len(t, resp.Bookings, 0)
			},
		},
		{
			name: "#3 - FAIL - store error",
			req: &bookingpb.GetBookingsByUserIdRequest{
				UserId: "user123",
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("GetByUserID", mock.Anything, "user123", &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBookingStore(t)
			tt.bookingStore(s)

			service := NewBookingService(s, mocks.NewMockPlaceStore(t))
			resp, err := service.GetByUserID(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}

func TestBookingService_ListBookings(t *testing.T) {
	now := time.Now().Format(time.RFC3339)

	tests := []struct {
		name          string
		req           *bookingpb.ListBookingsRequest
		bookingStore  func(*mocks.MockBookingStore)
		expectedError bool
		validate      func(*testing.T, *bookingpb.ListBookingsResponse)
	}{
		{
			name: "#1 - OK - with filters",
			req: &bookingpb.ListBookingsRequest{
				Limit:   20,
				Offset:  0,
				Status:  "CONFIRMED",
				PlaceId: "place123",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("List", mock.Anything, &booking.ListFilters{
					Limit:   20,
					Offset:  0,
					Status:  "CONFIRMED",
					PlaceID: "place123",
				}).Return([]*booking.Booking{
					{
						ID:        "booking1",
						UserID:    "user1",
						PlaceID:   "place123",
						Status:    booking.StatusConfirmed,
						CreatedAt: now,
						Details: &booking.BookingDetails{
							Description: "Booking 1",
						},
						TotalPrice: 100.0,
					},
				}, 1, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *bookingpb.ListBookingsResponse) {
				assert.Len(t, resp.Bookings, 1)
				assert.Equal(t, int32(1), resp.Total)
				assert.Equal(t, "booking1", resp.Bookings[0].Id)
				assert.Equal(t, bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED, resp.Bookings[0].Status)
			},
		},
		{
			name: "#2 - OK - multiple bookings",
			req: &bookingpb.ListBookingsRequest{
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("List", mock.Anything, &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return([]*booking.Booking{
					{
						ID:        "booking1",
						UserID:    "user1",
						PlaceID:   "place1",
						Status:    booking.StatusCreated,
						CreatedAt: now,
						Details:   &booking.BookingDetails{Description: "Booking 1"},
					},
					{
						ID:        "booking2",
						UserID:    "user2",
						PlaceID:   "place2",
						Status:    booking.StatusPending,
						CreatedAt: now,
						Details:   &booking.BookingDetails{Description: "Booking 2"},
					},
				}, 2, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *bookingpb.ListBookingsResponse) {
				assert.Len(t, resp.Bookings, 2)
				assert.Equal(t, int32(2), resp.Total)
			},
		},
		{
			name: "#3 - FAIL - invalid date format",
			req: &bookingpb.ListBookingsRequest{
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("List", mock.Anything, &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return([]*booking.Booking{
					{
						ID:        "booking1",
						UserID:    "user1",
						PlaceID:   "place1",
						Status:    booking.StatusCreated,
						CreatedAt: "invalid-date",
						Details:   &booking.BookingDetails{Description: "Booking 1"},
					},
				}, 1, nil)
			},
			expectedError: true,
		},
		{
			name: "#4 - FAIL - store error",
			req: &bookingpb.ListBookingsRequest{
				Limit:  10,
				Offset: 0,
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("List", mock.Anything, &booking.ListFilters{
					Limit:  10,
					Offset: 0,
				}).Return(nil, 0, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBookingStore(t)
			tt.bookingStore(s)

			service := NewBookingService(s, mocks.NewMockPlaceStore(t))
			resp, err := service.ListBookings(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}

func TestBookingService_CancelBooking(t *testing.T) {
	tests := []struct {
		name          string
		req           *bookingpb.CancelBookingRequest
		bookingStore  func(*mocks.MockBookingStore)
		expectedError bool
	}{
		{
			name: "#1 - OK",
			req: &bookingpb.CancelBookingRequest{
				Id: "booking123",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("Cancel", mock.Anything, "booking123").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "#2 - FAIL - cancellation failed",
			req: &bookingpb.CancelBookingRequest{
				Id: "booking456",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("Cancel", mock.Anything, "booking456").Return(errors.New("cancellation failed"))
			},
			expectedError: true,
		},
		{
			name: "#3 - FAIL - not found",
			req: &bookingpb.CancelBookingRequest{
				Id: "nonexistent",
			},
			bookingStore: func(bs *mocks.MockBookingStore) {
				bs.On("Cancel", mock.Anything, "nonexistent").Return(errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBookingStore(t)
			tt.bookingStore(s)

			service := NewBookingService(s, mocks.NewMockPlaceStore(t))
			_, err := service.CancelBooking(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBookingStatus_Conversions(t *testing.T) {
	tests := []struct {
		status      booking.BookingStatus
		protoStatus bookingpb.BookingStatus
		stringRep   string
	}{
		{booking.StatusCreated, bookingpb.BookingStatus_BOOKING_STATUS_CREATED, "CREATED"},
		{booking.StatusPaymentRequired, bookingpb.BookingStatus_BOOKING_STATUS_PAYMENT_REQUIRED, "PAYMENT_REQUIRED"},
		{booking.StatusPending, bookingpb.BookingStatus_BOOKING_STATUS_PENDING, "PENDING"},
		{booking.StatusConfirmed, bookingpb.BookingStatus_BOOKING_STATUS_CONFIRMED, "CONFIRMED"},
		{booking.StatusCancelled, bookingpb.BookingStatus_BOOKING_STATUS_CANCELLED, "CANCELLED"},
	}

	for _, tt := range tests {
		t.Run(tt.stringRep, func(t *testing.T) {
			// Test ToProto
			assert.Equal(t, tt.protoStatus, tt.status.ToProto())

			// Test String
			assert.Equal(t, tt.stringRep, tt.status.String())

			// Test FromProto
			assert.Equal(t, tt.status, booking.StatusFromProto(tt.protoStatus))
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name          string
		dateStr       string
		expectedError bool
	}{
		{
			name:          "valid RFC3339 date",
			dateStr:       time.Now().Format(time.RFC3339),
			expectedError: false,
		},
		{
			name:          "invalid date format",
			dateStr:       "2024-01-01",
			expectedError: true,
		},
		{
			name:          "empty string",
			dateStr:       "",
			expectedError: true,
		},
		{
			name:          "random string",
			dateStr:       "not-a-date",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDate(tt.dateStr)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
