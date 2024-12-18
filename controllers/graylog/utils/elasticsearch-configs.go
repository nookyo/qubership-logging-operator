package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

const (
	elasticsearchConfigsDir    string = "elasticsearch-configs"
	openSearchConfigsDir       string = "opensearch-configs"
	snapshotArchivesRequestUrl        = "_snapshot/archives"
)

type Request struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Body   interface{} `json:"body,omitempty"`
}

func (connector *GraylogConnector) SendRequestToOpenSearch(method string, urlPath string, data []byte, cr *loggingService.LoggingService) error {
	requestUrl, err := url.JoinPath(connector.OpenSearchRestClient.Host, urlPath)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(data)

	connector.Log.V(util.Debug).Info("Send " + method + " request to: " + requestUrl + " with body: " + string(data))

	request, err := http.NewRequest(method, requestUrl, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	connector.OpenSearchRestClient.SetAuthHeader(request)
	var response *http.Response
	response, err = connector.OpenSearchRestClient.Client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	b, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Fatalln(err)
	}
	connector.Log.V(util.Debug).Info(string(b))

	return nil
}

func (connector *GraylogConnector) UploadConfig(name string, cr *loggingService.LoggingService, configsDir string) error {
	fileContent, err := util.ReadFile(filepath.Join(dataDir, configsDir, name))
	if err != nil {
		return err
	}
	rawConfigs, err := util.ParseTemplate(fileContent, filepath.Join(dataDir, configsDir, name), cr.ToParams())
	if err != nil {
		return err
	}

	var data map[string]json.RawMessage
	if err = json.Unmarshal([]byte(rawConfigs), &data); err != nil {
		return err
	}

	var allRequests []Request
	if err = json.Unmarshal(data["requests"], &allRequests); err != nil {
		return err
	}

	for _, r := range allRequests {
		body, err := json.Marshal(r.Body)
		if err != nil {
			return err
		}
		if err = connector.SendRequestToOpenSearch(r.Method, r.URL, body, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) ManageOpensearchConfigs(cr *loggingService.LoggingService) (err error) {
	configsDirs := []string{openSearchConfigsDir, elasticsearchConfigsDir}
	for _, configsDir := range configsDirs {
		if _, err = os.Stat(filepath.Join(dataDir, configsDir)); err == nil {
			var files []os.DirEntry
			files, err = os.ReadDir(filepath.Join(dataDir, configsDir))
			if err != nil {
				return err
			}
			for _, f := range files {
				if err = connector.UploadConfig(f.Name(), cr, configsDir); err != nil {
					return err
				}
			}
		} else if os.IsNotExist(err) {
			connector.Log.V(util.Debug).Info("Directory " + filepath.Join(dataDir, configsDir) + " not found")
		} else {
			return err
		}
	}
	return nil
}

func (connector *GraylogConnector) ManageArchivesDirectory(cr *loggingService.LoggingService) error {
	elasticsearchHost := os.Getenv("ELASTICSEARCH_HOST")
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogArchivesDirectory), util.GraylogArchivesDirectory, cr.ToParams())
	if err != nil {
		return err
	}
	if strings.Contains(elasticsearchHost, "opensearch") {
		data = strings.Replace(data, "elasticsearch", "opensearch", 1)
	}
	if err = connector.SendRequestToOpenSearch(http.MethodPut, snapshotArchivesRequestUrl, []byte(data), cr); err != nil {
		return err
	}
	return nil
}
