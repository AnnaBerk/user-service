package client

import (
	"encoding/json"
	"net/http"
)

const baseAgeURL = "https://api.agify.io"
const baseGenderURL = "https://api.genderize.io"
const baseNationURL = "https://api.nationalize.io"

type AgeResponse struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

type GenderResponse struct {
	Count       int     `json:"count"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
}
type CountryProbability struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

type NationResponse struct {
	Count   int                  `json:"count"`
	Name    string               `json:"name"`
	Country []CountryProbability `json:"country"`
}

func GetAgeByName(name string) (AgeResponse, error) {
	resp, err := http.Get(baseAgeURL + "?name=" + name)
	if err != nil {
		return AgeResponse{}, err
	}
	defer resp.Body.Close()

	var r AgeResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}

func GetGenderByName(name string) (GenderResponse, error) {
	resp, err := http.Get(baseGenderURL + "?name=" + name)
	if err != nil {
		return GenderResponse{}, err
	}
	defer resp.Body.Close()

	var r GenderResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}

func GetNationalityByName(name string) (NationResponse, error) {
	resp, err := http.Get(baseNationURL + "?name=" + name)
	if err != nil {
		return NationResponse{}, err
	}
	defer resp.Body.Close()

	var r NationResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}
