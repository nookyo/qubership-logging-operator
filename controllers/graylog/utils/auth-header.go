package utils

import (
	"errors"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

func (connector *GraylogConnector) EnableAuthHeader(cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogAuthHeader), util.GraylogAuthHeader, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT(authHeaderUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't enable HTTP header auth")
	}
	return nil
}

func (connector *GraylogConnector) ManageAuthHeaderConfig(cr *loggingService.LoggingService) error {
	if err := connector.EnableAuthHeader(cr); err != nil {
		return err
	}
	return nil
}
