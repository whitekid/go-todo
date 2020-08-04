package models

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	RFC3339FullDate = "2006-01-02" // RFC3339FullDate ...
)

// Item todo item
type Item struct {
	ID      string `json:"id" format:"uuid" example:"628b92ab-6d95-4fbe-b7c6-09cf5cd8941c"`
	Title   string `json:"title" example:"do something in future"`
	DueDate Date   `json:"due_date" swaggertype:"string" example:"2006-01-02"`
	Rank    int    `json:"rank" format:"int" example:"1"` // rank order
}

// Date represents json full-date format, supports marshal json date
type Date struct {
	time.Time
}

func Today() (d Date) {
	d.Time = time.Now().UTC().Truncate(time.Hour * 24)
	return
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
func (i *Item) Validate() error {
	if i.Title == "" {
		return errors.New("title required")
	}

	return nil
}
