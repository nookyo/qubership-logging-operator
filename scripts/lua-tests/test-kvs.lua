-- different test strings
local test_strings = {
    -- original log from cassandra
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow]Compacting (1ac36160-6b81-11ef-b0be-0b820667842e) [/var/lib/cassandra/data/system/compaction_history-b4dbb7b4dc493fb5b3bfce6e434832ca/nb-2406-big-Data.db:level=0, /var/lib/cassandra/data/system/compaction_history-b4dbb7b4dc493fb5b3bfce6e434832ca/nb-2405-big-Data.db:level=0, /var/lib/cassandra/data/system/compaction_history-b4dbb7b4dc493fb5b3bfce6e434832ca/nb-2408-big-Data.db:level=0, /var/lib/cassandra/data/system/compaction_history-b4dbb7b4dc493fb5b3bfce6e434832ca/nb-2407-big-Data.db:level=0, ]",
    -- logs without message after key=value pairs
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][logId=cassandra-compacting]",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value with spaces]",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value with spaces]           ",
    -- logs with spaces between []
    "[2024-09-05T12:19:31,575]  [INFO]      [method=runMayThrow]             [key=value] test message [key_test=value test]",
    "[2024-09-05T12:19:31,575]        [INFO]          [method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05 12:19:31,575]  [INFO]      [method=runMayThrow]             [key=value] test message [key_test=value test]",
    "[2024-09-05 12:19:31,575]        [INFO]          [method=runMayThrow][key=value] test message [key_test=value test]",
    -- logs with message
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow]13215125328131",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow]:)(&^&^%$%%# test 123143",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow] [key1=value1] Compacting bla-bla",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow] Compacting bla-bla",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key1=value1][key2=value2][key3=value3][key4=value4][key5=value5][key6=value6][key7=value7] Compacting (1ac36160-6b81-11ef-b0be-0b820667842e) [key_test=value_test]",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key with spaces=value] Compacting bla-bla",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value with spaces] Comaption (1ac36160-6b81-11ef-b0be-0b820667842e)",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value with spaces] Comaption (1ac36160-6b81-11ef-b0be-0b820667842e) [key_test=value test]",
    -- logs with special symbols in key=value
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key!@#$%1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key^&*_1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key()1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key{}1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value!@#$%1] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=valuey^&*_1] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value()1] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value{}1] test message",
    -- logs with not ISO8601 timestamp
    "[2024-09-05][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05T12:19][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05T12:19:31][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05 12:19:31,575][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024/09/05 12:19:31,575][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[12:19:31,575][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[12:19:31 2024-09-05][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[12:19:31 09/05][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[1725625171][INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    -- logs with prohibited symbols in key=value
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key[1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key]1=value] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value[] test message",
    "[2024-09-05T12:19:31,575][INFO][method=runMayThrow][key=value]] test message",
    -- log with escaped key=value
    "[2024-07-20T11:51:32.234] [\"app.kubernetes.io/part-of=cm\"] [logId=bill-cycle-2] it is example of bill cycle log message",
    "[2024-07-20T11:51:32.234] ['app.kubernetes.io/part-of=cm'] [logId=bill-cycle-2] it is example of bill cycle log message",
    "[2024-07-20T11:51:32.234] [INFO ] [\"app.kubernetes.io/part-of=cm\"] [logId=bill-cycle-2] it is example of bill cycle log message",
    "[2024-07-20T11:51:32.234] [DEUBG] ['app.kubernetes.io/part-of=cm'] [logId=bill-cycle-2] it is example of bill cycle log message",
    -- logs with incorrect format
    "2024-09-05 12:19:31,575 [INFO][method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05 12:19:31,575] INFO [method=runMayThrow][key=value] test message [key_test=value test]",
    "2024-09-05 12:19:31,575 - INFO - [method=runMayThrow][key=value] test message [key_test=value test]",
    "[2024-09-05 12:19:31,575][INFO] key1=value1 key2=value2 test message [key_test=value test]",
    -- kvs like data in message
    "[2024-10-17T03:01:00.108][ERROR] [request_id=1729134060001.0.731003548650101] [tenant_id=a789eed9-217d-4b2b-8f62-09f7461b3aa0] [thread=       sds-7035] [class=o.qub.clo.crm.sds.cli.asy.BaseAsyncJobLauncher              ] [method=lambda$startJob$0             ] [version=v1] [traceId=67107dec9109eb2a9e243f8d5d2ef86d] [spanId=4b3ca3eef36c69a2] [originating_bi_id=                ] [business_identifiers=                ] Exception during job execution class org.qubership.cloud.crm.sds.client.quarkus.async.QuarkusAsyncJobLauncher: org.qubership.cloud.bss.errorhandling.exception.ErrorResponseException: httpStatus=404, errorResponse=ErrorResponse@1c4ce05{errors=[ErrorEntry@17ff7140{status=\"404\",code=\"-7000\",message=\"Endpoint not found.\",source=ErrorSource@17a8bc21{pointer=\"cloud-integration-platform-engine\"},extra={},debugDetail=\"{",
    "[2024-10-16T18:00:01.114] [DEBUG] [request_id=fe8dfb92-1118-4fd1-a1b0-19deadb2dbee] [tenant_id=-] [thread=main-9832a] [class=mongo:storage.go:236] [traceId=000000000000000039f6dc419d14fecc] [spanId=239edd5107afa67a] try to delete objects from certificates by filter map[$and:[map[meta.status:map[$ne:trusted]] map[$or:[map[meta.deactivatedAt:map[$lte:2024-09-16 18:00:01.114591759 +0000 UTC m=+7628575.165500196]] map[details.validTo:map[$lte:2024-09-16 18:00:01.114591759 +0000 UTC m=+7628575.165500196]]]]]]",
    [[
[2024-10-17T06:47:36.834] [INFO] [request_id=f76c8bbb-f8ee-4591-ba29-f34dee5d41cd] [tenant_id=] [thread=executor-thread-48266] [class=org.qubership.cloud.keycloak.handler.dump.DumpRequestHandler] 
duration=92ms
----------------------------REQUEST---------------------------
requestURI=/auth/realms/cloud-common/protocol/openid-connect/token
method=POST
contentEncoding=null
contentType=[application/x-www-form-urlencoded;charset=UTF-8]
contentLength=175"
}
]],
    "[2024-10-17 06:59:47,457] [INFO] [request_id=-] [tenant_id=-] [thread=kafka-admin-client-thread | a789eed9-217d-4b2b-8f62-09f7461b3aa0.env-1-data-management.cpm_ds.cpm_nrml_transformer.streaming-admin] [class=org.apache.kafka.clients.NetworkClient]  [traceId=-] [spanId=-] - [AdminClient clientId=a789eed9-217d-4b2b-8f62-09f7461b3aa0.env-1-data-management.cpm_ds.cpm_nrml_transformer.streaming-admin] Node 3 disconnected.",
    "[2024-10-17T00:00:00.112] [INFO ] [request_id=1729123199250.0.775342141558708] [tenant_id=a789eed9-217d-4b2b-8f62-09f7461b3aa0] [thread=or-thread-92333] [class=o.qub.clo.cal.qua.v2.cac.ser.imp.CacheServiceImpl           ] [method=loadCalendarsCacheForTenantId ] [traceId=                ] [spanId=                ] [originating_bi_id=] [business_identifiers=] [error_code=] [error_id=] - [Business Calendar Client] allCalendarsHierarchy [BusinessCalendarDtoV2(name=Default OOB Business Calendar, timeZone=UTC +00:00, businessArea=null, childBusinessCalendars=null, country=Iceland, region=null, isDefaultCalendar=true)] ...",
    [[
[2024-10-17T07:06:01.480] [INFO ] [request_id=9cb590e3-9e3c-41fe-8e58-ce49ff46af9a] [deployment_session_id=bf8dd1b2-dbbd-4e09-943a-63e2f8de9bdc] [tenant_id=] [thread=ReconcilerExecutor-CompositeReconciler-50] [class=CompositeReconciler] [phase=BackingOff] [name=composite-structure] [kind=Composite] [subKind=CompositeStructure] Reconcile composite Resource CustomResource{kind='Composite', apiVersion='core.qubership.org/v1', metadata=ObjectMeta(annotations={kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"core.qubership.org/v1","kind":"Composite","metadata":{"annotations":{},"labels":{"app.kubernetes.io/instance":"core-operator","app.kubernetes.io/managed-by":"saasDeployer","app.kubernetes.io/part-of":"Cloud-Core","app.kubernetes.io/processed-by-operator":"core-operator","deployer.cleanup/allow":"true","deployment.qubership.org/sessionId":"bf8dd1b2-dbbd-4e09-943a-63e2f8de9bdc"},"name":"composite-structure","namespace":"env-1-data-management"},"spec":{"baseline":{"originNamespace":"env-1-core"},"originNamespace":"env-1-data-management"},"subKind":"CompositeStructure"}
}, creationTimestamp=2024-07-11T06:36:22Z, deletionGracePeriodSeconds=null, deletionTimestamp=null, finalizers=[], generateName=null, generation=1, labels={app.kubernetes.io/instance=core-operator, app.kubernetes.io/managed-by=saasDeployer, app.kubernetes.io/part-of=Cloud-Core, app.kubernetes.io/processed-by-operator=core-operator, deployer.cleanup/allow=true, deployment.qubership.org/sessionId=bf8dd1b2-dbbd-4e09-943a-63e2f8de9bdc}, managedFields=[ManagedFieldsEntry(apiVersion=core.qubership.org/v1, fieldsType=FieldsV1, fieldsV1=FieldsV1(additionalProperties={f:metadata={f:annotations={.={}, f:kubectl.kubernetes.io/last-applied-configuration={}}, f:labels={.={}, f:app.kubernetes.io/instance={}, f:app.kubernetes.io/managed-by={}, f:app.kubernetes.io/part-of={}, f:app.kubernetes.io/processed-by-operator={}, f:deployer.cleanup/allow={}, f:deployment.qubership.org/sessionId={}}}, f:spec={.={}, f:baseline={.={}, f:originNamespace={}}, f:originNamespace={}}, f:subKind={}}), manager=kubectl-client-side-apply, operation=Update, subresource=null, time=2024-07-11T06:36:22Z, additionalProperties={}), ManagedFieldsEntry(apiVersion=core.qubership.org/v1, fieldsType=FieldsV1, fieldsV1=FieldsV1(additionalProperties={f:status={.={}, f:conditions={}, f:observedGeneration={}, f:phase={}, f:requestId={}}}), manager=fabric8-kubernetes-client, operation=Update, subresource=status, time=2024-10-17T07:06:00Z, additionalProperties={})], name=composite-structure, namespace=env-1-data-management, ownerReferences=[], resourceVersion=488503008, selfLink=null, uid=6e9f3ed4-6bdb-4a89-a23e-c54d113a76d3, additionalProperties={}), spec=RawExtension(super=AnyType(value={baseline={originNamespace=env-1-core}, originNamespace=env-1-data-management})), status=org.qubership.core.declarative.resources.base.DeclarativeStatus@246e759b}
]],
    "[2024-10-17T07:10:00.002][INFO ] [request_id=1729149000002.0.7622903492994286] [tenant_id=ce771702-5119-40dd-9175-c5de1b216f50] [thread=eduler_Worker-4] [class=c.n.c.crm.sds.service.execution.start.JobExecutionService   ] [method=execute                       ] [version=v1] [traceId=-               ] [spanId=-               ] [originating_bi_id=-               ] [business_identifiers=-               ] Try to execute job QuartzIntegrationJobData(jobId=shsRetryFlowJob, microserviceName=submission-handler-service, microserviceVersion=0, jobMetadataId=1f5192c4-b695-4f3a-a364-6ea10ed99522), trigger QuartIntegrationCronTrigger@1170fbdb{id=21b9f752-9d24-4fd3-8fdd-44667dac97a0,lockQualifier=4315,scheduleExpression=\"0 0/5 * * * ?\",jobParameters=JobParameters@64b3826b{parameters={\"JOB_FEATURE_LIST\"=[{name=ProhibitParallelRun, parameters={enabled=true}}]}}}, tenant ce771702-5119-40dd-9175-c5de1b216f50",
    [[
[2024-10-17T04:51:50.519] [ERROR] [request_id=899868a7-3576-4e59-8d8a-3400d619bea9] [deployment_session_id=] [tenant_id=] [thread=ReconcilerExecutor-MeshReconciler-58] [class=MeshReconciler] Unexpected error happened when processing entity=CustomResource{kind='Mesh', apiVersion='core.qubership.org/v1', metadata=ObjectMeta(annotations={kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"core.qubership.org/v1","kind":"Mesh","metadata":{"annotations":{},"labels":{"app.kubernetes.io/instance":"frontend-discovery-service","app.kubernetes.io/managed-by":"operator","app.kubernetes.io/managed-by-operator":"core-operator","app.kubernetes.io/part-of":"Cloud-Frontend-Platform","deployer.cleanup/allow":"true"},"name":"frontend-discovery-service-mesh-routes","namespace":"env-1-oss"},"spec":{"gateways":["frontend-discovery-composite-gateway"],"virtualServices":[{"hosts":["frontend-discovery-service"],"name":"frontend-discovery-service","routeConfiguration":{"routes":[{"destination":{"cluster":"frontend-discovery-service","endpoint":"http://frontend-discovery-service:8080"},"rules":[{"allowed":false,"match":{"prefix":"/actuator"}},{"allowed":true,"match":{"prefix":"/"}}]}],"version":""}}]},"subKind":"RouteConfiguration"}
}, creationTimestamp=2024-06-20T10:50:14Z, deletionGracePeriodSeconds=null, deletionTimestamp=null, finalizers=[], generateName=null, generation=1, labels={app.kubernetes.io/instance=frontend-discovery-service, app.kubernetes.io/managed-by=operator, app.kubernetes.io/managed-by-operator=core-operator, app.kubernetes.io/part-of=Cloud-Frontend-Platform, deployer.cleanup/allow=true}, managedFields=[ManagedFieldsEntry(apiVersion=core.qubership.org/v1, fieldsType=FieldsV1, fieldsV1=FieldsV1(additionalProperties={f:metadata={f:annotations={.={}, f:kubectl.kubernetes.io/last-applied-configuration={}}, f:labels={.={}, f:app.kubernetes.io/instance={}, f:app.kubernetes.io/managed-by={}, f:app.kubernetes.io/managed-by-operator={}, f:app.kubernetes.io/part-of={}, f:deployer.cleanup/allow={}}}, f:spec={.={}, f:gateways={}, f:virtualServices={}}, f:subKind={}}), manager=kubectl-client-side-apply, operation=Update, subresource=null, time=2024-06-20T10:50:14Z, additionalProperties={}), ManagedFieldsEntry(apiVersion=core.qubership.org/v1, fieldsType=FieldsV1, fieldsV1=FieldsV1(additionalProperties={f:status={.={}, f:conditions={}, f:observedGeneration={}, f:phase={}, f:requestId={}, f:updated={}}}), manager=fabric8-kubernetes-client, operation=Update, subresource=status, time=2024-10-16T18:51:50Z, additionalProperties={})], name=frontend-discovery-service-mesh-routes, namespace=env-1-oss, ownerReferences=[], resourceVersion=485915972, selfLink=null, uid=f1b53a4e-5190-4594-a597-724ff54ecf18, additionalProperties={}), spec=RawExtension(super=AnyType(value={gateways=[frontend-discovery-composite-gateway], virtualServices=[{hosts=[frontend-discovery-service], name=frontend-discovery-service, routeConfiguration={routes=[{destination={cluster=frontend-discovery-service, endpoint=http://frontend-discovery-service:8080}, rules=[{allowed=false, match={prefix=/actuator}}, {allowed=true, match={prefix=/}}]}], version=}}]})), status=org.qubership.core.declarative.resources.base.DeclarativeStatus@25fe817b}: jakarta.ws.rs.WebApplicationException: Unknown error, status code 500
	at org.jboss.resteasy.microprofile.client.DefaultResponseExceptionMapper.toThrowable(DefaultResponseExceptionMapper.java:42)
	at org.jboss.resteasy.microprofile.client.ExceptionMapping$HandlerException.mapException(ExceptionMapping.java:60)
	at io.quarkus.restclient.runtime.QuarkusProxyInvocationHandler.invoke(QuarkusProxyInvocationHandler.java:172)
	at jdk.proxy2/jdk.proxy2.$Proxy43.applyConfig(Unknown Source)
	at org.qubership.core.declarative.client.reconciler.MeshReconciler.reconcileInternal(MeshReconciler.java:50)
	at org.qubership.core.declarative.client.reconciler.MeshReconciler.reconcileInternal(MeshReconciler.java:30)
	at org.qubership.core.declarative.client.reconciler.CoreReconciler.reconcile(CoreReconciler.java:87)
	at org.qubership.core.declarative.client.reconciler.CoreReconciler.reconcile(CoreReconciler.java:34)
	at org.qubership.core.declarative.client.reconciler.MeshReconciler_ClientProxy.reconcile(Unknown Source)
	at io.javaoperatorsdk.operator.processing.Controller$1.execute(Controller.java:153)
	at io.javaoperatorsdk.operator.processing.Controller$1.execute(Controller.java:111)
	at io.javaoperatorsdk.operator.api.monitoring.Metrics.timeControllerExecution(Metrics.java:219)
	at io.javaoperatorsdk.operator.processing.Controller.reconcile(Controller.java:110)
	at io.javaoperatorsdk.operator.processing.event.ReconciliationDispatcher.reconcileExecution(ReconciliationDispatcher.java:140)
	at io.javaoperatorsdk.operator.processing.event.ReconciliationDispatcher.handleReconcile(ReconciliationDispatcher.java:121)
	at io.javaoperatorsdk.operator.processing.event.ReconciliationDispatcher.handleDispatch(ReconciliationDispatcher.java:91)
	at io.javaoperatorsdk.operator.processing.event.ReconciliationDispatcher.handleExecution(ReconciliationDispatcher.java:64)
	at io.javaoperatorsdk.operator.processing.event.EventProcessor$ReconcilerExecutor.run(EventProcessor.java:417)
	at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1144)
	at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:642)
	at java.base/java.lang.Thread.run(Thread.java:1583)
]],
    "[2024-10-17T07:18:38.492] [INFO ] [request_id=-] [tenant_id=-] [thread=XNIO-1 task-8] [class=c.n.c.d.c.v.AggregatedDatabaseAdministrationControllerV3] New database was created <200 OK OK,DatabaseResponseV3SingleCP{super=DatabaseResponseV3(id=null, classifier={microserviceName=datamarts-airflow, namespace=env-2-data-management, scope=service}, namespace=env-2-data-management, type=postgresql, name=dbaas_30980fcd5902417db9b7b8a26541caf9, externallyManageable=false, timeDbCreation=2024-08-15 19:34:53.11, settings=null, backupDisabled=false, physicalDatabaseId=postgresql-reporting:postgres), connectionProperties=[{password=***, role=admin, port=5432, host=pg-patroni.postgresql-reporting, name=dbaas_30980fcd5902417db9b7b8a26541caf9, url=jdbc:postgresql://pg-patroni.postgresql-reporting:5432/dbaas_30980fcd5902417db9b7b8a26541caf9, username=dbaas_11b25fc8098140a0885fa2892d7f17ea}]},[]>",
    "[2024-10-17T07:22:56.942] [INFO ] [request_id=-] [tenant_id=-] [thread=XNIO-1 task-18] [class=c.n.c.d.s.DBaaService] database for decryption = DatabaseRegistry(id=null, database=Database{id=null, oldClassifier=null, classifier={isServiceDb=true, microserviceName=aap-datahub-base, namespace=env-1-datahub, scope=service}, connectionProperties=[{role=admin, port=5432, host=pg-patroni.postgresql, name=dbaas_19ef1356b5804adb8e61a51687643bc8, url=jdbc:postgresql://pg-patroni.postgresql:5432/dbaas_19ef1356b5804adb8e61a51687643bc8, username=dbaas_f8f3decff5864ab2a7e0e167c33aaa46, encryptedPassword={v2c}{AES}{DEFAULT_KEY}{KtbHzo54Hf3xD49z7TIcSh8LL9bJ+sOtKRvvqrutCTNtR04iaf5U2nnvVtXKIpxC}, password=null}, {role=streaming, port=5432, host=pg-patroni.postgresql, name=dbaas_19ef1356b5804adb8e61a51687643bc8, url=jdbc:postgresql://pg-patroni.postgresql:5432/dbaas_19ef1356b5804adb8e61a51687643bc8, username=dbaas_ac72f68f31684e78a398a7c12ce7022b, encryptedPassword={v2c}{AES}{DEFAULT_KEY}{+y/JoTvEpa9sj6eZF9AbX9laA61x0sozojv2EzNiLGVtR04iaf5U2nnvVtXKIpxC}, password=null}, {role=rw, port=5432, host=pg-patroni.postgresql, name=dbaas_19ef1356b5804adb8e61a51687643bc8, url=jdbc:postgresql://pg-patroni.postgresql:5432/dbaas_19ef1356b5804adb8e61a51687643bc8, username=dbaas_cc433d4e8b4d4c879c3132db6a775331, encryptedPassword={v2c}{AES}{DEFAULT_KEY}{JuHmxdMHj5MPnrhEa23XDMooQ30pNL3wsiXmARk5y+JtR04iaf5U2nnvVtXKIpxC}, password=null}, {role=ro, port=5432, host=pg-patroni.postgresql, name=dbaas_19ef1356b5804adb8e61a51687643bc8, url=jdbc:postgresql://pg-patroni.postgresql:5432/dbaas_19ef1356b5804adb8e61a51687643bc8, username=dbaas_c7e8ff7540f045b2bb6cc10cb368e1d9, encryptedPassword={v2c}{AES}{DEFAULT_KEY}{pg0BvLgasZsA5ZZ/gHB1UTzmDnuLA4Wjb9CUUSVfzaVtR04iaf5U2nnvVtXKIpxC}, password=null}], resources=[DbResource(id=4feac2ed-2db0-4967-b367-24606d1deb70, kind=database, name=dbaas_19ef1356b5804adb8e61a51687643bc8), DbResource(id=e04ab9e2-d86f-4220-ac37-ebc8529a5e82, kind=user, name=dbaas_f8f3decff5864ab2a7e0e167c33aaa46), DbResource(id=1906f863-e635-4108-9e77-848fb3fbcbff, kind=user, name=dbaas_ac72f68f31684e78a398a7c12ce7022b), DbResource(id=6091b182-9274-4cac-9427-06f584eba938, kind=user, name=dbaas_cc433d4e8b4d4c879c3132db6a775331), DbResource(id=0f889d95-8203-4ce9-8f7e-5760a7a8841a, kind=user, name=dbaas_c7e8ff7540f045b2bb6cc10cb368e1d9)], namespace='env-1-datahub', type='postgresql', adapterId='0a0bc11e-0e95-444c-89c6-f796da7130f0', name='dbaas_19ef1356b5804adb8e61a51687643bc8', markedForDrop=false, timeDbCreation=2024-07-11 07:01:23.524, backupDisabled=true, settings=null, connectionDescription=null, warnings=null, externallyManageable=false, dbState=DbState(id=0427cbb0-e08d-42c4-b558-0e2bf6d4ccd9, state=CREATED, databaseState=CREATED, description=null, podName=null), physicalDatabaseId='postgresql:postgres', bgVersion='null'}, timeDbCreation=2024-07-11 07:01:23.524, classifier={isServiceDb=true, microserviceName=aap-datahub-base, namespace=env-1-datahub, scope=service}, namespace=env-1-datahub, type=postgresql)",
}

-- functions like in flientbit

function kv_parse(tag, timestamp, record)
    if record["log"] ~= nil and type(record["log"]) ~= "table" then
        local regex_msg = "%s*(%[[%d%s-:%.,T/]+%])%s*(%[%s*[%w]+%s*%])%s*(%[.+=.*%])%s*[%w%S]+"
        local regex_kvs = "%[([^=%[%]]+)=(%w*(.[^%[^%]]*))%]"
        local s = record["log"]

        time, level, kvs = string.match(s, regex_msg)

        if kvs ~= nil then
            for k, v in string.gmatch(kvs, regex_kvs) do
              record[k] = v
            end
        else
            -- return 0, that means the record will not be modified
            return 0, timestamp, record
        end

        -- return 2, that means the original timestamp is not modified and the record has been modified
        -- so it must be replaced by the returned values from the record 
        return 2, timestamp, record
    else
        -- return 0, that means the record will not be modified
        return 0, timestamp, record
    end
end

function kv_parse_new_gen(tag, timestamp, record)
    if record["log"] ~= nil and type(record["log"]) ~= "table" then
        -- regex to find the end of key=value string in the original string
        -- this regex search the place:
        -- * start from ]
        -- * with 0 or more space symbols
        -- * without [
        -- * start from alphabet symbol, digit or any symbol (expect [)
        local regex_kvs_end = "]%s*[^%[][%w%-%{%}%\\%/%.%,%!%@%#%$%%%^%&%*%(%)]%s*"
        local regex_kvs = "%[([^=%[%]]+)=(%w*(.[^%[^%]]*))%]"
        local s = record["log"]

        local kvs_position = string.find(s, regex_kvs_end, 1)
        local kvs = string.sub(s, 0, kvs_position)

        if kvs ~= nil then
            for k, v in string.gmatch(kvs, regex_kvs) do
              record[k] = v
            end
        else
            -- return 0, that means the record will not be modified
            return 0, timestamp, record
        end

        -- return 2, that means the original timestamp is not modified and the record has been modified
        -- so it must be replaced by the returned values from the record 
        return 2, timestamp, record
    else
        -- return 0, that means the record will not be modified
        return 0, timestamp, record
    end
end

-- test functions
-- call "like real" functions

function execute_real_func_test()
    for i, test_string in ipairs(test_strings) do
        test_structure = {}
        test_structure["log"] = test_string
        print("Original string:", test_string)

        local start_time = os.time()
        print ("Call kv_parse = ", start_time)
        code, time, test_structure = kv_parse("test", i, test_structure)
        local end_time = os.time()
        print ("Complete kv_parse = ", end_time, "Execution time =", end_time - start_time)

        print("Code:", code)
        print("Time:", time)
        for k,v in pairs(test_structure) do
            print("Record content:", k, "=", v)
        end
        print("------------------------------------------------------------------------")
    end
end

function execute_real_func_test_new_gen()
    for i, test_string in ipairs(test_strings) do
        test_structure = {}
        test_structure["log"] = test_string
        print("Original string:", test_string)

        local start_time = os.time()
        print ("Call kv_parse = ", start_time)
        code, time, test_structure = kv_parse_new_gen("test", i, test_structure)
        local end_time = os.time()
        print ("Complete kv_parse = ", end_time, "Execution time =", end_time - start_time)

        print("Code:", code)
        print("Time:", time)
        for k,v in pairs(test_structure) do
            print("Record content:", k, "=", v)
        end
        print("------------------------------------------------------------------------")
    end
end

-- syntetic functions

function execute_test()
    -- regex to parse logs
    local regex_msg = "%s*(%[[%d%s-:%.,T/]+%])%s*(%[%s*[%w]+%s*%])%s*(%[.+=.*%])%s*[%w%d%S]+"
    local regex_kvs = "%[([^=%[%]]+)=(%w*(.[^%[^%]]*))%]"

    -- execute test scenarious 
    for i, test_string in ipairs(test_strings) do
        print("Original string:", test_string)
        time, level, kvs = string.match(test_string, regex_msg)
        print("Raw group parsed:", time, "|", level, "|", kvs)

        for time, level, kvs in string.gmatch (test_string, regex_msg) do
            print("Pasred time:", time)
            print("Parsed level:",level)
            print()
            print("Parsed KVs:", kvs)

            for k, v in string.gmatch (kvs, regex_kvs) do
                print("Parsed key=value:", k, "=", v)
            end
        end
    print("------------------------------------------------------------------------")
    end
end

function execute_test_new_gen()
    -- regex to parse logs
    local regex_msg = "%s*(%[[%d%s-:%.,T/]+%])%s*(%[%s*[%w]+%s*%])%s*(%[.+=.*%])%s*[%w%d%S]+"
    local regex_kvs = "%[([^=%[%]]+)=(%w*(.[^%[^%]]*))%]"

    local kvs_end = "]%s*[^%[][%w%-%{%}%\\%/%.%,%!%@%#%$%%%^%&%*%(%)]%s*"

    -- execute test scenariou
    for i, test_string in ipairs(test_strings) do
        print("Original string:", test_string)

        local kvs_position = string.find(test_string, kvs_end, 1)
        print("Found position of KVs end = ", kvs_position)

        local kvs_substring = string.sub(test_string, 0, kvs_position)
        print("Substring with KVs = ", kvs_substring)
        for k, v in string.gmatch (kvs_substring, regex_kvs) do
            print("Parsed key=value:", k, "=", v)
        end
    print("------------------------------------------------------------------------")
    end
end

--print("====================================================================")
--print("Run test to check regex")
--print("====================================================================")
--execute_test()
--execute_test_new_gen()

print("====================================================================")
print("Run test to check function which will use Fluent")
print("====================================================================")
--execute_real_func_test()
execute_real_func_test_new_gen()