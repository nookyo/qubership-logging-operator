*** Variables ***
${GRAYLOG_PROTOCOL}             %{GRAYLOG_PROTOCOL}
${GRAYLOG_HOST}                 %{GRAYLOG_HOST}
${OPENSHIFT_DEPLOY}             %{OPENSHIFT_DEPLOY}
${GRAYLOG_PORT}                 %{GRAYLOG_PORT}
${GRAYLOG_USER}                 %{GRAYLOG_USER}
${GRAYLOG_PASS}                 %{GRAYLOG_PASS}
${OPERATION_RETRY_COUNT}        60x
${RETRY_COUNT_FOR_FIRST_TEST}   250x
${OPERATION_RETRY_INTERVAL}     5s
${FILES_PATH}                   ./source_files/log_generator
${DEPLOYMENT_FILE}              ${FILES_PATH}/deployment.yaml
${CONFIG_FILE}                  ${FILES_PATH}/config.yaml
${DATE_TIME_REGEXP}             [0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}\:[0-9]{2}:[0-9]{2}
${TAG_PREFIX}                   parsed.kubernetes.var.log.pods.${NAMESPASE}_
${FATAL_LEVEL}                  1
${ERROR_LEVEL}                  3
${WARN_LEVEL}                   4
${INFO_LEVEL}                   6
${DEBUG_LEVEL}                  7

*** Settings ***
Library  OperatingSystem
Suite Setup  Run Keywords  Setup
...  AND  Create Config
...  AND  Create Log Generator
Suite Teardown  Run Keywords  Delete Log Generator
...  AND  Delete Config Map
Resource        keywords.robot

*** Keywords ***
Setup
    ${headers}  Create Dictionary  Content-Type=application/json  Accept=application/json
    Set Global Variable  ${headers}
    ${auth}=  Create List  ${GRAYLOG_USER}  ${GRAYLOG_PASS}
    Create Session  graylog  ${GRAYLOG_PROTOCOL}://${GRAYLOG_HOST}:${GRAYLOG_PORT}  auth=${auth}  disable_warnings=1  verify=False  timeout=10
    Check Fluentbit And Fluentd
    IF  ${fluentd_exists} == True
        ${TAG_PREFIX}=  Set Variable  parsed.kubernetes.var.log.pods.${NAMESPASE}_
    ELSE
        ${TAG_PREFIX}=  Set Variable  var.log.pods.${NAMESPASE}_
    END
    Set Suite Variable  ${TAG_PREFIX}

Get Log Generator Name
    ${generator_names}=  Get Pod Names For Deployment Entity  qubership-log-generator  ${NAMESPASE}
    ${generator_pod_name}=  Set Variable  ${generator_names}[0]
    Set Suite Variable  ${generator_pod_name}
    Log to console  log generator pod name: ${generator_pod_name}

Create Log Generator
    IF  "${OPENSHIFT_DEPLOY}" == "true"
        Create Deployment Entity From File  ${DEPLOYMENT_FILE}  ${NAMESPASE}
    ELSE
        ${new_deployment}=  Add Security Context To Deployment  ${DEPLOYMENT_FILE}  ${NAMESPASE}
        Create Deployment Entity  ${new_deployment}  ${NAMESPASE}
    END
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Get Log Generator Name

Create Config
    Create Config Map From File  ${NAMESPASE}  ${CONFIG_FILE}

Delete Log Generator
    Delete Deployment Entity  qubership-log-generator  ${NAMESPASE}

Delete Config Map
    Delete Config Map By Name  qubership-log-generator-config  ${NAMESPASE}

Search messages by query
    [Arguments]  ${query}
    ${resp}=  GET On Session  graylog  url=/api/search/universal/relative?query=pod:${query}&range=3600&limit=50&sort=timestamp:desc&pretty=true  headers=${headers}
    ${messages}=  Get From Dictionary  ${resp.json()}  messages
    Set Suite Variable  ${messages}
    Should Not Be Empty  ${messages}

Check Message Parsing
    [Arguments]  ${query}  ${log_type}  ${expected_level}
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Search messages by query  ${query}
    ${message}=  Get From Dictionary  ${messages}[0]  message
    ${level}=   Get From Dictionary  ${message}  level
    Should Be Equal As Strings  ${level}  ${expected_level}
    ${message_field}=  Get From Dictionary  ${message}  message
    Set Suite Variable  ${message_field}
    Should Contain  ${message_field}  ${log_type}
    ${tag}=  Get From Dictionary  ${message}  tag
    Should Contain  ${tag}  ${TAG_PREFIX}${generator_pod_name}
    IF  ${fluentd_exists} == True
        ${time}=  Get From Dictionary  ${message}  time
    ELSE
        ${time}=  Get From Dictionary  ${message}  timestamp
    END
    Should Match Regexp  ${time}  ${DATE_TIME_REGEXP}

*** Test Cases ***
Test Create Log Generator And Check Messages Exist
    [Tags]  log-generator
    Wait Until Keyword Succeeds  ${RETRY_COUNT_FOR_FIRST_TEST}  ${OPERATION_RETRY_INTERVAL}
    ...  Search messages by query  "${generator_pod_name}"

Check Parsing Go Info Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_info_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${INFO_LEVEL}

Check Parsing Go Warning Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_warn_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${WARN_LEVEL}

Check Parsing Go Error Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_error_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${ERROR_LEVEL}

Check Parsing Go Debug Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_debug_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${DEBUG_LEVEL}

Check Parsing Go Fatal Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_fatal_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${FATAL_LEVEL}

Check Parsing Go Multiline Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  go_multiline_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}_parent"+AND+message%3A+"${log_type}_child"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${ERROR_LEVEL}
    Should Contain  ${message_field}  ${log_type}_child

Check Parsing Java Info Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  java_info_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${INFO_LEVEL}

Check Parsing Java Warning Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  java_warn_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${WARN_LEVEL}

Check Parsing Java Error Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  java_error_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${ERROR_LEVEL}

Check Parsing Java Multiline Logs
    [Tags]  log-generator
    ${log_type}=  Set Variable  java_multiline_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}_parent"+AND+message%3A+"${log_type}_child"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  ${ERROR_LEVEL}
    Should Contain  ${message_field}  ${log_type}_child

Check Parsing Json Info Logs
    [Tags]  log-generator
    Log To Console  ${\n}Config for json log does not match format from documentation. Level is not parsed. Default level = 6
    ${log_type}=  Set Variable  json_log
    ${query}=  Set Variable  "${generator_pod_name}"+AND+message%3A+"${log_type}"+NOT+message%3A+"templates"
    Check Message Parsing  ${query}  ${log_type}  6
