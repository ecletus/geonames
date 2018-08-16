package geonames

import (
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/aghape/helpers"
	"github.com/aghape/geonames/models"
	"fmt"
)

func MigrateDB(db *aorm.DB) {
	country := db.NewScope(&models.GeoNamesCountry{})
	state := db.NewScope(&models.GeoNamesState{})
	key, err := helpers.CheckReturnError(
		func() (key string, err error) {
			return "AutoMigrate", db.AutoMigrate(&models.GeoNamesCountry{}, &models.GeoNamesState{}).Error
		},
		func() (key string, err error) {
			return "CreateIndex", state.DB().AddIndex(state.TableName() + "_country_id", "country_id").Error
		},
		func() (key string, err error) {
			return "CreateFKs", state.DB().AddForeignKey("country_id", country.TableName() + "(id)", "RESTRICT", "RESTRICT").Error
		},
	)
	if err != nil {
		panic(fmt.Errorf("qor/geonames:setup.MigrateDB.%v: failed to migrate DB: %v", key, err))
	}
}