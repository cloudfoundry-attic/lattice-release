package droplet_runner

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	"github.com/cloudfoundry-incubator/lattice/ltc/config"
	"github.com/cloudfoundry-incubator/lattice/ltc/config/blob_store"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_runner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/goamz/goamz/s3"
)

//go:generate counterfeiter -o fake_droplet_runner/fake_droplet_runner.go . DropletRunner
type DropletRunner interface {
	UploadBits(dropletName, uploadPath string) error
	BuildDroplet(dropletName, buildpackUrl string) error
}

type dropletRunner struct {
	taskRunner task_runner.TaskRunner
	config     *config.Config
	blobStore  blob_store.BlobStore
	blobBucket blob_store.BlobBucket
}

func New(taskRunner task_runner.TaskRunner, config *config.Config, blobStore blob_store.BlobStore, blobBucket blob_store.BlobBucket) *dropletRunner {
	return &dropletRunner{
		taskRunner: taskRunner,
		config:     config,
		blobStore:  blobStore,
		blobBucket: blobBucket,
	}
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
