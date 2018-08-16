package import_

import (
	"runtime"
	"path"
	"io/ioutil"
	"encoding/json"
	"github.com/aghape/geonames/models"
	"github.com/moisespsena-go/aorm"
	"fmt"
	"hash/fnv"
	"strings"
)

type estado struct {
	IdEstados, IdPaisesEstados, NomeEstados string
}

type pais struct {
	IdTodosPaises, ISOTodosPaises, NomeTodosPaises string
}

type rpais struct {
	IdPaises, IdTodosPaises string
}

type jsondata struct {
	Estados []estado
	Paises  []rpais
	TodosPaises   []pais
	PaisesModels []*models.GeoNamesCountry
	PaisesModelsByID map[string]*models.GeoNamesCountry
	StatesModels []*models.GeoNamesState
	StatesByID map[string][2]string
}

type states_country struct {
	Country string
	States []string
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func (j *jsondata) Parse(data []byte, statesData []byte) (err error) {
	err = json.Unmarshal(data, j)
	if err != nil {
		return err
	}
	var paises []*models.GeoNamesCountry
	paisesById := make(map[string]*models.GeoNamesCountry)
	for _, p := range j.TodosPaises {
		pais := &models.GeoNamesCountry{p.ISOTodosPaises, p.NomeTodosPaises}
		paises = append(paises, pais)
		paisesById[p.IdTodosPaises] = pais
	}
	j.PaisesModels, j.PaisesModelsByID = paises, paisesById

	j.StatesByID = make(map[string][2]string)
	for _, s := range j.Estados {
		j.StatesByID[s.IdEstados] = [2]string{paisesById[s.IdPaisesEstados].ID + hash(strings.TrimSpace(s.NomeEstados)), s.NomeEstados}
	}

	//var states []*models.GeoNamesState

	var countries []states_country
	err = json.Unmarshal(statesData, &countries)
	if err != nil {
		return err
	}

	statesById := make(map[string]*models.GeoNamesState)

	for _, country := range countries {
		for _, stateName := range country.States {
			state := &models.GeoNamesState{ID:country.Country+hash(stateName), Name:stateName, CountryID:country.Country}
			j.StatesModels = append(j.StatesModels, state)
			if _, ok := statesById[state.ID]; ok {
				return fmt.Errorf("Duplication state error: %v", state.ID)
			}
			statesById[state.ID] = state
		}
	}

	for _, id := range j.StatesByID {
		if _, ok := statesById[id[0]]; !ok {
			fmt.Println(fmt.Errorf("State %v not found", id))
		}
	}

	return nil
}
func (j *jsondata) Save(dir string) error {
	data, err := json.Marshal(j.PaisesModels)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(dir, "countries.json"), data, 0664)
	if err != nil {
		return err
	}

	data, err = json.Marshal(j.StatesModels)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(dir, "states.json"), data, 0664)
	if err != nil {
		return err
	}
	return nil
}
func Parse() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	dir := path.Join(path.Dir(filename), "data")
	dataFile := path.Join(dir, "jsam.json")
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		panic(err)
	}
	statesData, err := ioutil.ReadFile(path.Join(dir, "states_list.json"))
	if err != nil {
		panic(err)
	}
	j := &jsondata{}
	err = j.Parse(data, statesData)
	if err != nil {
		panic(err)
	}
	err = j.Save(dir)
	if err != nil {
		panic(err)
	}
}

func Import(db *aorm.DB) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	dir := path.Join(path.Dir(filename), "data")
	filename = path.Join(dir, "countries.json")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var paises []*models.GeoNamesCountry
	err = json.Unmarshal(data, &paises)
	if err != nil {
		panic(err)
	}

	for i, p := range paises {
		err = db.FirstOrCreate(p).Error
		if err != nil {
			panic(fmt.Errorf("Failed to create Coutry[%v] = %v on DB: %v", i, p, err))
		}
	}
	filename = path.Join(dir, "states.json")
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var states []*models.GeoNamesState
	err = json.Unmarshal(data, &states)
	if err != nil {
		panic(err)
	}

	for i, s := range states {
		err = db.FirstOrCreate(s).Error
		if err != nil {
			panic(fmt.Errorf("Failed to create State[%v]= %v on DB: %v", i, s, err))
		}
	}
}