package gateway

// here is the place where u define all domain objects and interfaces used by application

type Booking struct {
	ID      string
	UserID  string
	Details BookingDetails
	Status  string
}

type BookingDetails struct {
	Location     string
	Date         string
	Participants int
	Extras       map[string]interface{}
}

type Payment struct {
	ID       string
	Amount   float64
	Currency string
	Method   string
	Status   string
}

type User struct {
	ID    string
	Email string
	Name  string
	Phone string
}

type PaymentsClient interface {
	Create(amount float64, currency, method string) error
	GetByID(paymentID string) ([]byte, error)
	GetStatus(paymentID string) (string, error)
}

type BookingClient interface {
	Create(b *Booking) (string, error)
	GetByID(id string) ([]byte, error)
	GetByUserID(userID string) ([]byte, error)
	Cancel(bookingID string) error
}
