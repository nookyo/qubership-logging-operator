package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

type GrokPattern struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type DTOGrokPattern struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

func ContainsPattern(pattern GrokPattern, list []GrokPattern) bool {
	for _, p := range list {
		if p.Name == pattern.Name {
			return true
		}
	}
	return false
}

func GetPattern(name string, patterns []GrokPattern) string {
	for _, p := range patterns {
		if p.Name == name {
			return p.Pattern
		}
	}
	return ""
}

func (connector *GraylogConnector) GetAllGrokPatterns() (string, error) {
	resp, _, err := connector.GET(grokUrl)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func CreateGrokPatternBody(p GrokPattern, allGrokPatterns []GrokPattern) (string, error) {
	pattern := DTOGrokPattern{Name: p.Name, Pattern: GetPattern(p.Name, allGrokPatterns)}
	config, err := json.Marshal(pattern)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (connector *GraylogConnector) UpdateGrokPatterns(existingGrokPatterns []GrokPattern, allGrokPatterns []GrokPattern) error {
	var failedPatterns []string
	for _, v := range existingGrokPatterns {
		if ContainsPattern(v, allGrokPatterns) {
			uri := grokUrl + "/" + v.Id
			config, err := CreateGrokPatternBody(v, allGrokPatterns)
			if err != nil {
				connector.Log.V(util.Warn).Info("can't marshal grok pattern: " + v.Pattern)
				failedPatterns = append(failedPatterns, v.Pattern)
			}
			_, statusCode, err := connector.PUT(uri, config)
			if err != nil {
				connector.Log.V(util.Warn).Info("can't update grok pattern: " + v.Name)
				failedPatterns = append(failedPatterns, v.Pattern)
			}
			if statusCode != http.StatusOK {
				return errors.New("can't update grok pattern: " + v.Name)
			}
		}
	}
	if len(failedPatterns) != 0 {
		return errors.New("Can't update patterns " + strings.Join(failedPatterns, ","))
	}
	return nil
}

func (connector *GraylogConnector) CreateGrokPatterns(existingGrokPatterns []GrokPattern, allGrokPatterns []GrokPattern) error {
	var failedPatterns []string
	for _, v := range allGrokPatterns {
		connector.Log.V(util.Debug).Info("Processing grok pattern " + v.Name)
		if !ContainsPattern(v, existingGrokPatterns) {
			connector.Log.V(util.Debug).Info("Pattern not found. Create it")
			config, err := CreateGrokPatternBody(v, allGrokPatterns)
			if err != nil {
				connector.Log.V(util.Warn).Info("can't marshal grok pattern: " + v.Pattern)
				failedPatterns = append(failedPatterns, v.Pattern)
			}
			_, statusCode, err := connector.POST(grokUrl, config)
			if err != nil {
				connector.Log.V(util.Warn).Info("can't create grok pattern: " + v.Name)
				failedPatterns = append(failedPatterns, v.Pattern)
			}
			if statusCode != http.StatusCreated {
				return errors.New("can't create grok pattern: " + v.Name)
			}
		}
	}
	if len(failedPatterns) != 0 {
		return errors.New("Can't create patterns " + strings.Join(failedPatterns, ","))
	}
	return nil
}

func (connector *GraylogConnector) CreateOrUpdateGrokPatterns(existingGrokPatterns []GrokPattern,
	cr *loggingService.LoggingService) error {
	rawAllGrokPatterns, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogGrokPatterns), util.GraylogGrokPatterns, cr.ToParams())
	if err != nil {
		return err
	}

	var data map[string]json.RawMessage
	if err = json.Unmarshal([]byte(rawAllGrokPatterns), &data); err != nil {
		return err
	}

	var allGrokPatterns []GrokPattern
	if err = json.Unmarshal(data["patterns"], &allGrokPatterns); err != nil {
		return err
	}
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err = connector.CreateGrokPatterns(existingGrokPatterns, allGrokPatterns); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err = connector.UpdateGrokPatterns(existingGrokPatterns, allGrokPatterns); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManageGrokPatterns(cr *loggingService.LoggingService) error {
	existingGrokPatterns, err := connector.GetAllGrokPatterns()
	if err != nil {
		return err
	}

	var data map[string]json.RawMessage
	if err = json.Unmarshal([]byte(existingGrokPatterns), &data); err != nil {
		return err
	}

	var grokPatterns []GrokPattern

	if err = json.Unmarshal(data["patterns"], &grokPatterns); err != nil {
		return err
	}

	if err = connector.CreateOrUpdateGrokPatterns(grokPatterns, cr); err != nil {
		return err
	}

	return nil
}
