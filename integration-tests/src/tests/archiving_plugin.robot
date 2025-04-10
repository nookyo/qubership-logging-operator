*** Settings ***
Library         SSHLibrary
Library         RequestsLibrary
Library         DateTime
Library	        Collections
Library         OperatingSystem
Library         lib/LoggingLibrary.py
Suite Setup     Setup

*** Variables ***
${GRAYLOG_PROTOCOL}             %{GRAYLOG_PROTOCOL}
${GRAYLOG_HOST}                 %{GRAYLOG_HOST}
${GRAYLOG_PORT}                 %{GRAYLOG_PORT}
${GRAYLOG_USER}                 %{GRAYLOG_USER}
${GRAYLOG_PASS}                 %{GRAYLOG_PASS}
${OPERATION_RETRY_COUNT}        30x
${OPERATION_RETRY_INTERVAL}     5s
${SSH_KEY}                      %{SSH_KEY}
${VM_USER}                      %{VM_USER}

*** Keywords ***
Setup
    ${headers}  Create Dictionary  X-Requested-By  Graylog Api Browser
    Set Global Variable  ${headers}
    ${auth}=  Create List  ${GRAYLOG_USER}  ${GRAYLOG_PASS}
    Create Session  graylog_session  ${GRAYLOG_PROTOCOL}://${GRAYLOG_HOST}  auth=${auth}  disable_warnings=1  verify=False  timeout=10
    Configure File On VM

Configure File On VM
    ${ssh_client}=  Create SSH Connection  ${GRAYLOG_HOST}  ${VM_USER}  ${SSH_KEY}
    ${result}=  Execute Command On VM  ${ssh_client}  sudo bash -c 'echo {"graylog": "/usr/share/elasticsearch/snapshots/graylog/graylog","gray_audit": "/usr/share/elasticsearch/snapshots/graylog/gray_audit"} > /srv/docker/graylog/graylog/config/directories.json'
    ${status_code}=  Get From List  ${result}  0
    Should Be Equal As Strings  ${status_code}  0
    ${result}=  Execute Command On VM  ${ssh_client}  sudo bash -c 'cat /srv/docker/graylog/graylog/config/directories.json'
    ${status_code}=  Get From List  ${result}  0
    Should Be Equal As Strings  ${status_code}  0
    ${text}=  Get From List  ${result}  1
    Should Contain  str(${text})  {graylog: /usr/share/elasticsearch/snapshots/graylog/graylog,gray_audit: /usr/share/elasticsearch/snapshots/graylog/gray_audit}

Get Last Index
    [Arguments]  ${stream}
    ${resp}=  GET On Session  graylog_session  url=/api/system/indexer/indices/open  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${indices}=  Get From Dictionary  ${resp.json()}  indices
    FOR  ${index}  IN  @{indices}
        ${contains}=  Evaluate   "${stream}_" in """${index}}"""
        IF  ${contains}
            ${index_name}=  Get From Dictionary  ${index}  index_name
            Return From Keyword  ${index}
        END
    END

Register FS Directory
    [Arguments]  ${stream}
    ${json}=  Create Dictionary  storageId=${stream}
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/settings/fs  headers=${headers}  json=${json}
    Should Be Equal As Strings  ${resp.status_code}  200
    Should Contain  str(${resp.content})  {"acknowledged":true}

Reload Config File
    [Arguments]  ${stream}
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/settings/reload  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    Should Contain  str(${resp.content})  "${stream}":"/usr/share/opensearch/snapshots/archives/${stream}"

Create Unique Name
    [Arguments]  ${prefix}
    ${date}=  Get Current Date  UTC  result_format=%Y%m%d%H%M%S
    ${archive_name}=  Set Variable  ${prefix}_${date}
    [Return]  ${archive_name}

Check Process Status
    [Arguments]  ${process}
    ${resp}=  GET On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/process/${process}  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    Should Contain  str(${resp.content})  "status":"Success"

