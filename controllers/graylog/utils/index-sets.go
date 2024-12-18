package utils

import (
	"errors"
	"fmt"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

func (connector *GraylogConnector) GetAllIndexSets() ([]Entity, error) {
	indexSets, err := connector.GetData(indexSetsUrl, "index_sets", nil)
	if err != nil {
		return nil, err
	}
	return indexSets, nil
}

func (connector *GraylogConnector) UpdateIndexSet(id string, cr *loggingService.LoggingService, path string, indexSetName string) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, path), path, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT(indexSetsUrl+"/"+id, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't update " + indexSetName)
	}
	return nil
}

func (connector *GraylogConnector) CreateIndexSet(cr *loggingService.LoggingService, path string, indexSetName string) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, path), path, cr.ToParams())
	if err != nil {
		return err
	}
	response, statusCode, err := connector.POST(indexSetsUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("can't create %s. Status code: %v. Response: %s", indexSetName, statusCode, response)
	}
	return nil
}

func (connector *GraylogConnector) CreateOrUpdateIndexSet(indexSets []Entity, cr *loggingService.LoggingService, path string, title string) error {
	id := GetIdByTitle(indexSets, title)
	if id == "" {
		if title == util.GraylogDefaultIndexSet {
			return errors.New("default index set not found")
		}
		if err := connector.CreateIndexSet(cr, path, title); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() || title == util.GraylogDefaultIndexSet {
		if err := connector.UpdateIndexSet(id, cr, path, title); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManageIndexSets(cr *loggingService.LoggingService) error {
	indexSets, err := connector.GetAllIndexSets()
	if err != nil {
		return err
	}

	availableIndexSets := connector.GetIndexSets()

	for k, v := range availableIndexSets {
		if err = connector.CreateOrUpdateIndexSet(indexSets, cr, v, k); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) GetIndexSets() map[string]string {
	availableIndexSets := map[string]string{
		util.GraylogDefaultIndexSet: util.GraylogIndexConfigs[util.GraylogDefaultIndexSet],
		util.GraylogAuditIndexSet:   util.GraylogIndexConfigs[util.GraylogAuditIndexSet],
	}

	for i := range connector.EnabledStreams {
		stream := connector.EnabledStreams[i].Title
		for k, v := range util.GraylogStreamsIndexTitles {
			if k == stream {
				availableIndexSets[v] = util.GraylogIndexConfigs[v]
				break
			}
		}
	}
	return availableIndexSets
}
