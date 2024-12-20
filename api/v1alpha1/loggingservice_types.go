/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EDIT THIS FILE! THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LoggingServiceSpec defines the desired state of LoggingService
type LoggingServiceSpec struct {
	Graylog                      *Graylog                      `json:"graylog,omitempty"`
	Fluentd                      *Fluentd                      `json:"fluentd,omitempty"`
	Fluentbit                    *Fluentbit                    `json:"fluentbit,omitempty"`
	CloudEventsReader            *CloudEventsReader            `json:"cloudEventsReader,omitempty"`
	MonitoringAgentLoggingPlugin *MonitoringAgentLoggingPlugin `json:"monitoringAgentLoggingPlugin,omitempty"`
	CloudURL                     string                        `json:"cloudURL,omitempty"`
	OSKind                       string                        `json:"osKind,omitempty"`
	ContainerRuntimeType         string                        `json:"containerRuntimeType,omitempty"`
	Ipv6                         bool                          `json:"ipv6,omitempty"`
	OpenshiftDeploy              bool                          `json:"openshiftDeploy,omitempty"`
}

// LoggingServiceCondition contains description of status of LoggingService
type LoggingServiceCondition struct {
	Type               string `json:"type"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Status             bool   `json:"status"`
}

// LoggingServiceStatus defines the observed state of LoggingService
type LoggingServiceStatus struct {
	Conditions []LoggingServiceCondition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// LoggingService is the Schema for the loggingservices API
type LoggingService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoggingServiceSpec   `json:"spec,omitempty"`
	Status LoggingServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoggingServiceList contains a list of LoggingService
type LoggingServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoggingService `json:"items"`
}

// Graylog contains Graylog-specific configuration
type Graylog struct {
	GraylogResources                         *v1.ResourceRequirements     `json:"graylogResources,omitempty"`
	MongoResources                           *v1.ResourceRequirements     `json:"mongoResources,omitempty"`
	InitResources                            *v1.ResourceRequirements     `json:"initResources,omitempty"`
	MongoDBUpgrade                           *MongoDBUpgrade              `json:"mongoDBUpgrade,omitempty"`
	AuthProxy                                *AuthProxy                   `json:"authProxy,omitempty"`
	TLS                                      *GraylogTLS                  `json:"tls,omitempty"`
	OpenSearch                               *OpenSearch                  `json:"openSearch,omitempty"`
	Annotations                              map[string]string            `json:"annotations,omitempty"`
	Labels                                   map[string]string            `json:"labels,omitempty"`
	NodeSelectorValue                        string                       `json:"nodeSelectorValue,omitempty"`
	PathRepo                                 string                       `json:"pathRepo,omitempty"`
	PriorityClassName                        string                       `json:"priorityClassName,omitempty"`
	DockerImage                              string                       `json:"dockerImage"`
	MongoDBImage                             string                       `json:"mongoDBImage"`
	LogLevel                                 string                       `json:"logLevel,omitempty"`
	ContentDeployPolicy                      string                       `json:"contentDeployPolicy"`
	JavaOpts                                 string                       `json:"javaOpts,omitempty"`
	ContentPackPaths                         string                       `json:"contentPackPaths,omitempty"`
	CustomPluginsPaths                       string                       `json:"customPluginsPaths,omitempty"`
	Host                                     string                       `json:"host"`
	InitSetupImage                           string                       `json:"initSetupImage"`
	User                                     string                       `json:"-"`
	Password                                 string                       `json:"-"`
	NodeSelectorKey                          string                       `json:"nodeSelectorKey,omitempty"`
	InitContainerDockerImage                 string                       `json:"initContainerDockerImage,omitempty"`
	GraylogSecretName                        string                       `json:"graylogSecretName"`
	ContentPacks                             []*ContentPackPathHTTPConfig `json:"contentPacks,omitempty"`
	Streams                                  []Stream                     `json:"streams,omitempty"`
	ProcessbufferProcessors                  int                          `json:"processbufferProcessors,omitempty"`
	OutputbufferProcessorThreadsMaxPoolSize  int                          `json:"outputbufferProcessorThreadsMaxPoolSize,omitempty"`
	RingSize                                 int                          `json:"ringSize,omitempty"`
	ElasticsearchMaxTotalConnectionsPerRoute int                          `json:"elasticsearchMaxTotalConnectionsPerRoute,omitempty"`
	ElasticsearchMaxTotalConnections         int                          `json:"elasticsearchMaxTotalConnections,omitempty"`
	OutputBatchSize                          int                          `json:"outputBatchSize,omitempty"`
	InputbufferRingSize                      int                          `json:"inputbufferRingSize,omitempty"`
	OutputbufferProcessors                   int                          `json:"outputbufferProcessors,omitempty"`
	MaxSize                                  int                          `json:"maxSize,omitempty"`
	InputbufferProcessors                    int                          `json:"inputbufferProcessors,omitempty"`
	StartupTimeout                           int                          `json:"startupTimeout,omitempty"`
	IndexShards                              int                          `json:"indexShards,omitempty"`
	IndexReplicas                            int                          `json:"indexReplicas,omitempty"`
	MaxNumberOfIndices                       int                          `json:"maxNumberOfIndices,omitempty"`
	LogsRotationSizeGb                       int                          `json:"logsRotationSizeGb,omitempty"`
	InputPort                                int                          `json:"inputPort"`
	S3Archive                                bool                         `json:"s3Archive"`
}

type ContentPackPathHTTPConfig struct {
	HTTPConfig *HTTPConfig `yaml:"http,omitempty" json:"tls,omitempty"`
	URL        string      `yaml:"url,omitempty" json:"url,omitempty"`
}

type OpenSearch struct {
	HTTPConfig *HTTPConfig `yaml:"http,omitempty" json:"tls,omitempty"`
	Host       string      `yaml:"url,omitempty" json:"url,omitempty"`
}

type HTTPConfig struct {
	Credentials *Credentials `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	TLSConfig   *TLSConfig   `yaml:"tlsConfig,omitempty" json:"tlsConfig,omitempty"`
}

