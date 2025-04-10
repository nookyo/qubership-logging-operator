*** Variables ***
${GRAYLOG_PROTOCOL}             %{GRAYLOG_PROTOCOL}
${GRAYLOG_HOST}                 %{GRAYLOG_HOST}
${GRAYLOG_PORT}                 %{GRAYLOG_PORT}
${GRAYLOG_USER}                 %{GRAYLOG_USER}
${GRAYLOG_PASS}                 %{GRAYLOG_PASS}
${OPERATION_RETRY_COUNT}        30x
${OPERATION_RETRY_INTERVAL}     5s

*** Settings ***
Library  String
Suite Setup    Run Keywords  Setup
...  AND  Check Fluentbit And Fluentd
Resource        keywords.robot

*** Keywords ***
Setup
    ${headers}  Create Dictionary  Content-Type=application/json  Accept=application/json
    Set Global Variable  ${headers}
    ${auth}=  Create List  ${GRAYLOG_USER}  ${GRAYLOG_PASS}
    Create Session  graylog  ${GRAYLOG_PROTOCOL}://${GRAYLOG_HOST}:${GRAYLOG_PORT}  auth=${auth}  disable_warnings=1  verify=False  timeout=10

Check Graylog Availability
    ${resp}=  Get On Session  graylog  url=/
    Should Be Equal As Strings  ${resp.status_code}  200

Check Streams Logs
    [Arguments]  ${stream_element}
    ${resp}=  Get On Session  graylog  url=/api/streams
    ${streams}=  Get From Dictionary  ${resp.json()}  streams
    Should Contain  str(${streams})  'title': '${stream_element}'

Get Graylog Version
    ${resp}=  Get On Session  graylog  url=/api/cluster
    @{keys}=  Get Dictionary Keys  ${resp.json()}
    ${key}=  Get From List  ${keys}  0
    ${dict}=  Get From Dictionary  ${resp.json()}  ${key}
    ${version}=  Get From Dictionary  ${dict}  version
    [Return]  ${version}

Search messages
    ${resp}=  GET On Session  graylog  url=/api/search/universal/relative?query=*&range=3600&limit=50&sort=timestamp:desc&pretty=true  headers=${headers}
    ${messages}=  Get From Dictionary  ${resp.json()}  messages
    Should Not Be Empty  ${messages}
    Set Suite Variable  ${messages}

Search messages by query
    [Arguments]  ${query}
    ${resp}=  GET On Session  graylog  url=/api/search/universal/relative?query=${query}&range=3600&limit=50&sort=timestamp:desc&pretty=true  headers=${headers}
    ${messages}=  Get From Dictionary  ${resp.json()}  messages
    Set Suite Variable  ${messages}
    Should Not Be Empty  ${messages}

Check Indexer Cluster Status
    ${resp}=  GET On Session  graylog  url=/api/system/indexer/cluster/health
    ${status}=  Get From Dictionary  ${resp.json()}  status
    Should Be Equal As Strings  ${status}  green

Check Pods Are Ready
    [Arguments]  ${object}  ${ready_name}  ${expected_name}
    ${status}=  Set Variable  ${object.status}
    ${ready}=  Set Variable  ${status.${ready_name}}
    ${expected}=  Set Variable  ${status.${expected_name}}
    Should Be Equal  ${ready}  ${expected}

Get Source List From Messages
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Search messages
    @{source_list}=    Create List
    FOR    ${el}    IN    @{messages}
        ${message}=   Get From Dictionary  ${el}  message
        ${source}=   Get From Dictionary  ${message}  source
        Append To List  ${source_list}  ${source}
    END
    RETURN  @{source_list}

Get Pod Names For Service
    [Arguments]  ${service_name}
    &{dict}=  Create Dictionary  component=${service_name}
    ${pods}=  Get Pod Names By Selector  ${NAMESPASE}  ${dict}
    RETURN  ${pods}

Check Message From Any Pod
    [Arguments]  ${service_name}
    @{pod_list}=  Get Pod Names For Service  ${service_name}
    FOR    ${el}    IN    @{pod_list}
        ${query}=  Set Variable  source%3A+"${el}"
        Search messages by query  ${query}
        ${lenth}=  Get Length    ${messages}
        Exit For Loop If    ${lenth} > 0
    END

Check Messages From All Pods
    [Arguments]  ${service_name}
    @{pod_list}=  Get Pod Names For Service  ${service_name}
    FOR    ${el}    IN    @{pod_list}
        ${query}=  Set Variable  source%3A+"${el}"
        Search messages by query  ${query}
    END

*** Test Cases ***
Test Graylog Availability Check
    [Tags]  smoke  graylog
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Graylog Availability

Check System Logs Stream Exists
    [Tags]  smoke  graylog
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Streams Logs  System logs

Check Audit Logs Stream Exists
    [Tags]  smoke  graylog
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Streams Logs  Audit logs

Check All Messages Stream Exists
    [Tags]  smoke  graylog
    ${version} =  Get Graylog Version
    @{numbers} =  Split String  ${version}  .
    ${first_number}  Convert To Integer  ${numbers[0]}
    IF  ${first_number} >= 5
        ${stream_logs_name}=  Set Variable  Default Stream
    ELSE
        ${stream_logs_name}=  Set Variable  All messages
    END
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Streams Logs  ${stream_logs_name}

Test Indexer Cluster Status
    [Tags]  smoke  graylog
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Indexer Cluster Status

Test Search Message
    [Tags]  smoke  graylog
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Search messages

Check Message From Any Fluentbit Pod
    [Tags]  smoke
    Skip If  ${fluentbit_exists} != True
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Message From Any Pod  logging-fluentbit

Check Message From Any Fluentd Pod
    [Tags]  smoke
    Skip If  ${fluentd_exists} != True
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Message From Any Pod  logging-fluentd

Check Messages From All Fluentbit Pods
    [Tags]  smoke
    Skip If  ${fluentbit_exists} != True
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Messages From All Pods  logging-fluentbit

Check Messages From All Fluentd Pods
    [Tags]  smoke
    Skip If  ${fluentd_exists} != True
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Messages From All Pods  logging-fluentd

Test Check Fluentbit Status
    [Tags]  smoke
    Skip If  ${fluentbit_exists} != True
    ${daemon}=  Get Daemon Set  logging-fluentbit  ${NAMESPASE}
    Check Pods Are Ready  ${daemon}  numberReady  currentNumberScheduled

Test Check Fluentd Status
    [Tags]  smoke
    Skip If  ${fluentd_exists} != True
    ${daemon}=  Get Daemon Set  logging-fluentd  ${NAMESPASE}
    Check Pods Are Ready  ${daemon}  numberReady  currentNumberScheduled

Test Check Events Reader Status
    [Tags]  smoke
    ${deployment}=  Get Deployment Entity  events-reader  ${NAMESPASE}
    Check Pods Are Ready  ${deployment}  ready_replicas  replicas
