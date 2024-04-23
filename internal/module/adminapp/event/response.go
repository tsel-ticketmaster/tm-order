package event

import "time"

type PromotorResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type LocationResponse struct {
	Country          string  `json:"country"`
	City             string  `json:"city"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
}

type ShowResponse struct {
	ID       string            `json:"id"`
	Venue    string            `json:"venue"`
	Type     string            `json:"type"`
	Location *LocationResponse `json:"location"`
	Time     time.Time         `json:"time"`
	Status   string            `json:"status"`
}

type CreateEventResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Promotors   []PromotorResponse
	Artists     []string `json:"artists"`
	Shows       []ShowResponse
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r *CreateEventResponse) PopulateFromEntity(e Event) {
	r.ID = e.ID
	r.Name = e.Name
	r.Description = e.Description
	r.Status = e.Status

	for _, v := range e.Promotors {
		r.Promotors = append(r.Promotors, PromotorResponse{
			Name:  v.Name,
			Email: v.Email,
			Phone: v.Phone,
		})
	}

	for _, v := range e.Artists {
		r.Artists = append(r.Artists, v.Name)
	}

	for _, v := range e.Shows {
		var location *LocationResponse
		if v.Location != nil {
			location = &LocationResponse{
				Country:          v.Location.Country,
				City:             v.Location.City,
				FormattedAddress: v.Location.FormattedAddress,
				Latitude:         v.Location.Latitude,
				Longitude:        v.Location.Longitude,
			}
		}
		r.Shows = append(r.Shows, ShowResponse{
			ID:       v.ID,
			Venue:    v.Venue,
			Type:     v.Type,
			Time:     v.Time,
			Status:   v.Status,
			Location: location,
		})
	}

	r.CreatedAt = e.CreatedAt
	r.UpdatedAt = e.UpdatedAt
}