type TLSConfig struct {
	CA                 *v1.SecretKeySelector `yaml:"ca,omitempty" json:"ca,omitempty"`
	Cert               *v1.SecretKeySelector `yaml:"cert,omitempty" json:"cert,omitempty"`
	Key                *v1.SecretKeySelector `yaml:"key,omitempty" json:"key,omitempty"`
	InsecureSkipVerify bool                  `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
}

type Credentials struct {
	User     *v1.SecretKeySelector `yaml:"username" json:"username"`
	Password *v1.SecretKeySelector `yaml:"password" json:"password"`
}

type TLS struct {
	GenerateCerts *GenerateCerts `json:"generateCerts,omitempty"`
	Certificates  `json:",inline"`
}

type Certificates struct {
	CA   *CA   `json:"ca,omitempty"`
	Cert *Cert `json:"cert,omitempty"`
	Key  *Key  `json:"key,omitempty"`
}

// Fluentd contains Fluentd-specific configuration
type Fluentd struct {
	Resources                  *v1.ResourceRequirements `json:"resources,omitempty"`
	Annotations                map[string]string        `json:"annotations,omitempty"`
	Labels                     map[string]string        `json:"labels,omitempty"`
	ExtraFields                map[string]string        `json:"extraFields,omitempty"`
	CustomFilterConf           string                   `json:"customFilterConf"`
	SystemLogType              string                   `json:"systemLogType"`
	CloudEventsReaderFormat    string                   `json:"cloudEventsReaderFormat,omitempty"`
	GraylogHost                string                   `json:"graylogHost,omitempty"`
	DockerImage                string                   `json:"dockerImage"`
	ConfigmapReload            *ConfigmapReload         `json:"configmapReload,omitempty"`
	GraylogProtocol            string                   `json:"graylogProtocol,omitempty"`
	PriorityClassName          string                   `json:"priorityClassName,omitempty"`
	MultilineFirstLineRegexp   string                   `json:"multilineFirstLineRegexp,omitempty"`
	LogLevel                   string                   `json:"logLevel,omitempty"`
	TotalLimitSize             string                   `json:"totalLimitSize,omitempty"`
	CustomInputConf            string                   `json:"customInputConf"`
	NodeSelectorKey            string                   `json:"nodeSelectorKey,omitempty"`
	CustomOutputConf           string                   `json:"customOutputConf"`
	NodeSelectorValue          string                   `json:"nodeSelectorValue,omitempty"`
	TLS                        FluentdTLS               `json:"tls,omitempty"`
	AdditionalVolumeMounts     []v1.VolumeMount         `json:"additionalVolumeMounts,omitempty"`
	ExcludePath                []string                 `json:"excludePath,omitempty"`
	AdditionalVolumes          []v1.Volume              `json:"additionalVolumes,omitempty"`
	Tolerations                []v1.Toleration          `json:"tolerations,omitempty"`
	QueueLimitLength           int                      `json:"queueLimitLength,omitempty"`
	GraylogPort                int                      `json:"graylogPort,omitempty"`
	BillCycleConf              bool                     `json:"billCycleConf,omitempty"`
	SystemLogging              bool                     `json:"systemLogging,omitempty"`
	SystemAuditLogging         bool                     `json:"systemAuditLogging,omitempty"`
	KubeAuditLogging           bool                     `json:"kubeAuditLogging,omitempty"`
	KubeApiserverAuditLogging  bool                     `json:"kubeApiserverAuditLogging,omitempty"`
	ContainerLogging           bool                     `json:"containerLogging,omitempty"`
	WatchKubernetesMetadata    bool                     `json:"watchKubernetesMetadata,omitempty"`
	SecurityContextPrivileged  bool                     `json:"securityContextPrivileged,omitempty"`
	FileStorage                bool                     `json:"useFileStorage,omitempty"`
	GraylogOutput              bool                     `json:"graylogOutput"`
	GraylogBufferFlushInterval string                   `json:"graylogBufferFlushInterval,omitempty"`
	Compress                   string                   `json:"compress,omitempty"`
	MockKubeData               bool                     `json:"mockKubeData,omitempty"`
	Output                     *OutputFluentd           `json:"output,omitempty"`
}

// Fluentbit contains Fluentbit-specific configuration
type Fluentbit struct {
	Resources                 *v1.ResourceRequirements `json:"resources,omitempty"`
	Annotations               map[string]string        `json:"annotations,omitempty"`
	Labels                    map[string]string        `json:"labels,omitempty"`
	ExtraFields               map[string]string        `json:"extraFields,omitempty"`
	Aggregator                *FluentbitAggregator     `json:"aggregator,omitempty"`
	MemBufLimit               string                   `json:"memBufLimit,omitempty"`
	SystemLogType             string                   `json:"systemLogType"`
	GraylogHost               string                   `json:"graylogHost,omitempty"`
	NodeSelectorValue         string                   `json:"nodeSelectorValue,omitempty"`
	GraylogProtocol           string                   `json:"graylogProtocol,omitempty"`
	DockerImage               string                   `json:"dockerImage"`
	ConfigmapReload           *ConfigmapReload         `json:"configmapReload,omitempty"`
	PriorityClassName         string                   `json:"priorityClassName,omitempty"`
	TotalLimitSize            string                   `json:"totalLimitSize,omitempty"`
	CustomInputConf           string                   `json:"customInputConf"`
	CustomFilterConf          string                   `json:"customFilterConf"`
	CustomOutputConf          string                   `json:"customOutputConf"`
	CustomLuaScriptConf       map[string]string        `json:"customLuaScriptConf,omitempty"`
	LogLevel                  string                   `json:"logLevel,omitempty"`
	MultilineFirstLineRegexp  string                   `json:"multilineFirstLineRegexp,omitempty"`
	NodeSelectorKey           string                   `json:"nodeSelectorKey,omitempty"`
	MultilineOtherLinesRegexp string                   `json:"multilineOtherLinesRegexp,omitempty"`
	TLS                       FluentbitTLS             `json:"tls,omitempty"`
	AdditionalVolumes         []v1.Volume              `json:"additionalVolumes,omitempty"`
	AdditionalVolumeMounts    []v1.VolumeMount         `json:"additionalVolumeMounts,omitempty"`
	Tolerations               []v1.Toleration          `json:"tolerations,omitempty"`
	GraylogPort               int                      `json:"graylogPort,omitempty"`
	SecurityContextPrivileged bool                     `json:"securityContextPrivileged,omitempty"`
	WatchKubernetesMetadata   bool                     `json:"watchKubernetesMetadata,omitempty"`
	MockKubeData              bool                     `json:"mockKubeData,omitempty"`
	SystemLogging             bool                     `json:"systemLogging"`
	BillCycleConf             bool                     `json:"billCycleConf,omitempty"`
	GraylogOutput             bool                     `json:"graylogOutput"`
	SystemAuditLogging        bool                     `json:"systemAuditLogging,omitempty"`
	KubeAuditLogging          bool                     `json:"kubeAuditLogging,omitempty"`
	KubeApiserverAuditLogging bool                     `json:"kubeApiserverAuditLogging,omitempty"`
	ContainerLogging          bool                     `json:"containerLogging,omitempty"`
	ExcludePath               string                   `json:"excludePath,omitempty"`
	Output                    *OutputFluentbit         `json:"output,omitempty"`
}

// FluentbitAggregator contains Fluentbit-aggregator-specific configuration
type FluentbitAggregator struct {
	Labels                    map[string]string        `json:"labels,omitempty"`
	Volume                    *Volume                  `json:"volume,omitempty"`
	Resources                 *v1.ResourceRequirements `json:"resources,omitempty"`
	Annotations               map[string]string        `json:"annotations,omitempty"`
	ExtraFields               map[string]string        `json:"extraFields,omitempty"`
	MemBufLimit               string                   `json:"memBufLimit,omitempty"`
	DockerImage               string                   `json:"dockerImage"`
	ConfigmapReload           *ConfigmapReload         `json:"configmapReload,omitempty"`
	NodeSelectorValue         string                   `json:"nodeSelectorValue,omitempty"`
	MultilineOtherLinesRegexp string                   `json:"multilineOtherLinesRegexp,omitempty"`
	GraylogHost               string                   `json:"graylogHost,omitempty"`
	MultilineFirstLineRegexp  string                   `json:"multilineFirstLineRegexp,omitempty"`
	GraylogProtocol           string                   `json:"graylogProtocol,omitempty"`
	NodeSelectorKey           string                   `json:"nodeSelectorKey,omitempty"`
	PriorityClassName         string                   `json:"priorityClassName,omitempty"`
	TotalLimitSize            string                   `json:"totalLimitSize,omitempty"`
	CustomFilterConf          string                   `json:"customFilterConf"`
	CustomOutputConf          string                   `json:"customOutputConf"`
	CustomLuaScriptConf       map[string]string        `json:"customLuaScriptConf,omitempty"`
	TLS                       FluentbitTLS             `json:"tls,omitempty"`
	Tolerations               []v1.Toleration          `json:"tolerations,omitempty"`
	StartupTimeout            int                      `json:"startupTimeout,omitempty"`
	Replicas                  int                      `json:"replicas,omitempty"`
	GraylogPort               int                      `json:"graylogPort,omitempty"`
	Install                   bool                     `json:"install"`
	SecurityContextPrivileged bool                     `json:"securityContextPrivileged,omitempty"`
	GraylogOutput             bool                     `json:"graylogOutput,omitempty"`
	Output                    *OutputFluentbit         `json:"output,omitempty"`
}

// CloudEventsReader contains EventsReader-specific configuration
type CloudEventsReader struct {
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	DockerImage       string                   `json:"dockerImage"`
	PriorityClassName string                   `json:"priorityClassName,omitempty"`
	NodeSelectorKey   string                   `json:"nodeSelectorKey,omitempty"`
	NodeSelectorValue string                   `json:"nodeSelectorValue,omitempty"`
	Labels            map[string]string        `json:"labels,omitempty"`
	Annotations       map[string]string        `json:"annotations,omitempty"`
	Args              []string                 `json:"args,omitempty"`
	Install           bool                     `json:"install"`
}

// MonitoringAgentLoggingPlugin contains MonitoringAgentLoggingPlugin-specific configuration
//+kubebuilder:deprecatedversion:warning="MonitoringAgentLoggingPlugin section is no longer need and use"
type MonitoringAgentLoggingPlugin struct {
	Resources          *v1.ResourceRequirements `json:"resources,omitempty"`
	Annotations        map[string]string        `json:"annotations,omitempty"`
	Labels             map[string]string        `json:"labels,omitempty"`
	InfluxDBName       string                   `json:"influxDBName,omitempty"`
	InfluxDBSecretName string                   `json:"influxDBSecretName,omitempty"`
	InfluxDBHost       string                   `json:"influxDBHost,omitempty"`
	NodeSelectorKey    string                   `json:"nodeSelectorKey,omitempty"`
	NodeSelectorValue  string                   `json:"nodeSelectorValue,omitempty"`
	SaSecret           string                   `json:"saSecret"`
	SaSecretVolume     string                   `json:"saSecretVolume"`
	PriorityClassName  string                   `json:"priorityClassName,omitempty"`
	DockerImage        string                   `json:"dockerImage"`
	InfluxDBPort       int                      `json:"influxDBPort,omitempty"`
	InfluxDBMode       bool                     `json:"influxDBMode"`
}

type GraylogTLS struct {
	HTTP  *HTTPGraylogTLS  `json:"http,omitempty"`
	Input *InputGraylogTLS `json:"input,omitempty"`
}

type HTTPGraylogTLS struct {
	GenerateCerts      *GenerateCerts `json:"generateCerts,omitempty"`
	Cert               *Cert          `json:"cert,omitempty"`
	Key                *Key           `json:"key,omitempty"`
	CACerts            string         `json:"cacerts,omitempty"`
	KeyFilePassword    string         `json:"keyFilePassword,omitempty"`
	Enabled            bool           `json:"enabled,omitempty"`
	InsecureSkipVerify bool           `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
}

