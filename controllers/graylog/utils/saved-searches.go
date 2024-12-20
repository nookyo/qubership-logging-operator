package utils

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
)

const (
	savedSearchesDir string = "saved-searches/"
)

func FindById(slice []Entity, val string) (int, bool) {
	for i, item := range slice {
		if item.Id == val {
			return i, true
		}
	}
	return -1, false
}

func (connector *GraylogConnector) GetAllViews() ([]Entity, error) {
	views, err := connector.GetData("views", "views", nil)
	if err != nil {
		return nil, err
	}
	return views, nil
}

func (connector *GraylogConnector) Upload(path string, template string, cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, template), template, cr.ToParams())
	if err != nil {
		return err
	}

	response, statusCode, err := connector.POST(path, data)

	if err != nil {
		return err
	}
	connector.Log.V(util.Debug).Info("Status code: " + strconv.Itoa(statusCode))
	if (statusCode != http.StatusOK) && (statusCode != http.StatusCreated) {
		connector.Log.V(util.Debug).Info("Response: " + response)
		return errors.New("can't upload the content. Path: " + path)
	}

	return nil
}

func (connector *GraylogConnector) UploadCustomSavedSearch(name string, cr *loggingService.LoggingService) error {
	fileContent, err := util.ReadFile(filepath.Join(dataDir, savedSearchesDir, name))
	if err != nil {
		return err
	}
	data, err := util.ParseTemplate(fileContent, filepath.Join(dataDir, savedSearchesDir, name), cr.ToParams())
	if err != nil {
		return err
	}
	if strings.HasSuffix(name, "search.json") {
		_, statusCode, err := connector.POST("views/search", data)
		if err != nil {
			return err
		}
		if statusCode != http.StatusCreated {
			return errors.New("can't upload saved-search " + name)
		}
	} else if strings.HasSuffix(name, "view.json") {
		r, _ := regexp.Compile("(\"id\"\\s?:\\s?\")([a-z,0-9]*)\"")
		viewIds := r.FindStringSubmatch(data)
		if len(viewIds) > 0 {
			views, err := connector.GetAllViews()
			if err != nil {
				return err
			}

			viewId := viewIds[2]
			_, found := FindById(views, viewId)
			if found {
				_, _, err := connector.DELETE("views/" + viewId)
				if err != nil {
					return err
				}
			}
		}

		_, statusCode, err := connector.POST("views", data)
		if err != nil {
			return err
		}
		if (statusCode != http.StatusOK) && (statusCode != http.StatusCreated) {
			return errors.New("can't upload view " + name)
		}
	} else {
		return errors.New("Incorrect filename of saved-search: " + name)
	}
	return nil
}

func (connector *GraylogConnector) UpdateOrCreateSearch(savedSearches []Entity, title string, template string, cr *loggingService.LoggingService) error {
	if err := connector.Upload("views/search", template, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) CreateSearch(savedSearches []Entity, title string, template string, cr *loggingService.LoggingService) error {
	searchId := GetIdByTitle(savedSearches, title)

	if searchId == "" {
		if err := connector.Upload("views/search", template, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) CreateView(savedSearches []Entity, title string, template string, cr *loggingService.LoggingService) error {
	searchId := GetIdByTitle(savedSearches, title)

	if searchId == "" {
		if err := connector.Upload("views/", template, cr); err != nil {
			return err
		}
	}

	return nil
}

func (connector *GraylogConnector) UpdateOrCreateSavedSearches(savedSearches []Entity, cr *loggingService.LoggingService) error {
	if err := connector.UpdateOrCreateSearch(savedSearches, "Cloud events", util.GraylogCloudEventsSearch, cr); err != nil {
		return err
	}

	if err := connector.UpdateOrCreateSearch(savedSearches, "User session history", util.GraylogUserSessionHistorySearch, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) CreateSavedSearches(savedSearches []Entity, cr *loggingService.LoggingService) error {
	if err := connector.CreateSearch(savedSearches, "Cloud events", util.GraylogCloudEventsSearch, cr); err != nil {
		return err
	}

	if err := connector.CreateSearch(savedSearches, "User session history", util.GraylogUserSessionHistorySearch, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) CreateViews(savedSearches []Entity, cr *loggingService.LoggingService) error {
	if err := connector.CreateView(savedSearches, "Cloud events", util.GraylogCloudEventsView, cr); err != nil {
		return err
	}

	if err := connector.CreateView(savedSearches, "User session history", util.GraylogUserSessionHistoryView, cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) CreateOrUpdateSavedSearches(savedSearches []Entity, cr *loggingService.LoggingService) error {
	if cr.Spec.Graylog.IsOnlyCreate() {
		if err := connector.CreateSavedSearches(savedSearches, cr); err != nil {
			return err
		}
	} else if cr.Spec.Graylog.IsForceUpdate() {
		if err := connector.UpdateOrCreateSavedSearches(savedSearches, cr); err != nil {
			return err
		}
	}
	return nil
}

func (connector *GraylogConnector) ManageCustomSavedSearches(cr *loggingService.LoggingService) error {
	if _, err := os.Stat(filepath.Join(dataDir, savedSearchesDir)); err == nil {
		connector.Log.V(util.Debug).Info("Start processing custom saved searches")
		files, err := os.ReadDir(filepath.Join(dataDir, savedSearchesDir))
		if err != nil {
			return err
		}
		for _, f := range files {
			if err = connector.UploadCustomSavedSearch(f.Name(), cr); err != nil {
				return err
			}
		}
	} else if os.IsNotExist(err) {
		connector.Log.V(util.Debug).Info("Directory " + filepath.Join(dataDir, savedSearchesDir) + " not found")
	} else {
		return err
	}
	return nil
}

func (connector *GraylogConnector) ManageSavedSearches(cr *loggingService.LoggingService) error {
	savedSearches, err := connector.GetData("views", "views", nil)
	if err != nil {
		return err
	}

	if err = connector.CreateOrUpdateSavedSearches(savedSearches, cr); err != nil {
		return err
	}

	if err = connector.CreateViews(savedSearches, cr); err != nil {
		return err
	}

	if err = connector.ManageCustomSavedSearches(cr); err != nil {
		return err
	}

	return nil
}
