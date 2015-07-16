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
	"github.com/goamz/goamz/s3"
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
	blobStore      blob_store.BlobStore
	blobBucket     blob_store.BlobBucket
	targetVerifier target_verifier.TargetVerifier
	appExaminer    app_examiner.AppExaminer
}

type BuildResult struct {
	DetectedBuildpack    string       `json:"detected_buildpack"`
	ExecutionMetadata    string       `json:"execution_metadata"`
	BuildpackKey         string       `json:"buildpack_key"`
	DetectedStartCommand StartCommand `json:"detected_start_command"`
}

type StartCommand struct {
	Web string `json:"web"`
}

type Annotation struct {
	DropletSource struct {
		config.BlobTargetInfo
		DropletName string `json:"droplet_name"`
	} `json:"droplet_source"`
}

func New(appRunner app_runner.AppRunner, taskRunner task_runner.TaskRunner, config *config.Config, blobStore blob_store.BlobStore, blobBucket blob_store.BlobBucket, targetVerifier target_verifier.TargetVerifier, appExaminer app_examiner.AppExaminer) *dropletRunner {
	return &dropletRunner{
		appRunner:      appRunner,
		taskRunner:     taskRunner,
		config:         config,
		blobStore:      blobStore,
		blobBucket:     blobBucket,
		targetVerifier: targetVerifier,
		appExaminer:    appExaminer,
	}
}

func (dr *dropletRunner) ListDroplets() ([]Droplet, error) {
	listResp, err := dr.blobBucket.List("", "/", "", 0)
	if err != nil {
		return nil, err
	}

	droplets := []Droplet{}
	for _, prefix := range listResp.CommonPrefixes {
		subList, err := dr.blobBucket.List(prefix, "/", "", 0)
		if err != nil {
			continue
		}

		for _, key := range subList.Contents {
			if key.Key == prefix+"droplet.tgz" {
				droplet := Droplet{Name: strings.TrimRight(prefix, "/"), Size: key.Size}

				if modTime, err := time.Parse("2006-01-02T15:04:05.999Z", key.LastModified); err == nil {
					droplet.Created = modTime
				}

				droplets = append(droplets, droplet)
				break
			}
		}
	}

	return droplets, nil
}

func (dr *dropletRunner) UploadBits(dropletName, uploadPath string) error {
	fileInfo, err := os.Stat(uploadPath)
	if err != nil {
		return err
	}

	uploadFile, err := os.Open(uploadPath)
	if err != nil {
		return err
	}

	if targetUp, err := dr.targetVerifier.VerifyBlobTarget(
		dr.config.BlobTarget().TargetHost,
		dr.config.BlobTarget().TargetPort,
		dr.config.BlobTarget().AccessKey,
		dr.config.BlobTarget().SecretKey,
		dr.config.BlobTarget().BucketName,
	); !targetUp {
		return err
	}

	return dr.blobBucket.PutReader(fmt.Sprintf("%s/bits.tgz", dropletName), uploadFile, fileInfo.Size(), blob_store.DropletContentType, blob_store.DefaultPrivilege, s3.Options{})
}

