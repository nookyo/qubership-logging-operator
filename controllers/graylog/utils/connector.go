package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
)

const (
	GraylogPort                 = "9000"
	contentpacksUrl             = "system/content_packs"
	oobContentPackId            = "0fac53ed-df74-4ba6-88c2-aa16b4b8542d"
	oobContentPackInstallUrl    = "system/content_packs/" + oobContentPackId + "/1/installations"
	contentPackInstallationsUrl = "system/content_packs/%s/installations"
	contentPackInstallationUrl  = "system/content_packs/%s/installations/%s"
	grokUrl                     = "system/grok"
	indexSetsUrl                = "system/indices/index_sets"
	inputsUrl                   = "system/inputs"
	pipelineUrl                 = "system/pipelines/pipeline"
	processingRulesUrl          = "system/pipelines/rule"
	authHeaderUrl               = "system/authentication/http-header-auth-config"
)

type GraylogConnector struct {
	RestClient           *util.RestClient
	OpenSearchRestClient *util.RestClient
	Log                  logr.Logger
	Assets               embed.FS
	EnabledStreams       []Stream
	TLSEnabled           bool
}

type Streams struct {
	IntegrationLogsEnabled bool
	AccessLogsEnabled      bool
	BillCycleLogsEnabled   bool
}

type Stream struct {
	Title              string
	RotationStrategy   string
	RotationPeriod     string
	MaxSize            int
	MaxNumberOfIndices int
}

func CreateConnector(ctx context.Context, cr *loggingService.LoggingService, assets embed.FS, clientSet kubernetes.Interface) (*GraylogConnector, error) {
	var user *util.Сreds
	var tlsConfig *tls.Config
	var err error

	if cr.Spec.Graylog != nil && cr.Spec.Graylog.TLS != nil &&
		cr.Spec.Graylog.TLS.HTTP != nil && cr.Spec.Graylog.TLS.HTTP.Enabled {

		tlsConfig, err = cr.Spec.Graylog.TLS.HTTP.GetCertificates(ctx, clientSet, cr.GetNamespace())
		if err != nil {
			return nil, err
		}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns
	transport.MaxConnsPerHost = transport.MaxIdleConns
	transport.TLSClientConfig = tlsConfig
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(util.ConnectionTimeout) * time.Second,
	}

	if len(cr.Spec.Graylog.User) != 0 && len(cr.Spec.Graylog.Password) != 0 {
		user = &util.Сreds{
			Name:     cr.Spec.Graylog.User,
			Password: cr.Spec.Graylog.Password,
		}
	} else {
		name := os.Getenv("GRAYLOG_USERNAME")
		pwd := os.Getenv("GRAYLOG_PASSWORD")
		if len(name) != 0 && len(pwd) != 0 {
			user = &util.Сreds{
				Name:     name,
				Password: pwd,
			}
		}
	}

	restClient := &util.RestClient{
		Client: httpClient,
		Auth:   user,
		Host:   util.GraylogComponentName + "." + cr.GetNamespace() + ".svc" + ":" + GraylogPort + "/api/",
	}

	tlsConfig = nil

	var host string
	if cr.Spec.Graylog.OpenSearch != nil {
		if cr.Spec.Graylog.OpenSearch.HTTPConfig != nil {
			var name, pwd, token string
			name, pwd, token, tlsConfig, err = cr.Spec.Graylog.OpenSearch.HTTPConfig.GetCredentialsAndCertificates(ctx, clientSet, cr.GetNamespace())
			if err != nil {
				return nil, err
			}

			if (name != "" && pwd != "") || token != "" {
				user = &util.Сreds{
					Name:     name,
					Password: pwd,
					Token:    token,
				}
			}

			host = cr.Spec.Graylog.OpenSearch.Host
		}
	}

	if len(host) == 0 {
		var u *url.URL
		u, err = url.Parse(os.Getenv("ELASTICSEARCH_HOST"))
		if err != nil {
			return nil, err
		}

		if tlsConfig == nil && u.Scheme == "https" {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		hostUrl := &url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     u.Path,
			RawQuery: u.RawQuery,
			Fragment: u.Fragment,
		}
		host = hostUrl.String()
		password, _ := u.User.Password()
		user = &util.Сreds{
			Name:     u.User.Username(),
			Password: password,
		}
	}
	transport = http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns
	transport.MaxConnsPerHost = transport.MaxIdleConns
	transport.TLSClientConfig = tlsConfig
	httpClient = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(util.ConnectionTimeout) * time.Second,
	}

	openSearchClient := &util.RestClient{
		Client: httpClient,
		Auth:   user,
		Host:   host,
	}

	var enabledStreams = GetStreams(cr)
	return &GraylogConnector{
		Log:                  util.Logger("connector"),
		RestClient:           restClient,
		OpenSearchRestClient: openSearchClient,
		Assets:               assets,
		EnabledStreams:       enabledStreams,
		TLSEnabled:           cr.Spec.Graylog.TLS.HTTP.Enabled,
	}, nil
}

