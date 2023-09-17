package service

import "user-service/api_clients/client"

type EnrichedFIO struct {
	Name        string
	Surname     string
	Patronymic  string
	Age         int
	Gender      string
	Nationality []client.CountryProbability
}

func (f *FIOService) enrichFIOData(fioMessage FIO) (EnrichedFIO, error) {
	enriched := EnrichedFIO{
		Name:       fioMessage.Name,
		Surname:    fioMessage.Surname,
		Patronymic: fioMessage.Patronymic,
	}

	// Обогащение возраста
	ageData, err := client.GetAgeByName(fioMessage.Name)
	if err != nil {
		return EnrichedFIO{}, err
	}
	enriched.Age = ageData.Age

	// Обогащение пола
	genderData, err := client.GetGenderByName(fioMessage.Name)
	if err != nil {
		return EnrichedFIO{}, err
	}
	enriched.Gender = genderData.Gender

	// Обогащение национальности
	nationalityData, err := client.GetNationalityByName(fioMessage.Name)
	if err != nil {
		return EnrichedFIO{}, err
	}
	enriched.Nationality = nationalityData.Country

	return enriched, nil
}
