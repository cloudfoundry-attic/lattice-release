package droplet_runner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

const (
	DropletStack  = "cflinuxfs2"
	DropletRootFS = "preloaded:" + DropletStack
)

//go:generate counterfeiter -o fake_droplet_runner/fake_droplet_runner.go . DropletRunner
type DropletRunner interface {
	UploadBits(dropletName, uploadPath string) error
	BuildDroplet(taskName, dropletName, buildpackUrl string, environment map[string]string) error
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
	appRunner      app_runner.AppRunner
	taskRunner     task_runner.TaskRunner
	config         *config.Config
	blobStore      BlobStore
	targetVerifier target_verifier.TargetVerifier
	appExaminer    app_examiner.AppExaminer
}

//go:generate counterfeiter -o fake_blob_store/fake_blob_store.go . BlobStore
type BlobStore interface {
	List() ([]blob_store.Blob, error)
	Delete(path string) error
	Upload(path string, contents io.ReadSeeker) error
	Download(path string) (io.ReadCloser, error)
}

type annotation struct {
	DropletSource struct {
		config.BlobTargetInfo
		DropletName string `json:"droplet_name"`
	} `json:"droplet_source"`
}

func New(appRunner app_runner.AppRunner, taskRunner task_runner.TaskRunner, config *config.Config, blobStore BlobStore, targetVerifier target_verifier.TargetVerifier, appExaminer app_examiner.AppExaminer) DropletRunner {
	return &dropletRunner{
		appRunner:      appRunner,
		taskRunner:     taskRunner,
		config:         config,
		blobStore:      blobStore,
		targetVerifier: targetVerifier,
		appExaminer:    appExaminer,
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

	return dr.blobStore.Upload(path.Join(dropletName, "bits.tgz"), uploadFile)
}

func (dr *dropletRunner) BuildDroplet(taskName, dropletName, buildpackUrl string, environment map[string]string) error {
	builderConfig := buildpack_app_lifecycle.NewLifecycleBuilderConfig([]string{buildpackUrl}, true, false)

	action := &models.SerialAction{
		Actions: []models.Action{
			&models.DownloadAction{
				From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
				To:   "/tmp",
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"get",
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					path.Join(dropletName, "bits.tgz"),
					"/tmp/bits.tgz",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/bin/mkdir",
				Dir:  "/",
				Args: []string{"/tmp/app"},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/bin/tar",
				Dir:  "/",
				Args: []string{"-C", "/tmp/app", "-xf", "/tmp/bits.tgz"},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/builder",
				Dir:  "/",
				Args: builderConfig.Args(),
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"put",
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					path.Join(dropletName, "droplet.tgz"),
					"/tmp/droplet",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"put",
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					path.Join(dropletName, "result.json"),
					"/tmp/result.json",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3tool",
				Dir:  "/",
				Args: []string{
					"delete",
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					path.Join(dropletName, "bits.tgz"),
				},
				User: "vcap",
			},
		},
	}

	environment["CF_STACK"] = DropletStack

	createTaskParams := task_runner.NewCreateTaskParams(
		action,
		taskName,
		DropletRootFS,
		"lattice",
		"BUILD",
		environment,
		[]models.SecurityGroupRule{},
	)

	return dr.taskRunner.CreateTask(createTaskParams)
}

func (dr *dropletRunner) LaunchDroplet(appName, dropletName string, startCommand string, startArgs []string, appEnvironmentParams app_runner.AppEnvironmentParams) error {
	executionMetadata, err := dr.getExecutionMetadata(path.Join(dropletName, "result.json"))
	if err != nil {
		return err
	}

	dropletAnnotation := annotation{}
	dropletAnnotation.DropletSource.BlobTargetInfo.TargetHost = dr.config.BlobTarget().TargetHost
	dropletAnnotation.DropletSource.BlobTargetInfo.TargetPort = dr.config.BlobTarget().TargetPort
	dropletAnnotation.DropletSource.BlobTargetInfo.BucketName = dr.config.BlobTarget().BucketName
	dropletAnnotation.DropletSource.DropletName = dropletName

	annotationBytes, err := json.Marshal(dropletAnnotation)
	if err != nil {
		return err
	}

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

		Setup: &models.SerialAction{
			LogSource: appName,
			Actions: []models.Action{
				&models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
					To:   "/tmp",
					User: "vcap",
				},
				&models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
					User: "vcap",
				},
				&models.RunAction{
					Path: "/tmp/s3tool",
					Args: []string{
						"get",
						dr.config.BlobTarget().AccessKey,
						dr.config.BlobTarget().SecretKey,
						fmt.Sprintf("http://%s:%d", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
						dr.config.BlobTarget().BucketName,
						path.Join(dropletName, "droplet.tgz"),
						"/tmp/droplet.tgz",
					},
					User: "vcap",
				},
				&models.RunAction{
					Path: "/bin/tar",
					Dir:  "/home/vcap",
					Args: []string{"-zxf", "/tmp/droplet.tgz"},
					User: "vcap",
				},
			},
		},
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

func dropletMatchesAnnotation(blobTarget config.BlobTargetInfo, dropletName string, a annotation) bool {
	return a.DropletSource.DropletName == dropletName &&
		a.DropletSource.TargetHost == blobTarget.TargetHost &&
		a.DropletSource.TargetPort == blobTarget.TargetPort &&
		a.DropletSource.BucketName == blobTarget.BucketName
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

		if dropletMatchesAnnotation(dr.config.BlobTarget(), dropletName, dropletAnnotation) {
			return fmt.Errorf("app %s was launched from droplet", app.ProcessGuid)
		}
	}

	blobs, err := dr.blobStore.List()
	if err != nil {
		return err
	}

	for _, blob := range blobs {
		if strings.HasPrefix(blob.Path, dropletName+"/") {
			if err := dr.blobStore.Delete(blob.Path); err != nil {
				return err
			}
		}
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
