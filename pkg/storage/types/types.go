package types

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("Not found")
)

const (
	RFC3339FullDate = "2006-01-02" // RFC3339FullDate ...
)

// Interface represent storage abstract layer
type Interface interface {
	TodoService() TodoStorage
}

// TodoStorage represents todo storage
type TodoStorage interface {
	Create(*TodoItem) error

	List() ([]TodoItem, error) // list todo items

	// return ErrNotFound if item not found
	Get(itemID string) (*TodoItem, error)

	// return ErrNotFound if item not found
	Update(item *TodoItem) error

	Delete(itemID string) error
}

// TodoItem todo item
type TodoItem struct {
	ID      string `json:"id" format:"uuid" example:"628b92ab-6d95-4fbe-b7c6-09cf5cd8941c"`
	Title   string `json:"title" example:"do something in future"`
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
	if i.Title == "" {
		return errors.New("title required")
	}

	return nil
}