type InputGraylogTLS struct {
	TLS                `json:",inline"`
	KeyFilePassword    string `json:"keyFilePassword,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
}

type FluentdTLS struct {
	TLS              `json:",inline"`
	FluentdTLSParams `json:",inline"`
}

type FluentdTLSParams struct {
	Version         string `json:"version,omitempty"`
	Enabled         bool   `json:"enabled,omitempty"`
	NoDefaultCA     bool   `json:"noDefaultCA,omitempty"`
	AllCiphers      bool   `json:"allCiphers,omitempty"`
	RescueSslErrors bool   `json:"rescueSslErrors,omitempty"`
	NoVerify        bool   `json:"noVerify,omitempty"`
}

type FluentdLokiTLS struct {
	Certificates     `json:",inline"`
	FluentdTLSParams `json:",inline"`
}

type FluentbitTLS struct {
	TLS                `json:",inline"`
	FluentbitTLSParams `json:",inline"`
}

type FluentbitTLSParams struct {
	KeyPasswd string `json:"keyPasswd,omitempty"`
	Enabled   bool   `json:"enabled,omitempty"`
	Verify    bool   `json:"verify,omitempty"`
}

type FluentbitLokiTLS struct {
	Certificates       `json:",inline"`
	FluentbitTLSParams `json:",inline"`
}

type AuthProxy struct {
	Resources          *v1.ResourceRequirements `json:"resources,omitempty"`
	BindPasswordSecret *v1.SecretKeySelector    `json:"bindPasswordSecret,omitempty"`
	// CA contains selectors for the Secret containing TLS certificate for LDAP database or OAuth authentication server
	CA CA `json:"ca,omitempty"`
	// Cert contains selectors for the Secret containing TLS certificate for client authentication
	// to LDAP database or OAuth authentication server
	Cert Cert `json:"cert,omitempty"`
	// Key contains selectors for the Secret containing TLS private key for client authentication
	// to LDAP database or OAuth authentication server
	Key     Key    `json:"key,omitempty"`
	Image   string `json:"image"`
	Install bool   `json:"install"`
}

type Volume struct {
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageSize      string `json:"storageSize,omitempty"`
	Bind             bool   `json:"bind,omitempty"`
}

type CA struct {
	SecretName string `json:"secretName,omitempty"`
	SecretKey  string `json:"secretKey,omitempty"`
}

type Cert struct {
	SecretName string `json:"secretName,omitempty"`
	SecretKey  string `json:"secretKey,omitempty"`
}

type Key struct {
	SecretName string `json:"secretName,omitempty"`
	SecretKey  string `json:"secretKey,omitempty"`
}

type Stream struct {
	Name               string `json:"name"`
	RotationStrategy   string `json:"rotationStrategy,omitempty"`
	RotationPeriod     string `json:"rotationPeriod,omitempty"`
	MaxSize            int    `json:"maxSize,omitempty"`
	MaxNumberOfIndices int    `json:"maxNumberOfIndices,omitempty"`
	Install            bool   `json:"install"`
}

// GenerateCerts define settings for cert-manager.
type GenerateCerts struct {
	SecretName string `json:"secretName,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// MongoDBUpgrade is used for the sequential MongoDB upgrading from 3.6 to 5.0
type MongoDBUpgrade struct {
	MongoDBImage40 string `json:"mongoDBImage40"`
	MongoDBImage42 string `json:"mongoDBImage42"`
	MongoDBImage44 string `json:"mongoDBImage44"`
}

type LoggingServiceParameters struct {
	Release
	Values LoggingServiceSpec
}

type Release struct {
	Namespace string
}

type ConfigmapReload struct {
	DockerImage string                   `json:"dockerImage"`
	Resources   *v1.ResourceRequirements `json:"resources,omitempty"`
}

type OutputFluentbit struct {
	Loki *LokiFluentbit `json:"loki,omitempty"`
}

type LokiFluentbit struct {
	Enabled       bool              `json:"enabled,omitempty"`
	Host          string            `json:"host,omitempty"`
	Tenant        string            `json:"tenant,omitempty"`
	Auth          *Auth             `json:"auth,omitempty"`
	StaticLabels  string            `json:"staticLabels,omitempty"`
	LabelsMapping string            `json:"labelsMapping,omitempty"`
	TLS           *FluentbitLokiTLS `json:"tls,omitempty"`
	ExtraParams   string            `json:"extraParams,omitempty"`
}

type Auth struct {
	Token    *v1.SecretKeySelector `yaml:"token" json:"token,omitempty"`
	User     *v1.SecretKeySelector `yaml:"username" json:"user,omitempty"`
	Password *v1.SecretKeySelector `yaml:"password" json:"password,omitempty"`
}

type OutputFluentd struct {
	Loki *LokiFluentd `json:"loki,omitempty"`
}

type LokiFluentd struct {
	Enabled       bool            `json:"enabled,omitempty"`
	Host          string          `json:"host,omitempty"`
	Tenant        string          `json:"tenant,omitempty"`
	Auth          *Auth           `json:"auth,omitempty"`
	StaticLabels  string          `json:"staticLabels,omitempty"`
	LabelsMapping string          `json:"labelsMapping,omitempty"`
	TLS           *FluentdLokiTLS `json:"tls,omitempty"`
	ExtraParams   string          `json:"extraParams,omitempty"`
}

func (in *LoggingService) ToParams() LoggingServiceParameters {
	return LoggingServiceParameters{
		Values: in.Spec,
		Release: Release{
			Namespace: in.GetNamespace(),
		},
	}
}

func init() {
	SchemeBuilder.Register(&LoggingService{}, &LoggingServiceList{})
}

func (in *Graylog) IsForceUpdate() bool {
	return in.ContentDeployPolicy == "force-update"
}

func (in *Graylog) IsOnlyCreate() bool {
	return in.ContentDeployPolicy == "only-create"
}

func (in *Graylog) IsInstall() bool {
	return in != nil
}

func (in *Fluentd) IsInstall() bool {
	return in != nil
}

func (in *Fluentbit) IsInstall() bool {
	return in != nil
}

func (in *CloudEventsReader) IsInstall() bool {
	return in != nil
}

func (in *MonitoringAgentLoggingPlugin) IsInstall() bool {
	return in != nil && in.InfluxDBMode
}

func (graylogTLS *HTTPGraylogTLS) GetCertificates(ctx context.Context, clientSet kubernetes.Interface, namespace string) (tlsConf *tls.Config, err error) {
	if len(namespace) == 0 {
		err = errors.New("namespace is not setup")
		return
	}

	if graylogTLS.InsecureSkipVerify {
		tlsConf = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		var certValue, pkey string
		var secret *v1.Secret
		if len(graylogTLS.CACerts) != 0 {
			secret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, graylogTLS.CACerts, metav1.GetOptions{})
			if err != nil {
				return nil, err
			} else {
				if len(secret.Data) == 0 {
					err = fmt.Errorf("can't find Certificate Authority %s", graylogTLS.CACerts)
					return
				}
				if graylogTLS.Cert != nil && graylogTLS.Cert.SecretName != "" && graylogTLS.Key != nil && graylogTLS.Key.SecretName != "" {
					if value, containsKey := secret.Data[graylogTLS.Cert.SecretKey]; containsKey {
						certValue = string(value)
					} else {
						var certSecret *v1.Secret
						certSecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, graylogTLS.Cert.SecretName, metav1.GetOptions{})
						if err != nil {
							return nil, err
						} else {
							certValue = string(certSecret.Data[graylogTLS.Cert.SecretKey])
						}
					}

					if value, containsKey := secret.Data[graylogTLS.Key.SecretKey]; containsKey {
						pkey = string(value)
					} else {
						var pKeySecret *v1.Secret
						pKeySecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, graylogTLS.Key.SecretName, metav1.GetOptions{})
						if err != nil {
							return nil, err
						} else {
							pkey = string(pKeySecret.Data[graylogTLS.Key.SecretKey])
						}
					}
				}
			}

			caCertPool := x509.NewCertPool()
			for _, caCert := range secret.Data {
				ok := caCertPool.AppendCertsFromPEM(caCert)
				if !ok {
					err = errors.New("can't parse Certificate Authority")
					return
				}
			}

			if pkey != "" && certValue != "" {
				var clientCert tls.Certificate
				clientCert, err = tls.X509KeyPair([]byte(certValue), []byte(pkey))
				if err != nil {
					return
				}
				tlsConf = &tls.Config{
					RootCAs:      caCertPool,
					Certificates: []tls.Certificate{clientCert},
				}
			} else {
				tlsConf = &tls.Config{
					RootCAs: caCertPool,
				}
			}
		} else {
			if graylogTLS.GenerateCerts != nil && graylogTLS.GenerateCerts.Enabled {
				secret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, graylogTLS.GenerateCerts.SecretName, metav1.GetOptions{})
				if err != nil {
					return nil, err
				} else {
					if len(secret.Data) == 0 {
						err = fmt.Errorf("can't find generate certificates %s", graylogTLS.GenerateCerts.SecretName)
						return
					}

					caCertPool := x509.NewCertPool()
					if secret.Type == v1.SecretTypeTLS {
						for key, cert := range secret.Data {
							if key == v1.ServiceAccountRootCAKey {
								ok := caCertPool.AppendCertsFromPEM(cert)
								if !ok {
									err = errors.New("can't parse Certificate Authority")
									return
								}
							}
							if key == v1.TLSCertKey {
								certValue = string(cert)
							}
							if key == v1.TLSPrivateKeyKey {
								pkey = string(cert)
							}
						}
					}

					if pkey != "" && certValue != "" {
						var clientCert tls.Certificate
						clientCert, err = tls.X509KeyPair([]byte(certValue), []byte(pkey))
						if err != nil {
							return
						}
						tlsConf = &tls.Config{
							RootCAs:      caCertPool,
							Certificates: []tls.Certificate{clientCert},
						}
					} else {
						tlsConf = &tls.Config{
							RootCAs: caCertPool,
						}
					}
				}
			} else {
				err = errors.New("can't find any certificate configured")
			}
		}
	}

	return
}

