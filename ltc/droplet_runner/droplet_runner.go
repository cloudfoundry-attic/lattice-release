package droplet_runner

import (
	"fmt"
	"os"

	"strings"

	"time"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/goamz/goamz/s3"
)

//go:generate counterfeiter -o fake_droplet_runner/fake_droplet_runner.go . DropletRunner
type DropletRunner interface {
	UploadBits(dropletName, uploadPath string) error
	BuildDroplet(dropletName, buildpackUrl string) error
	LaunchDroplet(dropletName string, appEnvironmentParams app_runner.AppEnvironmentParams) error
	ListDroplets() ([]Droplet, error)
}

type Droplet struct {
	Name    string
	Created *time.Time
}

type dropletRunner struct {
	taskRunner     task_runner.TaskRunner
	config         *config.Config
	blobStore      blob_store.BlobStore
	blobBucket     blob_store.BlobBucket
	targetVerifier target_verifier.TargetVerifier
}

func New(taskRunner task_runner.TaskRunner, config *config.Config, blobStore blob_store.BlobStore, blobBucket blob_store.BlobBucket, targetVerifier target_verifier.TargetVerifier) *dropletRunner {
	return &dropletRunner{
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
		"preloaded:cflinuxfs2",
		"lattice",
		"BUILD",
		[]models.SecurityGroupRule{},
	)

	return dr.taskRunner.CreateTask(createTaskParams)
}

func (dr *dropletRunner) LaunchDroplet(dropletName string, appEnvironmentParams app_runner.AppEnvironmentParams) error {
	return nil
}
