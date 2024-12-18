package utils

import (
	"errors"
	"net/http"
	"strconv"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

func (connector *GraylogConnector) GetAllInputs() ([]Entity, error) {
	inputs, err := connector.GetData(inputsUrl, "inputs", nil)
	if err != nil {
		return nil, err
	}
	return inputs, nil
}

func (connector *GraylogConnector) UpdateDefaultInput(id string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogInput), util.GraylogInput, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT(inputsUrl+"/"+id, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't update default input")
	}
	return nil
}

func (connector *GraylogConnector) CreateDefaultInput(cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogInput), util.GraylogInput, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.POST(inputsUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't create default input")
	}
	return nil
}

func (connector *GraylogConnector) CreateOrUpdateDefaultInput(inputs []Entity, cr *loggingService.LoggingService) error {
	id := GetIdByTitle(inputs, "input-"+strconv.Itoa(cr.Spec.Graylog.InputPort))
	if id == "" {
		if err := connector.CreateDefaultInput(cr); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdateDefaultInput(id, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManageInputs(cr *loggingService.LoggingService) error {
	inputs, err := connector.GetAllInputs()
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateDefaultInput(inputs, cr); err != nil {
		return err
	}
	return nil
}
