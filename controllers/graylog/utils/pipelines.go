package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

type PipelinePattern struct {
	PipelineIds []string `json:"pipeline_ids"`
	StreamId    string   `json:"stream_id"`
}

func (connector *GraylogConnector) GetAllPipelines() ([]Entity, error) {
	streams, err := connector.GetData(pipelineUrl, "", nil)
	if err != nil {
		return nil, err
	}
	return streams, nil
}

func CreatePipelineBody(pipelineId string, streamId string) (string, error) {
	pattern := PipelinePattern{PipelineIds: []string{pipelineId}, StreamId: streamId}
	config, err := json.Marshal(pattern)

	if err != nil {
		return "", err
	}
	return string(config), nil
}

func (connector *GraylogConnector) UpdatePipeline(pipelines []Entity, cr *loggingService.LoggingService) error {
	logsRoutingPipelineId := GetIdByTitle(pipelines, "Logs routing")

	if logsRoutingPipelineId != "" {
		data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogPipeline), util.GraylogPipeline, cr.ToParams())
		if err != nil {
			return err
		}

		_, statusCode, err := connector.PUT(pipelineUrl+"/"+logsRoutingPipelineId, data)
		if err != nil {
			return err
		}
		if statusCode != http.StatusOK {
			return errors.New("can't update pipeline " + logsRoutingPipelineId)
		}
	} else {
		if err := connector.CreatePipeline(pipelines, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreatePipeline(pipelines []Entity, cr *loggingService.LoggingService) error {
	logsRoutingPipelineId := GetIdByTitle(pipelines, "Logs routing")

	if logsRoutingPipelineId == "" {
		data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogPipeline), util.GraylogPipeline, cr.ToParams())
		if err != nil {
			return err
		}

		_, statusCode, err := connector.POST(pipelineUrl, data)
		if err != nil {
			return err
		}
		if statusCode != http.StatusOK {
			return errors.New("can't create pipeline")
		}
	}

	return nil
}

func (connector *GraylogConnector) ConnectPipeline(pipelines []Entity) error {
	logsRoutingPipelineId := GetIdByTitle(pipelines, "Logs routing")

	streams, err := connector.GetAllStreams()
	if err != nil {
		return err
	}

	streamId := GetIdByTitle(streams, util.GraylogDefaultStream)
	if streamId == "" {
		streamId = GetIdByTitle(streams, util.GraylogAllMessagesStream)
		if streamId == "" {
			return errors.New("neither " + util.GraylogDefaultStream + " nor " + util.GraylogAllMessagesStream + " streams were not found")
		}
	}

	data, err := CreatePipelineBody(logsRoutingPipelineId, streamId)
	if err != nil {
		return err
	}

	_, statusCode, err := connector.POST("system/pipelines/connections/to_stream", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't connect pipeline")
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdatePipelines(pipelines []Entity, cr *loggingService.LoggingService) error {
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err := connector.CreatePipeline(pipelines, cr); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdatePipeline(pipelines, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManagePipelines(cr *loggingService.LoggingService) error {
	pipelines, err := connector.GetAllPipelines()
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdatePipelines(pipelines, cr); err != nil {
		return err
	}

	pipelines, err = connector.GetAllPipelines()
	if err != nil {
		return err
	}

	if err = connector.ConnectPipeline(pipelines); err != nil {
		return err
	}

	return nil
}
