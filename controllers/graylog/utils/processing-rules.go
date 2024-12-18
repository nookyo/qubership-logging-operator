package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

type RulePattern struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

func (connector *GraylogConnector) GetAllProcessingRules() ([]Entity, error) {
	rules, err := connector.GetData(processingRulesUrl, "", nil)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func CreateRuleBody(title string, description string, source string) (string, error) {
	pattern := RulePattern{Title: title, Description: description, Source: source}
	config, err := json.Marshal(pattern)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (connector *GraylogConnector) CreateRuleData(streams []Entity, title string, template string) (string, error) {
	streamId := GetIdByTitle(streams, title)
	if streamId == "" {
		if title == util.GraylogDefaultStream {
			streamId = GetIdByTitle(streams, util.GraylogAllMessagesStream)
			if streamId == "" {
				return "", errors.New("neither " + util.GraylogDefaultStream + " nor " + util.GraylogAllMessagesStream + " streams were not found")
			}
		}
	}

	settings := map[string]string{
		"streamId": streamId,
	}

	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, settings)
	if err != nil {
		return "", err
	}

	return data, nil
}

func (connector *GraylogConnector) UpdateRule(ruleId string, title string, description string, source string) error {
	data, err := CreateRuleBody(title, description, source)
	if err != nil {
		return err
	}

	_, statusCode, err := connector.PUT(processingRulesUrl+"/"+ruleId, data)
	if err != nil {
		return err
	}
	if (statusCode != http.StatusOK) && (statusCode != http.StatusCreated) {
		return errors.New("Can't update rule: " + title)
	}

	return nil
}

func (connector *GraylogConnector) CreateRule(title string, description string, source string) error {
	data, err := CreateRuleBody(title, description, source)
	if err != nil {
		return err
	}

	_, statusCode, err := connector.POST(processingRulesUrl, data)
	if err != nil {
		return err
	}
	if (statusCode != http.StatusOK) && (statusCode != http.StatusCreated) {
		return errors.New("Can't create rule: " + title)
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateRules(streams []Entity, processingRules []Entity, ruleTitle string, streamTitle string, description string, pathConfig string) error {
	ruleId := GetIdByTitle(processingRules, ruleTitle)

	if ruleId != "" {

		data, err := connector.CreateRuleData(streams, streamTitle, pathConfig)
		if err != nil {
			return err
		}

		if err = connector.UpdateRule(
			ruleId,
			ruleTitle,
			description,
			data); err != nil {
			return err
		}
	} else {
		if err := connector.OnlyCreateRules(streams, processingRules, ruleTitle, streamTitle, description, pathConfig); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) UpdateProcessingRules(processingRules []Entity) error {
	streams, err := connector.GetAllStreams()
	if err != nil {
		return err
	}

	ruleToStream := connector.GetProcessingRules()

	for r, s := range ruleToStream {
		if err = connector.CreateOrUpdateRules(streams, processingRules, r, s, util.GraylogRuleDescriptions[r], util.GraylogRuleConfigs[r]); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) OnlyCreateRules(streams []Entity, processingRules []Entity, ruleTitle string, streamTitle string, description string, pathConfig string) error {

	ruleId := GetIdByTitle(processingRules, ruleTitle)

	if ruleId == "" {
		data, err := connector.CreateRuleData(streams, streamTitle, pathConfig)
		if err != nil {
			return err
		}

		if err = connector.CreateRule(
			ruleTitle,
			description,
			data); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateProcessingRules(processingRules []Entity) error {
	streams, err := connector.GetAllStreams()
	if err != nil {
		return err
	}

	rulesToStream := connector.GetProcessingRules()

	for r, s := range rulesToStream {
		if err = connector.OnlyCreateRules(streams, processingRules, r, s, util.GraylogRuleDescriptions[r], util.GraylogRuleConfigs[r]); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateProcessingRules(processingRules []Entity, cr *loggingService.LoggingService) error {
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err := connector.CreateProcessingRules(processingRules); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdateProcessingRules(processingRules); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManageProcessingRules(cr *loggingService.LoggingService) error {
	processingRules, err := connector.GetAllProcessingRules()
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateProcessingRules(processingRules, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) GetProcessingRules() map[string]string {
	ruleToStream := map[string]string{
		util.GraylogUnsupportedSymbolsRule:     util.GraylogUnsupportedSymbolsRuleStream,
		util.GraylogRemoveKubernetesRule:       util.GraylogRemoveKubernetesRuleStream,
		util.GraylogRemoveKubernetesLabelsRule: util.GraylogRemoveKubernetesLabelsRuleStream,
	}
	for i := range connector.EnabledStreams {
		stream := connector.EnabledStreams[i].Title
		for k, v := range util.GraylogStreamsRules {
			if k == stream {
				ruleToStream[v] = k
				break
			}
		}
	}
	return ruleToStream
}
