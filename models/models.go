package models

type GeoNamesCountry struct {
	ID   string `gorm:"size:2;primary_key"`
	Name string `gorm:"size:255"`
}

func (c *GeoNamesCountry) Flag() string {
	return "http://www.geonames.org/flags/x/" + c.ID + ".gif"
}

type GeoNamesState struct {
	ID        string           `gorm:"size:255;primary_key"`
	Name      string           `gorm:"size:255"`
	CountryID string           `gorm:"size:2"`
	Country   *GeoNamesCountry `json:"-"`
}
