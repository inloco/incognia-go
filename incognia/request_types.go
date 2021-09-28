package incognia

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type StructuredAddress struct {
	Locale       string `json:"locale"`
	CountryName  string `json:"country_name"`
	CountryCode  string `json:"country_code"`
	State        string `json:"state"`
	City         string `json:"city"`
	Borough      string `json:"borough"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complements  string `json:"complements"`
	PostalCode   string `json:"postal_code"`
}

type Address struct {
	Coordinates       *Coordinates
	StructuredAddress *StructuredAddress
	AddressLine       string
}

type postAssessmentRequestBody struct {
	InstallationId    string             `json:"installation_id"`
	AddressLine       string             `json:"address_line,omitempty"`
	StructuredAddress *StructuredAddress `json:"structured_address,omitempty"`
	Coordinates       *Coordinates       `json:"address_coordinates,omitempty"`
}