func (connector *GraylogConnector) DELETE(url string) (string, int, error) {
	return connector.Send(url, http.MethodDelete, "")
}

func (connector *GraylogConnector) GET(url string) (string, int, error) {
	return connector.Send(url, http.MethodGet, "")
}

func (connector *GraylogConnector) POST(url string, data string) (string, int, error) {
	return connector.Send(url, http.MethodPost, data)
}

func (connector *GraylogConnector) PUT(url string, data string) (string, int, error) {
	return connector.Send(url, http.MethodPut, data)
}

func (connector *GraylogConnector) Send(urlPath string, method string, data string) (string, int, error) {
	var err error
	protocol := "http"
	if connector.TLSEnabled {
		protocol = "https"
	}

	var u *url.URL
	u, err = url.Parse(urlPath)
	if err != nil {
		return "", -1, err
	}

	var base *url.URL
	base, err = url.Parse(protocol + "://" + connector.RestClient.Host)
	if err != nil {
		return "", -1, err
	}

	uri := base.ResolveReference(u)

	request, err := http.NewRequest(method, uri.String(), strings.NewReader(data))
	if err != nil {
		return "", -1, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Requested-By", "Graylog API Browser")

	connector.Log.V(util.Debug).Info("Send " + method + " request to: " + uri.String() + " with body: " + data)

	connector.RestClient.SetAuthHeader(request)
	var response *http.Response
	response, err = connector.RestClient.Client.Do(request)
	if err != nil {
		return "", -1, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return "", -1, err
	}
	connector.Log.V(util.Debug).Info("Status code: " + response.Status)
	connector.Log.V(util.Debug).Info("Response: " + buf.String())
	return buf.String(), response.StatusCode, nil
}

func ManageRequiredStreams(cr *loggingService.LoggingService) []loggingService.Stream {
	systemLogs := false
	auditLogs := false
	for i := range cr.Spec.Graylog.Streams {
		name := cr.Spec.Graylog.Streams[i].Name
		if strings.EqualFold(name, "system logs") {
			systemLogs = true
		}
		if strings.EqualFold(name, "audit logs") {
			auditLogs = true
		}
	}
	if !auditLogs {
		s := *new(loggingService.Stream)
		s.Name = util.GraylogAuditStream
		s.Install = true
		s.RotationStrategy = "timeBased"
		s.RotationPeriod = "P1M"
		s.MaxNumberOfIndices = 5
		cr.Spec.Graylog.Streams = append(cr.Spec.Graylog.Streams, s)
	}
	if !systemLogs {
		s := *new(loggingService.Stream)
		s.Name = util.GraylogSystemStream
		s.Install = true
		cr.Spec.Graylog.Streams = append(cr.Spec.Graylog.Streams, s)
	}
	return cr.Spec.Graylog.Streams
}

func GetStreams(cr *loggingService.LoggingService) []Stream {
	var enabledStreams []Stream
	var streams []loggingService.Stream
	if cr.Spec.Graylog != nil && len(cr.Spec.Graylog.Streams) == 0 {
		streams = GetDefaultStreams()
	} else {
		streams = ManageRequiredStreams(cr)
	}
	for i := range streams {
		var needSkip = false
		for j := i + 1; j < len(streams); j++ {
			if streams[i].Name == streams[j].Name {
				needSkip = true
			}
		}
		if !needSkip && streams[i].Install {
			for k := range util.GraylogStreamsRules {
				if streams[i].Name == k {
					rotationStrategy := "sizeBased"
					if streams[i].RotationStrategy != "" &&
						(strings.EqualFold(streams[i].RotationStrategy, "sizeBased") ||
							strings.EqualFold(streams[i].RotationStrategy, "timeBased")) {
						rotationStrategy = streams[i].RotationStrategy
					}
					enabledStreams = append(enabledStreams, Stream{
						Title:              streams[i].Name,
						RotationStrategy:   rotationStrategy,
						RotationPeriod:     streams[i].RotationPeriod,
						MaxSize:            streams[i].MaxSize,
						MaxNumberOfIndices: streams[i].MaxNumberOfIndices,
					})
					break
				}
			}
		}
	}
	return enabledStreams
}
