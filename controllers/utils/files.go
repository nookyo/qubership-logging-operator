package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	v11 "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes"
)

const (
	BufferSize int = 100
)

type RestClient struct {
	Client *http.Client
	Auth   *Сreds
	Host   string
}

type Сreds struct {
	Name     string
	Password string
	Token    string
}

func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error(err, fmt.Sprintf("The file by path %s cannot be read", filePath))
		return "", err
	}

	return string(data), nil
}

// MustAssetReader loads and return the asset for the given name as bytes reader.
// Panics when the asset loading would return an error.
func MustAssetReader(assets embed.FS, asset string) string {
	content, _ := assets.ReadFile(asset)
	return string(content)
}

func ParseTemplate(fileContent, filePath string, parameters interface{}) (string, error) {
	funcMap := sprig.TxtFuncMap()
	funcMap["resIndex"] = GetFromResourceMap
	funcMap["timeNow"] = GetTimeNow
	funcMap["getAggregators"] = GetAggregatorIds

	goTemplate, err := template.New(filePath).Funcs(funcMap).Parse(fileContent)
	if err != nil {
		logger.Error(err, fmt.Sprintf("The template for file %s cannot be parsed", filePath))
		return "", err
	}

	writer := strings.Builder{}

	if err := goTemplate.Execute(&writer, parameters); err != nil {
		logger.Error(err, fmt.Sprintf("The template for file %s cannot be executed", filePath))
		return "", err
	}

	return writer.String(), nil
}

func DownloadFile(fullURLFile string, fileName string) error {
	logger.V(Debug).Info("Try download " + fullURLFile + " into file " + fileName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	// WARN: it necessary for https requests without valid certificates like on the dev-environment.
	// It can lead to the security issue on the production server
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: tr,
	}

	var urlFile *url.URL
	urlFile, err = url.Parse(fullURLFile)
	if err != nil {
		return err
	}

	// Preventing path traversal
	urlFile = urlFile.ResolveReference(urlFile)

	resp, err := client.Get(urlFile.String())
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	defer file.Close()

	logger.Info("Downloaded a file " + fileName + " with size " + strconv.Itoa(int(size)))

	return nil
}

func DownloadFileTLS(ctx context.Context, contentPackPath *v11.ContentPackPathHTTPConfig, fileName string, clientSet kubernetes.Interface, namespace string) error {
	logger.V(Debug).Info("Try download " + contentPackPath.URL + " into file " + fileName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	var user *Сreds
	var tlsConfig *tls.Config

	if contentPackPath != nil {
		if contentPackPath.HTTPConfig != nil {
			var name, pwd, token string
			name, pwd, token, tlsConfig, err = contentPackPath.HTTPConfig.GetCredentialsAndCertificates(ctx, clientSet, namespace)
			if err != nil {
				return err
			}
			if (name != "" && pwd != "") || token != "" {
				user = &Сreds{
					Name:     name,
					Password: pwd,
					Token:    token,
				}
			}
		}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns
	transport.MaxConnsPerHost = transport.MaxIdleConns
	transport.TLSClientConfig = tlsConfig
	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: transport,
		Timeout:   time.Duration(ConnectionTimeout) * time.Second,
	}

	var urlFile *url.URL
	urlFile, err = url.Parse(contentPackPath.URL)
	if err != nil {
		return err
	}

	// Preventing path traversal
	urlFile = urlFile.ResolveReference(urlFile)

	restClient := &RestClient{
		Client: client,
		Auth:   user,
	}

	var request *http.Request
	request, err = http.NewRequest(http.MethodGet, urlFile.String(), nil)
	if err != nil {
		return err
	}
	restClient.SetAuthHeader(request)

	var response *http.Response
	response, err = restClient.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var size int64
	size, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	defer file.Close()

	logger.Info("Downloaded a file " + fileName + " with size " + strconv.FormatInt(size, 10))

	return nil
}

func (restClient *RestClient) SetAuthHeader(request *http.Request) {

	if restClient.Auth != nil {
		var b bytes.Buffer
		if restClient.Auth.Name != "" && restClient.Auth.Password != "" {
			b.WriteString("Basic ")
			b.WriteString(base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s:%s", restClient.Auth.Name, restClient.Auth.Password))))
		} else if restClient.Auth.Token != "" {
			b.WriteString(fmt.Sprintf("Bearer %s", restClient.Auth.Token))
		}
		request.Header.Add("Authorization", b.String())
	}
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return filenames, err
			}
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func DataFromDirectory(assets embed.FS, directoryPath string, parameters interface{}) (map[string]string, error) {
	data := map[string]string{}

	files, err := assets.ReadDir(directoryPath)
	if err != nil {
		return data, err
	}

	for _, file := range files {

		if file.IsDir() {
			dirName := path.Join(directoryPath, file.Name())
			configs, err := assets.ReadDir(dirName)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Failed to read directory %s", dirName))
				return data, err
			}

			for _, file := range configs {
				filePath := path.Join(dirName, file.Name())
				data[file.Name()], err = ParseTemplate(MustAssetReader(assets, filePath), filePath, parameters)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Failed to read from file %s", filePath))
					return data, err
				}
			}
		} else {
			filePath := path.Join(directoryPath, file.Name())
			data[file.Name()], err = ParseTemplate(MustAssetReader(assets, filePath), filePath, parameters)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Failed to read from file %s", filePath))
				return data, err
			}
		}
	}

	return data, nil
}
