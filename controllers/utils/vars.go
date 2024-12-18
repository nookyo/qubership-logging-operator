package utils

import (
	"path"
	"time"
)

var (
	BasePath = "assets/"

	LoggingServiceStatus = "ReconcileCycleStatus"

	FluentdComponentName      = "logging-fluentd"
	FluentdStatus             = "ReconcileFluentdStatus"
	FluentdConfigMapDirectory = "fluentd.configmap/conf.d"
	FluentdDaemonSet          = path.Join(BasePath, "daemon-set.yaml")
	FluentdServiceTemplate    = path.Join(BasePath, "service.yaml")
	FluentdLabels             = map[string]string{"component": "logging-fluentd"}

	FluentbitComponentName = "logging-fluentbit"
	FluentbitStatus        = "ReconcileFluentbitStatus"
	FluentbitDaemonSet     = path.Join(BasePath, "daemon-set.yaml")
	FluentbitService       = path.Join(BasePath, "service.yaml")
	FluentbitLabels        = map[string]string{"component": "logging-fluentbit"}

	ForwarderFluentbitComponentName  = "logging-fluentbit-forwarder"
	HAFluentStatus                   = "ReconcileHAFluentStatus"
	ForwarderFluentbitDaemonSet      = path.Join(BasePath, "flb-forwarder-daemon-set.yaml")
	ForwarderFluentbitService        = path.Join(BasePath, "flb-forwarder-service.yaml")
	ForwarderFluentbitLabels         = map[string]string{"component": "logging-fluentbit-forwarder"}
	AggregatorFluentbitComponentName = "logging-fluentbit-aggregator"
	AggregatorFluentbitStatefulSet   = path.Join(BasePath, "flb-aggregator-stateful-set.yaml")
	AggregatorFluentbitService       = path.Join(BasePath, "flb-aggregator-service.yaml")
	AggregatorFluentbitConfigMap     = path.Join(BasePath, "flb-aggregator-configmap.yaml")
	AggregatorFluentbitLabels        = map[string]string{"component": "logging-fluentbit-aggregator"}

	ForwarderFluentbitConfigMapDirectory  = "forwarder.configmap/conf.d"
	AggregatorFluentbitConfigMapDirectory = "aggregator.configmap/conf.d"
	FluentbitConfigMapDirectory           = "fluentbit.configmap/conf.d"

	EventsReaderComponentName = "events-reader"
	EventsReaderStatus        = "ReconcileEventsReaderStatus"
	EventsReaderDeployment    = path.Join(BasePath, "deployment.yaml")
	EventsReaderService       = path.Join(BasePath, "service.yaml")

	MonitoringAgentComponentName = "monitoring-agent-logging-plugin"
	MonitoringAgentStatus        = "ReconcileMonitoringAgentLoggingPluginStatus"
	MonitoringAgentSecret        = path.Join(BasePath, "secret.yaml")
	MonitoringAgentDeployment    = path.Join(BasePath, "deployment.yaml")
	MonitoringAgentService       = path.Join(BasePath, "service.yaml")

	GraylogComponentName            = "graylog-service"
	GraylogServiceAccountName       = "logging-graylog"
	GraylogDeploymentName           = "graylog"
	GraylogStatefulsetName          = "graylog"
	GraylogClaimName                = "graylog-claim"
	MongoClaimName                  = "mongo-claim"
	GraylogStatus                   = "ReconcileGraylogStatus"
	GraylogConfig                   = "config/"
	GraylogServiceAccount           = path.Join(BasePath, "service-account.yaml")
	GraylogStatefulset              = path.Join(BasePath, "statefulset.yaml")
	GraylogService                  = path.Join(BasePath, "service.yaml")
	GraylogConfigMapDirectory       = path.Join(GraylogConfig, "configmap")
	GraylogGrokPatterns             = path.Join(GraylogConfig, "grok_patterns.json")
	GraylogDefaultStream            = "Default Stream"
	GraylogAllMessagesStream        = "All messages"
	GraylogAuditIndexSet            = "Audit index set"
	GraylogAuditStream              = "Audit logs"
	GraylogSystemIndexSet           = "Default index set"
	GraylogSystemStream             = "System logs"
	GraylogAccessIndexSet           = "Access index set"
	GraylogAccessStream             = "Access logs"
	GraylogIntegrationIndexSet      = "Integration index set"
	GraylogIntegrationStream        = "Integration logs"
	GraylogBillCycleIndexSet        = "Bill Cycle index set"
	GraylogBillCycleStream          = "Bill Cycle logs"
	GraylogNginxIndexSet            = "Ingress-nginx index set"
	GraylogNginxStream              = "Nginx logs"
	GraylogDefaultIndexSet          = "Default index set"
	GraylogKubernetesEventsStream   = "Kubernetes events"
	GraylogKubernetesEventsIndexSet = "Kubernetes events index set"
	GraylogStreamsIndexTitles       = map[string]string{
		GraylogAuditStream:            GraylogAuditIndexSet,
		GraylogSystemStream:           GraylogSystemIndexSet,
		GraylogAccessStream:           GraylogAccessIndexSet,
		GraylogIntegrationStream:      GraylogIntegrationIndexSet,
		GraylogBillCycleStream:        GraylogBillCycleIndexSet,
		GraylogNginxStream:            GraylogNginxIndexSet,
		GraylogKubernetesEventsStream: GraylogKubernetesEventsIndexSet,
	}
	GraylogIndexConfigs = map[string]string{
		GraylogAuditIndexSet:            path.Join(GraylogConfig, "indexes/audit_index.json"),
		GraylogSystemIndexSet:           path.Join(GraylogConfig, "indexes/default_index.json"),
		GraylogAccessIndexSet:           path.Join(GraylogConfig, "indexes/access_index.json"),
		GraylogIntegrationIndexSet:      path.Join(GraylogConfig, "indexes/integration_index.json"),
		GraylogBillCycleIndexSet:        path.Join(GraylogConfig, "indexes/bill_cycle_index.json"),
		GraylogNginxIndexSet:            path.Join(GraylogConfig, "indexes/nginx_index.json"),
		GraylogDefaultIndexSet:          path.Join(GraylogConfig, "indexes/default_index.json"),
		GraylogKubernetesEventsIndexSet: path.Join(GraylogConfig, "indexes/k8s_event_index.json"),
	}
	GraylogStreamsDescriptions = map[string]string{
		GraylogAuditStream:            "Audit log messages from OC nodes sent through fluent bit",
		GraylogSystemStream:           "System log messages from OC nodes sent through fluent bit",
		GraylogAccessStream:           "Access log messages from OC nodes sent through fluent bit",
		GraylogIntegrationStream:      "Integration log messages from OC nodes sent through fluent bit",
		GraylogBillCycleStream:        "Bill Cycle log messages from OC nodes sent through fluent bit",
		GraylogNginxStream:            "Nginx log messages from OC nodes sent through fluent bit",
		GraylogKubernetesEventsStream: "Kubernetes events as logs sent through fluent bit",
	}
	GraylogInput                               = path.Join(GraylogConfig, "input.json")
	GraylogMessageProcessors                   = path.Join(GraylogConfig, "messageprocessors.json")
	GraylogAuthHeader                          = path.Join(GraylogConfig, "auth_header.json")
	GraylogArchivesDirectory                   = path.Join(GraylogConfig, "archives.json")
	GraylogReplaceTimestampExtractorName       = "replace_timestamp"
	GraylogKubernetesExtractorName             = "kuber_extractor"
	GraylogKubernetesLabelsExtractorName       = "kuber_labels_extractor"
	GraylogDockerExtractorName                 = "docker_extractor"
	GraylogMicroserviceFrameworkExtractorName  = "Microservice Framework Backend"
	GraylogReplaceTimestampExtractorAsset      = "replace_timestamp_extractor.json"
	GraylogKubernetesExtractorAsset            = "kubernetes_extractor.json"
	GraylogKubernetesLabelsExtractorAsset      = "kubernetes_labels_extractor.json"
	GraylogDockerExtractorAsset                = "docker_extractor.json"
	GraylogMicroserviceFrameworkExtractorAsset = "microservice_framework_extractor.json"
	Graylog4ExtractorsBasePath                 = path.Join(GraylogConfig, "extractors/graylog_4/")
	Graylog5ExtractorsBasePath                 = path.Join(GraylogConfig, "extractors/graylog_5/")
	Graylog4Extractors                         = map[string]string{
		GraylogReplaceTimestampExtractorName:      path.Join(Graylog4ExtractorsBasePath, GraylogReplaceTimestampExtractorAsset),
		GraylogKubernetesExtractorName:            path.Join(Graylog4ExtractorsBasePath, GraylogKubernetesExtractorAsset),
		GraylogKubernetesLabelsExtractorName:      path.Join(Graylog4ExtractorsBasePath, GraylogKubernetesLabelsExtractorAsset),
		GraylogDockerExtractorName:                path.Join(Graylog4ExtractorsBasePath, GraylogDockerExtractorAsset),
		GraylogMicroserviceFrameworkExtractorName: path.Join(Graylog4ExtractorsBasePath, GraylogMicroserviceFrameworkExtractorAsset),
	}
	Graylog5Extractors = map[string]string{
		GraylogReplaceTimestampExtractorName:      path.Join(Graylog5ExtractorsBasePath, GraylogReplaceTimestampExtractorAsset),
		GraylogKubernetesExtractorName:            path.Join(Graylog5ExtractorsBasePath, GraylogKubernetesExtractorAsset),
		GraylogKubernetesLabelsExtractorName:      path.Join(Graylog5ExtractorsBasePath, GraylogKubernetesLabelsExtractorAsset),
		GraylogDockerExtractorName:                path.Join(Graylog5ExtractorsBasePath, GraylogDockerExtractorAsset),
		GraylogMicroserviceFrameworkExtractorName: path.Join(Graylog5ExtractorsBasePath, GraylogMicroserviceFrameworkExtractorAsset),
	}
	GraylogAuditProcessingRule              = "Route Audit logs"
	GraylogSystemLogsProcessingRule         = "Route System logs"
	GraylogRemoveKubernetesRule             = "Remove kubernetes field"
	GraylogRemoveKubernetesLabelsRule       = "Remove kubernetes_labels field"
	GraylogUnsupportedSymbolsRule           = "Processing unsupported symbols"
	GraylogKubernetesEventsRule             = "Route Kubernetes events"
	GraylogRemoveKubernetesRuleStream       = GraylogDefaultStream
	GraylogRemoveKubernetesLabelsRuleStream = GraylogDefaultStream
	GraylogUnsupportedSymbolsRuleStream     = GraylogDefaultStream
	GraylogAccessProcessingRule             = "Route Access logs"
	GraylogIntegrationProcessingRule        = "Route Integration logs"
	GraylogBillCycleProcessingRule          = "Route Bill Cycle logs"
	GraylogNginxProcessingRule              = "Route Nginx logs"
	GraylogStreamsRules                     = map[string]string{
		GraylogAuditStream:                      GraylogAuditProcessingRule,
		GraylogSystemStream:                     GraylogSystemLogsProcessingRule,
		GraylogAccessStream:                     GraylogAccessProcessingRule,
		GraylogIntegrationStream:                GraylogIntegrationProcessingRule,
		GraylogBillCycleStream:                  GraylogBillCycleProcessingRule,
		GraylogNginxStream:                      GraylogNginxProcessingRule,
		GraylogRemoveKubernetesRuleStream:       GraylogRemoveKubernetesRule,
		GraylogRemoveKubernetesLabelsRuleStream: GraylogRemoveKubernetesLabelsRule,
		GraylogUnsupportedSymbolsRuleStream:     GraylogUnsupportedSymbolsRule,
		GraylogKubernetesEventsStream:           GraylogKubernetesEventsRule,
	}
	GraylogRuleConfigs = map[string]string{
		GraylogAuditProcessingRule:        path.Join(GraylogConfig, "processing_rules/audit_logs.rule"),
		GraylogSystemLogsProcessingRule:   path.Join(GraylogConfig, "processing_rules/system_logs.rule"),
		GraylogAccessProcessingRule:       path.Join(GraylogConfig, "processing_rules/access_logs.rule"),
		GraylogIntegrationProcessingRule:  path.Join(GraylogConfig, "processing_rules/int_logs.rule"),
		GraylogBillCycleProcessingRule:    path.Join(GraylogConfig, "processing_rules/bill_cycle_logs.rule"),
		GraylogNginxProcessingRule:        path.Join(GraylogConfig, "processing_rules/nginx_logs.rule"),
		GraylogRemoveKubernetesRule:       path.Join(GraylogConfig, "processing_rules/remove_kubernetes.rule"),
		GraylogRemoveKubernetesLabelsRule: path.Join(GraylogConfig, "processing_rules/remove_kubernetes_labels.rule"),
		GraylogUnsupportedSymbolsRule:     path.Join(GraylogConfig, "processing_rules/unsupported_symbols.rule"),
		GraylogKubernetesEventsRule:       path.Join(GraylogConfig, "processing_rules/k8s_event_logs.rule"),
	}
	GraylogRuleDescriptions = map[string]string{
		GraylogAuditProcessingRule:        "Route Audit logs to the appropriate stream",
		GraylogSystemLogsProcessingRule:   "Route System logs to the appropriate stream",
		GraylogAccessProcessingRule:       "Route Access logs to the appropriate stream",
		GraylogIntegrationProcessingRule:  "Route Integration logs to the appropriate stream",
		GraylogBillCycleProcessingRule:    "Route Bill Cycle logs to the appropriate stream",
		GraylogNginxProcessingRule:        "Route Nginx logs to the appropriate stream",
		GraylogRemoveKubernetesRule:       "Remove kubernetes field",
		GraylogRemoveKubernetesLabelsRule: "Remove kubernetes labels field",
		GraylogUnsupportedSymbolsRule:     "Processing unsupported symbols (Replace '/' to '_')",
		GraylogKubernetesEventsRule:       "Route Kubernetes events to the appropriate stream",
	}
	GraylogDashboard                = path.Join(GraylogConfig, "dashboard.json")
	GraylogDashboardInstallation    = path.Join(GraylogConfig, "dashboardInstallation.json")
	GraylogPipeline                 = path.Join(GraylogConfig, "pipeline.json")
	GraylogCloudEventsSearch        = path.Join(GraylogConfig, "saved_searches/cloud-events-search.json")
	GraylogUserSessionHistorySearch = path.Join(GraylogConfig, "saved_searches/user-session-history-search.json")
	GraylogCloudEventsView          = path.Join(GraylogConfig, "saved_searches/cloud-events-view.json")
	GraylogUserSessionHistoryView   = path.Join(GraylogConfig, "saved_searches/user-session-history-view.json")
	GraylogOperatorRole             = path.Join(GraylogConfig, "roles/operator.json")
	GraylogAuditViewerRole          = path.Join(GraylogConfig, "roles/auditViewer.json")
	GraylogOperatorUser             = path.Join(GraylogConfig, "user_accounts/operator.json")
	GraylogAuditViewerUser          = path.Join(GraylogConfig, "user_accounts/auditViewer.json")
	GraylogAdminWithTrustedHeader   = path.Join(GraylogConfig, "user_accounts/admin_with_trusted_header.json")
	GraylogStartupTimeout           = time.Minute * 10
	GraylogMongoUpgradeJobTimeout   = time.Minute * 2
	GraylogLabels                   = map[string]string{"name": "graylog"}
	GraylogSecretSelector           = "graylog=secret"
	GraylogConfigFileName           = "graylog.conf"
	GraylogPasswordField            = "root_password_sha2"
	GraylogUserField                = "root_username"
	GraylogMongoUpgradeOrderedJobs  = []string{
		"mongo-upgrade-job-40",
		"mongo-upgrade-job-42",
		"mongo-upgrade-job-44",
		"mongo-upgrade-job-50",
	}
	GraylogMongoUpgradeAssets = map[string]string{
		"mongo-upgrade-job-40": path.Join(BasePath, "mongo-upgrade-job-40.yaml"),
		"mongo-upgrade-job-42": path.Join(BasePath, "mongo-upgrade-job-42.yaml"),
		"mongo-upgrade-job-44": path.Join(BasePath, "mongo-upgrade-job-44.yaml"),
		"mongo-upgrade-job-50": path.Join(BasePath, "mongo-upgrade-job-50.yaml"),
	}
	GraylogMongoUpgradeLabels = map[string]string{"name": "mongo-upgrade-job"}

	ComponentPendingStatus            = "ComponentPendingStatus"
	ComponentPendingTimeout           = time.Minute * 5
	FluentbitAggregatorPendingTimeout = time.Minute * 8

	ConnectionTimeout = 10

	InitialDelay = time.Second * 5
)
