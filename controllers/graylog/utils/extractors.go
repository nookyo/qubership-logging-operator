package utils

import (
	"errors"
	"net/http"
	"strconv"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

func (connector *GraylogConnector) GetAllExtractors(id string) ([]Entity, error) {
	extractors, err := connector.GetData("system/inputs/"+id+"/extractors", "extractors", nil)
	if err != nil {
		return nil, err
	}
	return extractors, nil
}

func (connector *GraylogConnector) UpdateMessageProcessorsOrder(cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogMessageProcessors), util.GraylogMessageProcessors, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT("system/messageprocessors/config", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't update messageprocessors order")
	}
	return nil
}

func (connector *GraylogConnector) CreateExtractor(template string, inputId string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.POST("system/inputs/"+inputId+"/extractors/", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't create extractor " + template)
	}
	return nil
}

func (connector *GraylogConnector) UpdateExtractor(template string, inputId string, id string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT("system/inputs/"+inputId+"/extractors/"+id, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't update extractor " + template)
	}
	return nil
}

func (connector *GraylogConnector) DeleteExtractor(inputId string, id string) error {
	_, statusCode, err := connector.DELETE("system/inputs/" + inputId + "/extractors/" + id)
	if err != nil {
		return err
	}
	if statusCode != http.StatusNoContent {
		return errors.New("can't delete extractor " + id)
	}
	return nil
}

func (connector *GraylogConnector) UpdateAllExtractors(extractors []Entity, extractorsAssets map[string]string, inputId string, cr *loggingService.LoggingService) error {
	for name, assetPath := range extractorsAssets {
		id := GetIdByTitle(extractors, name)
		if id != "" {
			if err := connector.UpdateExtractor(assetPath, inputId, id, cr); err != nil {
				return err
			}
		} else {
			if err := connector.CreateExtractor(assetPath, inputId, cr); err != nil {
				return err
			}
		}
	}

	id := GetIdByTitle(extractors, "os_extractor")
	if id != "" {
		if err := connector.DeleteExtractor(inputId, id); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateExtractors(extractors []Entity, extractorsAssets map[string]string, inputId string, cr *loggingService.LoggingService) error {
	for name, assetPath := range extractorsAssets {
		id := GetIdByTitle(extractors, name)
		if id == "" {
			if err := connector.CreateExtractor(assetPath, inputId, cr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateExtractors(extractors []Entity, extractorsAssets map[string]string, id string, cr *loggingService.LoggingService) error {
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err := connector.CreateExtractors(extractors, extractorsAssets, id, cr); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdateAllExtractors(extractors, extractorsAssets, id, cr); err != nil {
			return err
		}
	}
	return nil
}

func (connector *GraylogConnector) ManageExtractors(cr *loggingService.LoggingService, installGraylog5 bool) error {
	inputs, err := connector.GetAllInputs()
	if err != nil {
		return err
	}

	id := GetIdByTitle(inputs, "input-"+strconv.Itoa(cr.Spec.Graylog.InputPort))

	extractors, err := connector.GetAllExtractors(id)
	if err != nil {
		return err
	}

	// Extractors APIs for Graylog 4 and 5 are different
	var extractorsAssets map[string]string
	if installGraylog5 {
		extractorsAssets = util.Graylog5Extractors
	} else {
		extractorsAssets = util.Graylog4Extractors
	}
	if err = connector.CreateOrUpdateExtractors(extractors, extractorsAssets, id, cr); err != nil {
		return err
	}

	if err = connector.UpdateMessageProcessorsOrder(cr); err != nil {
		return err
	}

	return nil
}
