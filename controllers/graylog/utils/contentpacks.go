package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	util "github.com/Netcracker/qubership-logging-operator/controllers/utils"
	"k8s.io/client-go/kubernetes"
)

const contentPacksDir string = "content-packs"

var dataDir = os.TempDir()

func (connector *GraylogConnector) UploadContentPack(name string, cr *loggingService.LoggingService) error {
	fileContent, err := util.ReadFile(filepath.Join(dataDir, contentPacksDir, name))
	if err != nil {
		return err
	}
	data, err := util.ParseTemplate(fileContent, filepath.Join(dataDir, contentPacksDir, name), cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.POST(contentpacksUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't upload content-pack " + name)
	}
	return nil
}

// replaceFuncContentPackInstallations is needed to unmarshall json response to Entity struct.
// GET response of content pack installations has ids of installation under key `_id`
func replaceFuncContentPackInstallations(data string) string {
	return strings.ReplaceAll(data, "_id", "id")
}

// DeleteDefaultContentPack gets content pack installation, deletes its
// (because content pack can not be deleted if installations exist)
// and deletes content pack
func (connector *GraylogConnector) DeleteDefaultContentPack() error {
	installations, err := connector.GetData(fmt.Sprintf(contentPackInstallationsUrl, oobContentPackId), "installations", replaceFuncContentPackInstallations)
	if err != nil {
		return err
	}
	for _, installation := range installations {
		_, statusCode, err := connector.DELETE(fmt.Sprintf(contentPackInstallationUrl, oobContentPackId, installation.Id))
		if err != nil {
			return err
		}
		if statusCode != http.StatusOK {
			return errors.New("can't delete content pack installation")
		}
	}

	_, statusCode, err := connector.DELETE(contentpacksUrl + "/" + oobContentPackId)
	if err != nil {
		return err
	}
	if statusCode != http.StatusNoContent {
		return errors.New("can't delete content pack")
	}
	return nil
}

func (connector *GraylogConnector) UploadDefaultContentPack(cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogDashboard), util.GraylogDashboard, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.POST(contentpacksUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return errors.New("can't upload content pack")
	}
	return nil
}

func (connector *GraylogConnector) InstallDefaultContentPack(cr *loggingService.LoggingService) error {
	data, err := util.ParseTemplate(util.MustAssetReader(connector.Assets, util.GraylogDashboardInstallation), util.GraylogDashboardInstallation, cr.ToParams())
	if err != nil {
		return err
	}
	_, statusCode, err := connector.POST(oobContentPackInstallUrl, data)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.New("can't install dashboard")
	}
	return nil
}

func (connector *GraylogConnector) UploadContentPacks(cr *loggingService.LoggingService) error {
	if _, err := os.Stat(filepath.Join(dataDir, contentPacksDir)); err == nil {
		files, err := os.ReadDir(filepath.Join(dataDir, contentPacksDir))
		if err != nil {
			return err
		}
		for _, f := range files {
			if err = connector.UploadContentPack(f.Name(), cr); err != nil {
				return err
			}
		}
	} else if os.IsNotExist(err) {
		connector.Log.V(util.Debug).Info("Directory " + filepath.Join(dataDir, contentPacksDir) + " not found")
	} else {
		return err
	}
	return nil
}

func (connector *GraylogConnector) ManageContentPacks(cr *loggingService.LoggingService) error {
	contentPacks := strings.Split(cr.Spec.Graylog.ContentPackPaths, ",")

	for _, item := range contentPacks {
		fullName := filepath.Join(dataDir, filepath.Base(item))

		if err := util.DownloadFile(item, fullName); err != nil {
			return err
		}

		if strings.HasSuffix(fullName, ".zip") {
			connector.Log.V(util.Debug).Info("Unzip file " + fullName + " into " + filepath.Join(dataDir))
			files, err := util.Unzip(fullName, filepath.Join(dataDir))
			if err != nil {
				return err
			}

			connector.Log.V(util.Debug).Info("Unzipped:" + strings.Join(files, ";"))

			err = os.Remove(fullName)
			if err != nil {
				return err
			}
		} else if !strings.HasSuffix(fullName, ".json") {
			connector.Log.V(util.Error).Info("Incorrect content pack: " + fullName + " (it should be zip or json)")
		}
	}

	if err := connector.UploadContentPacks(cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) ManageContentPackTLS(ctx context.Context, cr *loggingService.LoggingService, clientSet kubernetes.Interface) error {

	for _, item := range cr.Spec.Graylog.ContentPacks {

		fullName := filepath.Join(dataDir, filepath.Base(item.URL))

		if err := util.DownloadFileTLS(ctx, item, fullName, clientSet, cr.GetNamespace()); err != nil {
			return err
		}

		if strings.HasSuffix(fullName, ".zip") {
			connector.Log.V(util.Debug).Info("Unzip file " + fullName + " into " + filepath.Join(dataDir))
			files, err := util.Unzip(fullName, filepath.Join(dataDir))
			if err != nil {
				return err
			}

			connector.Log.V(util.Debug).Info("Unzipped:" + strings.Join(files, ";"))

			err = os.Remove(fullName)
			if err != nil {
				return err
			}
		} else if !strings.HasSuffix(fullName, ".json") {
			connector.Log.V(util.Error).Info("Incorrect content pack: " + fullName + " (it should be zip o json)")
		}
	}

	if err := connector.UploadContentPacks(cr); err != nil {
		return err
	}

	return nil
}

func (connector *GraylogConnector) ManageDashboards(cr *loggingService.LoggingService) error {
	contentPacks, err := connector.GetData(contentpacksUrl, "content_packs", nil)
	if err != nil {
		return err
	}

	ExistsDefaultContentPack := false
	for _, contentPack := range contentPacks {
		if strings.EqualFold(contentPack.Id, oobContentPackId) {
			connector.Log.V(util.Debug).Info("Content pack with id " + oobContentPackId + " exists")
			ExistsDefaultContentPack = true
			if cr.Spec.Graylog.IsForceUpdate() {
				if err = connector.DeleteDefaultContentPack(); err != nil {
					return err
				}
			}
			break
		}
	}

	if cr.Spec.Graylog.IsForceUpdate() || (cr.Spec.Graylog.IsOnlyCreate() && !ExistsDefaultContentPack) {
		if err = connector.UploadDefaultContentPack(cr); err != nil {
			return err
		}

		if err = connector.InstallDefaultContentPack(cr); err != nil {
			return err
		}
	}
	return nil
}
