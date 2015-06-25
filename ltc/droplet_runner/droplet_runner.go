package droplet_runner

import (
	"fmt"
	"os"

	"strings"

	"path"
	"time"

	"encoding/json"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
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
	BuildDroplet(dropletName, buildpackUrl string) error
	LaunchDroplet(appName, dropletName string, startCommand string, startArgs []string, appEnvironmentParams app_runner.AppEnvironmentParams) error
	ListDroplets() ([]Droplet, error)
}

type Droplet struct {
	Name    string
	Created *time.Time
}

type dropletRunner struct {
	appRunner      app_runner.AppRunner
	taskRunner     task_runner.TaskRunner
	config         *config.Config
	blobStore      blob_store.BlobStore
	blobBucket     blob_store.BlobBucket
	targetVerifier target_verifier.TargetVerifier
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

func New(appRunner app_runner.AppRunner, taskRunner task_runner.TaskRunner, config *config.Config, blobStore blob_store.BlobStore, blobBucket blob_store.BlobBucket, targetVerifier target_verifier.TargetVerifier) *dropletRunner {
	return &dropletRunner{
		appRunner:      appRunner,
		taskRunner:     taskRunner,
		config:         config,
		blobStore:      blobStore,
		blobBucket:     blobBucket,
		targetVerifier: targetVerifier,
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
				droplet := Droplet{Name: strings.TrimRight(prefix, "/")}

				modTime, err := time.Parse("2006-01-02T15:04:05.999Z", key.LastModified)
				if err == nil {
					droplet.Created = &modTime
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

func (dr *dropletRunner) BuildDroplet(dropletName, buildpackUrl string) error {
	builderConfig := buildpack_app_lifecycle.NewLifecycleBuilderConfig([]string{buildpackUrl}, false, false)

	action := &models.SerialAction{
		Actions: []models.Action{
			&models.DownloadAction{
				From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-support.tgz",
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
			},
			&models.RunAction{
				Path: "/bin/mkdir",
				Dir:  "/",
				Args: []string{"/tmp/app"},
			},
			&models.RunAction{
				Path: "/bin/tar",
				Dir:  "/",
				Args: []string{"-C", "/tmp/app", "-xf", "/tmp/bits.tgz"},
			},
			&models.RunAction{
				Path: "/tmp/builder",
				Dir:  "/",
				Args: builderConfig.Args(),
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
			},
		},
	}
	createTaskParams := task_runner.NewCreateTaskParams(
		action,
		dropletName,
		DropletRootFS,
		"lattice",
		"BUILD",
		map[string]string{"CF_STACK": DropletStack},
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
		startCommand = "/tmp/lrp-launcher"
		if result.DetectedStartCommand.Web != "" {
			startArgs = []string{result.DetectedStartCommand.Web}
		} else {
			startArgs = []string{dropletName}
		}
	}

	appParams := app_runner.CreateAppParams{
		AppEnvironmentParams: appEnvironmentParams,

		Name:         appName,
		RootFS:       DropletRootFS,
		StartCommand: startCommand,
		AppArgs:      startArgs,

		Setup: &models.SerialAction{
			LogSource: appName,
			Actions: []models.Action{
				&models.DownloadAction{
					From: "http://file_server.service.dc1.consul:8080/v1/static/lattice-support.tgz",
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
				},
				&models.RunAction{
					Path: "/bin/mkdir",
					Args: []string{"/tmp/app"},
				},
				&models.RunAction{
					Path: "/bin/tar",
					Dir:  "/tmp/app",
					Args: []string{"-zxf", "/tmp/droplet.tgz"},
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