func (HTTPConfig *HTTPConfig) GetCredentialsAndCertificates(ctx context.Context, clientSet kubernetes.Interface, namespace string) (name, pwd, token string, tlsConf *tls.Config, err error) {
	if len(namespace) == 0 {
		err = errors.New("namespace is not setup")
		return
	}
	if HTTPConfig.Credentials != nil {
		name, pwd, token, err = HTTPConfig.Credentials.GetCredentials(ctx, clientSet, namespace)
		if err != nil {
			return
		}
	}

	if HTTPConfig.TLSConfig == nil || HTTPConfig.TLSConfig.InsecureSkipVerify {
		tlsConf = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tlsConf, err = HTTPConfig.TLSConfig.GetCertificates(ctx, clientSet, namespace)
	}

	return
}

func (credentials *Credentials) GetCredentials(ctx context.Context, clientSet kubernetes.Interface, namespace string) (name, pwd, token string, err error) {
	if len(namespace) == 0 {
		err = errors.New("namespace is not setup")
		return
	}
	switch {
	case credentials.User != nil && credentials.User.Name != "" && credentials.Password != nil && credentials.Password.Name != "":
		var userSecret *v1.Secret
		userSecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, credentials.User.Name, metav1.GetOptions{})
		if err != nil {
			return "", "", "", err
		}

		var pwdSecret *v1.Secret
		if credentials.User.Name != credentials.Password.Name {
			pwdSecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, credentials.Password.Name, metav1.GetOptions{})
			if err != nil {
				return "", "", "", err
			}
		} else {
			pwdSecret = userSecret
		}
		name = string(pwdSecret.Data[credentials.User.Key])
		pwd = string(pwdSecret.Data[credentials.Password.Key])
		return
	default:
		return "", "", "", errors.New("authorization data is empty or provided partially")
	}
}

