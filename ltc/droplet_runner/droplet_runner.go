package droplet_runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/blob_store/blob"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
)

const (
	DropletStack  = "cflinuxfs2"
	DropletRootFS = "preloaded:" + DropletStack
)

//go:generate counterfeiter -o fake_droplet_runner/fake_droplet_runner.go . DropletRunner
type DropletRunner interface {
	UploadBits(dropletName, uploadPath string) error
	BuildDroplet(taskName, dropletName, buildpackUrl string, environment map[string]string, memoryMB, cpuWeight, diskMB int) error
	LaunchDroplet(appName, dropletName, startCommand string, startArgs []string, appEnvironmentParams app_runner.AppEnvironmentParams) error
	ListDroplets() ([]Droplet, error)
	RemoveDroplet(dropletName string) error
	ExportDroplet(dropletName string) (io.ReadCloser, io.ReadCloser, error)
	ImportDroplet(dropletName, dropletPath, metadataPath string) error
}

type Droplet struct {
	Name    string
	Created time.Time
	Size    int64
}

type dropletRunner struct {
	appRunner       app_runner.AppRunner
	taskRunner      task_runner.TaskRunner
	config          *config.Config
	blobStore       BlobStore
	appExaminer     app_examiner.AppExaminer
	proxyConfReader ProxyConfReader
}

//go:generate counterfeiter -o fake_blob_store/fake_blob_store.go . BlobStore
type BlobStore interface {
	List() ([]blob.Blob, error)
	Delete(path string) error
	Upload(path string, contents io.ReadSeeker) error
	Download(path string) (io.ReadCloser, error)

	blob_store.DropletStore
}

type annotation struct {
	DropletSource struct {
		DropletName string `json:"droplet_name"`
	} `json:"droplet_source"`
}

//go:generate counterfeiter -o fake_proxyconf_reader/fake_proxyconf_reader.go . ProxyConfReader
type ProxyConfReader interface {
	ProxyConf() (ProxyConf, error)
}

type ProxyConf struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
	NoProxy    string `json:"no_proxy"`
}

func New(appRunner app_runner.AppRunner, taskRunner task_runner.TaskRunner, config *config.Config, blobStore BlobStore, appExaminer app_examiner.AppExaminer, proxyConfReader ProxyConfReader) DropletRunner {
	return &dropletRunner{
		appRunner:       appRunner,
		taskRunner:      taskRunner,
		config:          config,
		blobStore:       blobStore,
		appExaminer:     appExaminer,
		proxyConfReader: proxyConfReader,
	}
}

func (dr *dropletRunner) ListDroplets() ([]Droplet, error) {
	blobs, err := dr.blobStore.List()
	if err != nil {
		return nil, err
	}

	droplets := []Droplet{}
	for _, blob := range blobs {
		pathComponents := strings.Split(blob.Path, "/")
		if len(pathComponents) == 2 && pathComponents[len(pathComponents)-1] == "droplet.tgz" {
			droplets = append(droplets, Droplet{Name: pathComponents[len(pathComponents)-2], Size: blob.Size, Created: blob.Created})
		}
	}

	return droplets, nil
}

func (dr *dropletRunner) UploadBits(dropletName, uploadPath string) error {
	uploadFile, err := os.Open(uploadPath)
	if err != nil {
		return err
	}

	return dr.blobStore.Upload(path.Join(dropletName, "bits.zip"), uploadFile)
}

func (dr *dropletRunner) BuildDroplet(taskName, dropletName, buildpackUrl string, environment map[string]string, memoryMB, cpuWeight, diskMB int) error {
	builderConfig := buildpack_app_lifecycle.NewLifecycleBuilderConfig([]string{buildpackUrl}, true, false)

	action := models.WrapAction(&models.SerialAction{
		Actions: []*models.Action{
			models.WrapAction(&models.DownloadAction{
				From: "http://file-server.service.cf.internal:8080/v1/static/cell-helpers/cell-helpers.tgz",
				To:   "/tmp",
				User: "vcap",
			}),
			models.WrapAction(&models.DownloadAction{
				From: "http://file-server.service.cf.internal:8080/v1/static/buildpack_app_lifecycle/buildpack_app_lifecycle.tgz",
				To:   "/tmp",
				User: "vcap",
			}),
			dr.blobStore.DownloadAppBitsAction(dropletName),
			dr.blobStore.DeleteAppBitsAction(dropletName),
			models.WrapAction(&models.RunAction{
				Path: "/bin/chmod",
				Dir:  "/tmp/app",
				Args: []string{"-R", "a+X", "."},
				User: "vcap",
			}),
			models.WrapAction(&models.RunAction{
				Path: "/tmp/builder",
				Dir:  "/",
				Args: builderConfig.Args(),
				User: "vcap",
			}),
			dr.blobStore.UploadDropletAction(dropletName),
			dr.blobStore.UploadDropletMetadataAction(dropletName),
		},
	})

	environment["CF_STACK"] = DropletStack
	environment["MEMORY_LIMIT"] = fmt.Sprintf("%dM", memoryMB)

	proxyConf, err := dr.proxyConfReader.ProxyConf()
	if err != nil {
		return err
	}
	environment["http_proxy"] = proxyConf.HTTPProxy
	environment["https_proxy"] = proxyConf.HTTPSProxy
	environment["no_proxy"] = proxyConf.NoProxy

	createTaskParams := task_runner.NewCreateTaskParams(
		action,
		taskName,
		DropletRootFS,
		"lattice",
		"BUILD",
		environment,
		[]*models.SecurityGroupRule{},
		memoryMB,
		cpuWeight,
		diskMB,
	)

	return dr.taskRunner.CreateTask(createTaskParams)
}

