package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

type StreamPattern struct {
	MatchingType string `json:"matching_type"`
	Description  string `json:"description"`
	Title        string `json:"title"`
	IndexSetId   string `json:"index_set_id"`
}

func (connector *GraylogConnector) GetAllStreams() ([]Entity, error) {
	streams, err := connector.GetData("streams", "streams", nil)
	if err != nil {
		return nil, err
	}
	return streams, nil
}

func CreateStreamBody(description string, title string, indexSets []Entity, indexSetName string) (string, error) {
	indexSetId := GetIdByTitle(indexSets, indexSetName)

	pattern := StreamPattern{MatchingType: "AND", Description: description, Title: title, IndexSetId: indexSetId}
	config, err := json.Marshal(pattern)
	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (connector *GraylogConnector) UpdateStream(streamId string, indexSets []Entity, indexSetName string, description string, title string) error {
	data, err := CreateStreamBody(description, title, indexSets, indexSetName)
	if err != nil {
		return err
	}
	_, statusCode, err := connector.PUT("streams/"+streamId, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't update stream " + title)
	}

	return nil
}

func (connector *GraylogConnector) UpdateOrCreateStream(streams []Entity, indexSets []Entity, indexSetName string, description string, title string) error {
	streamId := GetIdByTitle(streams, title)

	if streamId != "" {
		if err := connector.UpdateStream(
			streamId,
			indexSets,
			indexSetName,
			description,
			title); err != nil {
			return err
		}
	} else {
		if err := connector.OnlyCreateStream(streams, indexSets, indexSetName, description, title); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) UpdateStreams(streams []Entity, indexSets []Entity) error {

	streamToIndex := connector.GetStreams()

	for s, i := range streamToIndex {
		if err := connector.UpdateOrCreateStream(streams, indexSets, i, util.GraylogStreamsDescriptions[s], s); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateStream(indexSets []Entity, indexSetName string, description string, title string) error {
	data, err := CreateStreamBody(description, title, indexSets, indexSetName)
	if err != nil {
		return err
	}

	_, statusCode, err := connector.POST("streams/", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't create stream " + title)
	}
	return nil
}

func (connector *GraylogConnector) OnlyCreateStream(streams []Entity, indexSets []Entity, indexSetName string, description string, title string) error {
	streamId := GetIdByTitle(streams, title)

	if streamId == "" {
		if err := connector.CreateStream(
			indexSets,
			indexSetName,
			description,
			title); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateStreams(streams []Entity, indexSets []Entity) error {

	streamToIndex := connector.GetStreams()

	for s, i := range streamToIndex {
		if err := connector.OnlyCreateStream(streams, indexSets, i, util.GraylogStreamsDescriptions[s], s); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ResumeStream(id string) error {
	_, statusCode, err := connector.POST("streams/"+id+"/resume", "")
	if err != nil {
		return err
	}
	if statusCode != http.StatusNoContent {
		return errors.New("can't resume stream " + id)
	}
	return nil
}

func (connector *GraylogConnector) ResumeStreamByTitle(streams []Entity, title string) error {
	streamId := GetIdByTitle(streams, title)
	if err := connector.ResumeStream(streamId); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) ResumeStreams(streams []Entity) error {
	titles := connector.GetStreams()
	for s := range titles {
		if err := connector.ResumeStreamByTitle(streams, s); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateStreams(streams []Entity, indexSets []Entity, cr *loggingService.LoggingService) error {
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err := connector.CreateStreams(streams, indexSets); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdateStreams(streams, indexSets); err != nil {
			return err
		}
	}
	return nil
}

func (connector *GraylogConnector) ManageStreams(cr *loggingService.LoggingService) error {
	streams, err := connector.GetAllStreams()
	if err != nil {
		return err
	}

	indexSets, err := connector.GetAllIndexSets()
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateStreams(streams, indexSets, cr); err != nil {
		return err
	}

	streams, err = connector.GetAllStreams()
	if err != nil {
		return err
	}

	if err = connector.ResumeStreams(streams); err != nil {
		return err
	}
	return nil
}

func (connector *GraylogConnector) GetStreams() map[string]string {
	streamToIndex := make(map[string]string)

	for i := range connector.EnabledStreams {
		stream := connector.EnabledStreams[i].Title
		for s, i := range util.GraylogStreamsIndexTitles {
			if s == stream {
				streamToIndex[s] = i
				break
			}
		}
	}
	return streamToIndex
}
func GetDefaultStreams() []loggingService.Stream {
	return []loggingService.Stream{
		{
			Name:               util.GraylogAuditStream,
			Install:            true,
			RotationStrategy:   "timeBased",
			RotationPeriod:     "P1M",
			MaxNumberOfIndices: 5,
		},
		{
			Name:               util.GraylogSystemStream,
			Install:            true,
			RotationStrategy:   "sizeBased",
			MaxSize:            1073741824,
			MaxNumberOfIndices: 20,
		},
		{
			Name:               util.GraylogKubernetesEventsStream,
			Install:            true,
			RotationStrategy:   "timeBased",
			RotationPeriod:     "P1M",
			MaxNumberOfIndices: 5,
		},
	}
}
