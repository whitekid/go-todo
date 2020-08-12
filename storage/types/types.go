package types

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	ErrNotFound         = errors.New("fot found")
	ErrNotAuthenticated = errors.New("not authenticated")
)

const (
	RFC3339FullDate = "2006-01-02" // RFC3339FullDate ...
)

// Interface represent storage abstract layer
type Interface interface {
	UserService() UserService
	TokenService() TokenService
	TodoService() TodoService

	Close()
}

type UserService interface {
	Get(email string) (*User, error)
	Create(user *User) error
	Delete(email string) error
}

// TokenService storage refresh token service,
// Token should be JWT token
type TokenService interface {
	Create(email string, refreshToken string) error
	Get(token string) (string, error)
	Delete(token string) error
}

// TodoService represents todo storage
type TodoService interface {
	Create(email string, item *TodoItem) error

	List(email string) ([]TodoItem, error) // list todo items

	// return ErrNotFound if item not found
	Get(email string, itemID string) (*TodoItem, error)

	// return ErrNotFound if item not found
	Update(email string, item *TodoItem) error

	Delete(email string, itemID string) error
}

// User user informations
type User struct {
	Email string `json:"email" validate:"required,email"`
}

// TodoItem todo item
type TodoItem struct {
	ID      string `json:"id" format:"uuid" example:"628b92ab-6d95-4fbe-b7c6-09cf5cd8941c" validate:"required,uuid"`
	Title   string `json:"title" example:"do something in future" validate:"required"`
	DueDate Date   `json:"due_date" swaggertype:"string" example:"2006-01-02"`
	Rank    int    `json:"rank" format:"int" example:"1"` // rank order
}

func Today() (d Date) {
	d.Time = time.Now().UTC().Truncate(time.Hour * 24)
	return
}

// Date represents json full-date format, supports marshal json date
type Date struct {
	time.Time
}

func (d *Date) String() string {
	return d.String()
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(RFC3339FullDate, s)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	s := d.Format(RFC3339FullDate)
	return json.Marshal(s)
}

// Validate validate items for save
func (i *TodoItem) Validate() error {
	return validator.New().Struct(i)
}