func (dr *dropletRunner) BuildDroplet(taskName, dropletName, buildpackUrl string, environment map[string]string) error {
	builderConfig := buildpack_app_lifecycle.NewLifecycleBuilderConfig([]string{buildpackUrl}, true, false)

	action := &models.SerialAction{
		Actions: []models.Action{
			&models.DownloadAction{
				From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
				To:   "/tmp",
			},
			&models.RunAction{
				Path: "/tmp/s3downloader",
				Dir:  "/",
				Args: []string{
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					fmt.Sprintf("%s/bits.tgz", dropletName),
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
				Path: "/tmp/s3uploader",
				Dir:  "/",
				Args: []string{
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					fmt.Sprintf("%s/droplet.tgz", dropletName),
					"/tmp/droplet",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3uploader",
				Dir:  "/",
				Args: []string{
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					fmt.Sprintf("%s/result.json", dropletName),
					"/tmp/result.json",
				},
				User: "vcap",
			},
			&models.RunAction{
				Path: "/tmp/s3deleter",
				Dir:  "/",
				Args: []string{
					dr.config.BlobTarget().AccessKey,
					dr.config.BlobTarget().SecretKey,
					fmt.Sprintf("http://%s:%d/", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
					dr.config.BlobTarget().BucketName,
					fmt.Sprintf("%s/bits.tgz", dropletName),
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
	result, err := dr.downloadJSON(path.Join(dropletName, "result.json"))
	if err != nil {
		return err
	}

	if startCommand == "" {
		if result.DetectedStartCommand.Web != "" {
			startArgs = []string{result.DetectedStartCommand.Web}
		} else {
			// This is go buildpack-specific behavior where the executable
			// happens to match the droplet name.
			startArgs = []string{dropletName}
		}
	} else {
		startArgs = append([]string{startCommand}, startArgs...)
	}

	annotation := Annotation{}
	annotation.DropletSource.BlobTargetInfo.TargetHost = dr.config.BlobTarget().TargetHost
	annotation.DropletSource.BlobTargetInfo.TargetPort = dr.config.BlobTarget().TargetPort
	annotation.DropletSource.BlobTargetInfo.BucketName = dr.config.BlobTarget().BucketName
	annotation.DropletSource.DropletName = dropletName

	annotationBytes, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	appParams := app_runner.CreateAppParams{
		AppEnvironmentParams: appEnvironmentParams,

		Name:         appName,
		RootFS:       DropletRootFS,
		StartCommand: "/tmp/lrp-launcher",
		AppArgs:      startArgs,

		Annotation: string(annotationBytes),

		Setup: &models.SerialAction{
			LogSource: appName,
			Actions: []models.Action{
				&models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-cell-helpers.tgz",
					To:   "/tmp",
				},
				&models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/healthcheck.tgz",
					To:   "/tmp",
				},
				&models.RunAction{
					Path: "/tmp/s3downloader",
					Args: []string{
						dr.config.BlobTarget().AccessKey,
						dr.config.BlobTarget().SecretKey,
						fmt.Sprintf("http://%s:%d", dr.config.BlobTarget().TargetHost, dr.config.BlobTarget().TargetPort),
						dr.config.BlobTarget().BucketName,
						dropletName + "/droplet.tgz",
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

func (dr *dropletRunner) downloadJSON(path string) (*BuildResult, error) {
	reader, err := dr.blobBucket.GetReader(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(reader)

	result := BuildResult{}
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func dropletMatchesAnnotation(blobTarget config.BlobTargetInfo, dropletName string, annotation Annotation) bool {
	return annotation.DropletSource.DropletName == dropletName &&
		annotation.DropletSource.TargetHost == blobTarget.TargetHost &&
		annotation.DropletSource.TargetPort == blobTarget.TargetPort &&
		annotation.DropletSource.BucketName == blobTarget.BucketName
}

func (dr *dropletRunner) RemoveDroplet(dropletName string) error {
	apps, err := dr.appExaminer.ListApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		annotation := Annotation{}
		if err := json.Unmarshal([]byte(app.Annotation), &annotation); err != nil {
			continue
		}

		if dropletMatchesAnnotation(dr.config.BlobTarget(), dropletName, annotation) {
			return fmt.Errorf("app %s was launched from droplet", app.ProcessGuid)
		}
	}

	listResp, err := dr.blobBucket.List(dropletName+"/", "/", "", 0)
	if err != nil {
		return err
	}

	for _, key := range listResp.Contents {
		err := dr.blobBucket.Del(key.Key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dr *dropletRunner) ImportDroplet(dropletName, dropletPath, metadataPath string) error {

	dropletInfo, _ := os.Stat(dropletPath)
	metadataInfo, _ := os.Stat(metadataPath)

	dropletFile, _ := os.Open(dropletPath)
	metadataFile, _ := os.Open(metadataPath)

	if err := dr.blobBucket.PutReader(dropletName+"/droplet.tgz", dropletFile, dropletInfo.Size(), blob_store.DropletContentType, blob_store.DefaultPrivilege, s3.Options{}); err != nil {
		return err
	}

	if err := dr.blobBucket.PutReader(dropletName+"/result.json", metadataFile, metadataInfo.Size(), blob_store.DropletContentType, blob_store.DefaultPrivilege, s3.Options{}); err != nil {
		return err
	}

	return nil
}

func (dr *dropletRunner) ExportDroplet(dropletName string) (io.ReadCloser, io.ReadCloser, error) {
	dropletReader, err := dr.blobBucket.GetReader(fmt.Sprintf("%s/droplet.tgz", dropletName))
	if err != nil {
		return nil, nil, fmt.Errorf("droplet not found: %s", err)
	}

	metadataReader, err := dr.blobBucket.GetReader(fmt.Sprintf("%s/result.json", dropletName))
	if err != nil {
		return nil, nil, fmt.Errorf("metadata not found: %s", err)
	}

	return dropletReader, metadataReader, err
}
