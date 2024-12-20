package utils

import "encoding/json"

type Entity struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

func GetIdByTitle(values []Entity, title string) string {
	for _, v := range values {
		if v.Title == title {
			return v.Id
		}
	}
	return ""
}

func (connector *GraylogConnector) GetRawData(url string) (string, error) {
	resp, _, err := connector.GET(url)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (connector *GraylogConnector) GetData(url string, field string, replaceFunc func(response string) string) ([]Entity, error) {
	allData, err := connector.GetRawData(url)
	if err != nil {
		return nil, err
	}

	if replaceFunc != nil {
		allData = replaceFunc(allData)
	}

	if field != "" {
		var data map[string]json.RawMessage
		if err = json.Unmarshal([]byte(allData), &data); err != nil {
			return nil, err
		}

		var parsedData []Entity
		if err = json.Unmarshal(data[field], &parsedData); err != nil {
			return nil, err
		}
		return parsedData, nil
	} else {
		var parsedData []Entity
		if err = json.Unmarshal([]byte(allData), &parsedData); err != nil {
			return nil, err
		}
		return parsedData, nil
	}
}
