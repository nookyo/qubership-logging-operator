*** Settings ***
Library  String
Library  Collections
Library  PlatformLibrary  managed_by_operator=true

*** Variables ***
${NAMESPACE}      %{LOGGING_PROJECT}

*** Keywords ***
Compare Images From Resources With Parameters
    [Arguments]  ${param_images}
    ${stripped_resources}=  Strip String  ${param_images}  characters=,  mode=right
    @{list_resources} =  Split String	${stripped_resources} 	,
    FOR  ${resource}  IN  @{list_resources}
      ${type}  ${name}  ${container_name}  ${image}=  Split String	${resource}
      ${resource_image}=  Get Resource Image  ${type}  ${name}  ${NAMESPACE}  ${container_name}
      Should Be Equal  ${resource_image}  ${image}
    END

*** Test Cases ***
Test Hardcoded Images
    [Tags]  logging_images  smoke
    ${param_images}=  Get Dd Images From Config Map  tests-config  ${NAMESPACE}
    Skip If  '${param_images}' == '${None}'  There is no image parameters, not possible to check case!
    Compare Images From Resources With Parameters  ${param_images}