func (tlsConfig *TLSConfig) GetCertificates(ctx context.Context, clientSet kubernetes.Interface, namespace string) (tlsConf *tls.Config, err error) {
	if len(namespace) == 0 {
		err = errors.New("namespace is not setup")
		return
	}

	var certValue, pkey string
	var caCert []byte
	if tlsConfig.CA != nil && tlsConfig.CA.Name != "" && tlsConfig.CA.Key != "" && namespace != "" {
		var caSecret *v1.Secret
		caSecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, tlsConfig.CA.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		} else {
			var found bool
			caCert, found = caSecret.Data[tlsConfig.CA.Key]
			if !found {
				err = fmt.Errorf("can't find Certificate Authority with key %s", tlsConfig.CA.Key)
				return
			}

			if tlsConfig.Cert != nil && tlsConfig.Cert.Name != "" && tlsConfig.Key != nil && tlsConfig.Key.Name != "" {
				if tlsConfig.Cert.Name != tlsConfig.CA.Name {
					var certSecret *v1.Secret
					certSecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, tlsConfig.Cert.Name, metav1.GetOptions{})
					if err != nil {
						return nil, err
					} else {
						certValue = string(certSecret.Data[tlsConfig.Cert.Key])
					}
				} else {
					certValue = string(caSecret.Data[tlsConfig.Cert.Key])
				}
				if tlsConfig.Key.Name != tlsConfig.CA.Name {
					var pKeySecret *v1.Secret
					pKeySecret, err = clientSet.CoreV1().Secrets(namespace).Get(ctx, tlsConfig.Key.Name, metav1.GetOptions{})
					if err != nil {
						return nil, err
					} else {
						pkey = string(pKeySecret.Data[tlsConfig.Key.Key])
					}
				} else {
					pkey = string(caSecret.Data[tlsConfig.Key.Key])
				}
			}
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			err = errors.New("can't parse Certificate Authority")
			return
		}
		if pkey != "" && certValue != "" {
			var clientCert tls.Certificate
			clientCert, err = tls.X509KeyPair([]byte(certValue), []byte(pkey))
			if err != nil {
				return
			}
			tlsConf = &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{clientCert},
			}
		} else {
			tlsConf = &tls.Config{
				RootCAs: caCertPool,
			}
		}
	}

	return
}
