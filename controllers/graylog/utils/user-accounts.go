package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func (connector *GraylogConnector) GetRole(role string) (int, error) {
	_, code, err := connector.GET("roles/" + role)
	if err != nil {
		return code, err
	}
	return code, err
}

func (connector *GraylogConnector) GetUser(user string) (int, error) {
	_, code, err := connector.GET("users/" + user)
	if err != nil {
		return code, err
	}
	return code, err
}

func (connector *GraylogConnector) GetAllUsers() ([]User, error) {
	users, err := connector.GetRawData("users")

	if err != nil {
		return nil, err
	}

	var data map[string]json.RawMessage

	if err = json.Unmarshal([]byte(users), &data); err != nil {
		return nil, err
	}

	var parsedData []User
	if err = json.Unmarshal(data["users"], &parsedData); err != nil {
		return nil, err
	}
	return parsedData, nil
}

func (connector *GraylogConnector) GetUserIdByName(username string) string {
	users, err := connector.GetAllUsers()

	if err != nil {
		return ""
	}

	for _, v := range users {
		if v.Username == username {
			return v.Id
		}
	}
	return ""
}

func (connector *GraylogConnector) CreateOperatorData(dashboards []Entity, template string) (string, error) {
	streams, err := connector.GetAllStreams()
	if err != nil {
		return "", err
	}

	allEventsStreamId := GetIdByTitle(streams, "All events")
	if allEventsStreamId == "" {
		return "", errors.New("all events stream not found")
	}

	defaultStreamId := GetIdByTitle(streams, util.GraylogDefaultStream)
	if defaultStreamId == "" {
		defaultStreamId = GetIdByTitle(streams, util.GraylogAllMessagesStream)
		if defaultStreamId == "" {
			return "", errors.New("neither " + util.GraylogDefaultStream + " nor " + util.GraylogAllMessagesStream + " streams were not found")
		}
	}

	allSystemEventsStreamId := GetIdByTitle(streams, "All system events")
	if allSystemEventsStreamId == "" {
		return "", errors.New("all system events stream not found")
	}

	systemLogsStreamId := GetIdByTitle(streams, "System logs")
	if systemLogsStreamId == "" {
		return "", errors.New("system logs stream not found")
	}
	sourcesDashboardId := GetIdByTitle(dashboards, "Sources by Service")
	if sourcesDashboardId == "" {
		return "", errors.New("dashboard sources for operator user not found")
	}

	settings := map[string]string{
		"allEventsStreamId":       allEventsStreamId,
		"defaultStreamId":         defaultStreamId,
		"allSystemEventsStreamId": allSystemEventsStreamId,
		"systemLogsStreamId":      systemLogsStreamId,
		"sourcesDashboardId":      sourcesDashboardId,
	}

	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, settings)
	if err != nil {
		return "", err
	}

	return data, nil
}

func (connector *GraylogConnector) CreateAuditViewerData(dashboards []Entity, template string) (string, error) {

	streams, err := connector.GetAllStreams()
	if err != nil {
		return "", err
	}

	allEventsStreamId := GetIdByTitle(streams, "All events")
	if allEventsStreamId == "" {
		return "", errors.New("all events stream not found")
	}

	defaultStreamId := GetIdByTitle(streams, util.GraylogDefaultStream)
	if defaultStreamId == "" {
		defaultStreamId = GetIdByTitle(streams, util.GraylogAllMessagesStream)
		if defaultStreamId == "" {
			return "", errors.New("neither " + util.GraylogDefaultStream + " nor " + util.GraylogAllMessagesStream + " streams were not found")
		}
	}

	allSystemEventsStreamId := GetIdByTitle(streams, "All system events")
	if allSystemEventsStreamId == "" {
		return "", errors.New("all system events stream not found")
	}

	systemLogsStreamId := GetIdByTitle(streams, "System logs")
	if systemLogsStreamId == "" {
		return "", errors.New("system logs stream not found")
	}

	auditLogsStreamId := GetIdByTitle(streams, "Audit logs")
	if auditLogsStreamId == "" {
		return "", errors.New("audit logs stream not found")
	}

	sourcesDashboardId := GetIdByTitle(dashboards, "Sources by Service")
	if sourcesDashboardId == "" {
		return "", errors.New("dashboard sources for audit viewer not found")
	}

	settings := map[string]string{
		"allEventsStreamId":       allEventsStreamId,
		"defaultStreamId":         defaultStreamId,
		"allSystemEventsStreamId": allSystemEventsStreamId,
		"systemLogsStreamId":      systemLogsStreamId,
		"auditLogsStreamId":       auditLogsStreamId,
		"sourcesDashboardId":      sourcesDashboardId,
	}

	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, settings)
	if err != nil {
		return "", err
	}

	return data, nil
}

