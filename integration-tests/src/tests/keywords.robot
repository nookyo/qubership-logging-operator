*** Variables ***
${NAMESPASE}                    %{LOGGING_PROJECT}

*** Settings ***
Library  lib/LoggingLibrary.py
Library  PlatformLibrary  managed_by_operator=true
Library	 RequestsLibrary
Library	 Collections

*** Keywords ***
Check Fluentbit And Fluentd
    ${fluentbit_exists}=  Check Daemon Set Exists  logging-fluentbit
    Set Suite Variable  ${fluentbit_exists}
    ${fluentd_exists}=  Check Daemon Set Exists  logging-fluentd
    Set Suite Variable  ${fluentd_exists}

Check Daemon Set Exists
    [Arguments]  ${name}
    @{daemon_sets}=  Get Daemon Sets  ${NAMESPASE}
    FOR    ${daemon_set}    IN    @{daemon_sets}
        ${metadata}=  Set Variable  ${daemon_set.metadata}
        ${daemon_set_name}=  Set Variable  ${metadata.name}
        IF  "${name}" == "${daemon_set_name}"  RETURN  True
    END
    RETURN  False