func (dr *dropletRunner) LaunchDroplet(appName, dropletName string, startCommand string, startArgs []string, appEnvironmentParams app_runner.AppEnvironmentParams) error {
	executionMetadata, err := dr.getExecutionMetadata(path.Join(dropletName, "result.json"))
	if err != nil {
		return err
	}

	dropletAnnotation := annotation{}
	dropletAnnotation.DropletSource.DropletName = dropletName

	annotationBytes, err := json.Marshal(dropletAnnotation)
	if err != nil {
		return err
	}

	if appEnvironmentParams.EnvironmentVariables == nil {
		appEnvironmentParams.EnvironmentVariables = map[string]string{}
	}

	appEnvironmentParams.EnvironmentVariables["PWD"] = "/home/vcap"
	appEnvironmentParams.EnvironmentVariables["TMPDIR"] = "/home/vcap/tmp"
	appEnvironmentParams.WorkingDir = "/home/vcap"

	proxyConf, err := dr.proxyConfReader.ProxyConf()
	if err != nil {
		return err
	}
	appEnvironmentParams.EnvironmentVariables["http_proxy"] = proxyConf.HTTPProxy
	appEnvironmentParams.EnvironmentVariables["https_proxy"] = proxyConf.HTTPSProxy
	appEnvironmentParams.EnvironmentVariables["no_proxy"] = proxyConf.NoProxy

	appParams := app_runner.CreateAppParams{
		AppEnvironmentParams: appEnvironmentParams,

		Name:         appName,
		RootFS:       DropletRootFS,
		StartCommand: "/tmp/launcher",
		AppArgs: []string{
			"/home/vcap/app",
			strings.Join(append([]string{startCommand}, startArgs...), " "),
			executionMetadata,
		},

		Annotation: string(annotationBytes),

		Setup: models.WrapAction(&models.SerialAction{
			LogSource: appName,
			Actions: []*models.Action{
				models.WrapAction(&models.DownloadAction{
					From: "http://file-server.service.cf.internal:8080/v1/static/cell-helpers/cell-helpers.tgz",
					To:   "/tmp",
					User: "vcap",
				}),
				models.WrapAction(&models.DownloadAction{
					From: "http://file-server.service.cf.internal:8080/v1/static/buildpack_app_lifecycle/buildpack_app_lifecycle.tgz",
					To:   "/tmp",
					User: "vcap",
				}),
				dr.blobStore.DownloadDropletAction(dropletName),
			},
		}),
	}

	return dr.appRunner.CreateApp(appParams)
}

func (dr *dropletRunner) getExecutionMetadata(path string) (string, error) {
	reader, err := dr.blobStore.Download(path)
	if err != nil {
		return "", err
	}

	var result struct {
		ExecutionMetadata string `json:"execution_metadata"`
	}

	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return "", err
	}

	return result.ExecutionMetadata, nil
}

func dropletMatchesAnnotation(dropletName string, a annotation) bool {
	return a.DropletSource.DropletName == dropletName
}

func (dr *dropletRunner) RemoveDroplet(dropletName string) error {
	apps, err := dr.appExaminer.ListApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		dropletAnnotation := annotation{}
		if err := json.Unmarshal([]byte(app.Annotation), &dropletAnnotation); err != nil {
			continue
		}

		if dropletMatchesAnnotation(dropletName, dropletAnnotation) {
			return fmt.Errorf("app %s was launched from droplet", app.ProcessGuid)
		}
	}

	blobs, err := dr.blobStore.List()
	if err != nil {
		return err
	}

	found := false
	for _, blob := range blobs {
		if strings.HasPrefix(blob.Path, dropletName+"/") {
			if err := dr.blobStore.Delete(blob.Path); err != nil {
				return err
			} else {
				found = true
			}
		}
	}

	if !found {
		return errors.New("droplet not found")
	}

	return nil
}

func (dr *dropletRunner) ExportDroplet(dropletName string) (io.ReadCloser, io.ReadCloser, error) {
	dropletReader, err := dr.blobStore.Download(path.Join(dropletName, "droplet.tgz"))
	if err != nil {
		return nil, nil, fmt.Errorf("droplet not found: %s", err)
	}

	metadataReader, err := dr.blobStore.Download(path.Join(dropletName, "result.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("metadata not found: %s", err)
	}

	return dropletReader, metadataReader, err
}

func (dr *dropletRunner) ImportDroplet(dropletName, dropletPath, metadataPath string) error {
	dropletFile, err := os.Open(dropletPath)
	if err != nil {
		return err
	}
	metadataFile, err := os.Open(metadataPath)
	if err != nil {
		return err
	}

	if err := dr.blobStore.Upload(path.Join(dropletName, "droplet.tgz"), dropletFile); err != nil {
		return err
	}

	if err := dr.blobStore.Upload(path.Join(dropletName, "result.json"), metadataFile); err != nil {
		return err
	}

	return nil
}
