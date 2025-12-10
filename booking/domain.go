package booking

type Booking struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	PlaceID        string          `json:"place_id,omitempty"`
	Place          *Place          `json:"place,omitempty"`
	Details        *BookingDetails `json:"details"`
	Status         string          `json:"status"`
	CreatedAt      string          `json:"created_at"`
	ParticipantIDs []string        `json:"participant_ids,omitempty"`
	TotalPrice     float64         `json:"total_price,omitempty"`
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
	BasePrice     float64  `json:"base_price,omitempty"`
	AvailableFrom string   `json:"available_from,omitempty"`
	AvailableTo   string   `json:"available_to,omitempty"`
}

type ListFilters struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Status string `json:"status,omitempty"`
}

type PlaceFilters struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Category string `json:"category,omitempty"`
}