func (connector *GraylogConnector) CreateRole(data string) error {
	_, statusCode, err := connector.POST("roles", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't create role " + data)
	}
	return nil
}

func (connector *GraylogConnector) CreateUser(template string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, cr.ToParams())
	if err != nil {
		return err
	}

	_, statusCode, err := connector.POST("users", data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't create user " + data)
	}
	return nil
}

func (connector *GraylogConnector) UpdateUser(user string, template string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, cr.ToParams())
	if err != nil {
		return err
	}

	_, statusCode, err := connector.PUT("users/"+connector.GetUserIdByName(user), data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusNoContent {
		return errors.New("can't update user " + data)
	}
	return nil
}

func (connector *GraylogConnector) UpdateRole(role string, data string) error {
	_, statusCode, err := connector.PUT("roles/"+role, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't update role " + role)
	}
	return nil
}

func (connector *GraylogConnector) CreateOrUpdateRole(role string, data string, cr *loggingService.LoggingService) error {
	roleStatus, _ := connector.GetRole(role)

	if roleStatus == http.StatusNotFound {
		if err := connector.CreateRole(data); err != nil {
			return err
		}
	} else if (cr.Spec.Graylog.IsForceUpdate()) && (roleStatus == http.StatusOK) {
		if err := connector.UpdateRole(role, data); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateUser(user string, template string, cr *loggingService.LoggingService) error {
	userStatus, _ := connector.GetUser(user)

	if userStatus == http.StatusNotFound {
		if err := connector.CreateUser(template, cr); err != nil {
			return err
		}
	} else if (cr.Spec.Graylog.IsForceUpdate()) && (userStatus == http.StatusOK) {
		if err := connector.UpdateUser(user, template, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateRoles(dashboards []Entity, cr *loggingService.LoggingService) error {
	data, err := connector.CreateOperatorData(dashboards, util.GraylogOperatorRole)
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateRole("operator", data, cr); err != nil {
		return err
	}

	data, err = connector.CreateAuditViewerData(dashboards, util.GraylogAuditViewerRole)
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateRole("AuditViewer", data, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateUsers(cr *loggingService.LoggingService) error {
	if err := connector.CreateOrUpdateUser("operator", util.GraylogOperatorUser, cr); err != nil {
		return err
	}

	if err := connector.CreateOrUpdateUser("auditViewer", util.GraylogAuditViewerUser, cr); err != nil {
		return err
	}

	if err := connector.CreateOrUpdateUser("graylog_api_th_user", util.GraylogAdminWithTrustedHeader, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) ManageUserAccounts(cr *loggingService.LoggingService, installGraylog5 bool) error {
	// Dashboards API has different names of field with returned dashboards depending on the Graylog version
	// Graylog 4: "views"
	// Graylog 5: "elements"
	var getDashboardDataField string
	if installGraylog5 {
		getDashboardDataField = "elements"
	} else {
		getDashboardDataField = "views"
	}
	dashboards, err := connector.GetData("dashboards", getDashboardDataField, nil)
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateRoles(dashboards, cr); err != nil {
		return err
	}

	if err = connector.CreateOrUpdateUsers(cr); err != nil {
		return err
	}

	return nil
}