Create Archive
    [Arguments]  ${json}
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/archive  headers=${headers}  json=${json}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${process}=  Set Variable  ${resp.content}
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Process Status  ${process}
    ${archive_name}=  Get From Dictionary  ${json}  name
    ${resp}=  GET On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/archive/${archive_name}  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    Should Contain  str(${resp.content})  "state":"SUCCESS"

Restore Archive
    [Arguments]  ${json}
    ${indices_count_before_restore}=  Get Indices Count For Stream  restored
    ${archive_name}=  Get From Dictionary  ${json}  name
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/restore/${archive_name}  headers=${headers}  json=${json}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${process}=  Set Variable  ${resp.content}
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Process Status  ${process}
    ${indices_count_after_restore}=  Get Indices Count For Stream  restored
    Should Be True  ${indices_count_before_restore} < ${indices_count_after_restore}

Delete Archive
    [Arguments]  ${archive_name}
    ${resp}=  DELETE On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/graylog/${archive_name}  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${process}=  Set Variable  ${resp.content}
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Check Process Status  ${process}

Get Message Count For Index
    [Arguments]  ${index}
    ${primary_shards}=  Get From Dictionary  ${index}  primary_shards
    ${documents}=  Get From Dictionary  ${primary_shards}  documents
    ${documents_count}=  Get From Dictionary  ${documents}  count
    Return From Keyword  ${documents_count}

Get Indices Count For Stream
    [Arguments]  ${stream}
    ${resp}=  GET On Session  graylog_session  url=/api/system/indexer/indices/open  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${indices}=  Get From Dictionary  ${resp.json()}  indices
    ${indices_count}=    Set Variable    0
    FOR  ${index}  IN  @{indices}
        ${contains}=  Evaluate   "${stream}_" in """${index}}"""
        IF  ${contains}
            ${indices_count}=   Evaluate    ${indices_count} + 1
        END
    END
    [Return]  ${indices_count}

Create Schedule Job
    [Arguments]  ${job_name}
    ${list}=  Create List  graylog
    ${json}=  Create Dictionary  name=${job_name}  prefixes=${list}  time=2m  period=0 */1 * ? * *  storageId=graylog
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/schedule  headers=${headers}  json=${json}
    Should Be Equal As Strings  ${resp.status_code}  200

Unschedule Job
    [Arguments]  ${job_name}
    ${resp}=  POST On Session  graylog_session  url=/api/plugins/org.qubership.graylog2.plugin/archiving/unschedule/${job_name}  headers=${headers}
    Should Be Equal As Strings  ${resp.status_code}  200

*** Test Cases ***
Create And Restore Archive By Index
    [Tags]  archiving-plugin
    Register FS Directory  graylog
    Reload Config File  graylog
    ${archive_name}=  Create Unique Name  test_archive
    ${graylog_index}=  Get Last Index  graylog
    Should Not Be Equal As Strings  ${graylog_index}  None
    ${graylog_index_name}=  Get From Dictionary  ${graylog_index}  index_name
    ${list}=  Create List  ${graylog_index_name}
    ${json}=  Create Dictionary  name=${archive_name}  indices=${list}  storageId=graylog
    Create Archive  ${json}
    Restore Archive  ${json}
    Delete Archive  ${archive_name}

Create And Restore Archive By Prefix
    [Tags]  archiving-plugin
    Register FS Directory  gray_audit
    Reload Config File  gray_audit
    ${archive_name}=  Create Unique Name  test_archive
    ${list}=  Create List  gray_audit
    ${json}=  Create Dictionary  name=${archive_name}  prefixes=${list}  storageId=gray_audit
    Create Archive  ${json}
    Restore Archive  ${json}
    Delete Archive  ${archive_name}

Test Schedule Job
    [Tags]  archiving-plugin
    ${job_name}=  Create Unique Name  schedule_job
    Create Schedule Job  ${job_name}
    [Teardown]  Unschedule Job  ${job_name}
