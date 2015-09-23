package model_helpers

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/gomega"
)

func NewValidActualLRP(guid string, index int32) *models.ActualLRP {
	actualLRP := &models.ActualLRP{
		ActualLRPKey:         models.NewActualLRPKey(guid, index, "some-domain"),
		ActualLRPInstanceKey: models.NewActualLRPInstanceKey("some-guid", "some-cell"),
		ActualLRPNetInfo:     models.NewActualLRPNetInfo("some-address", models.NewPortMapping(2222, 4444)),
		CrashCount:           33,
		CrashReason:          "badness",
		State:                models.ActualLRPStateRunning,
		Since:                1138,
		ModificationTag: models.ModificationTag{
			Epoch: "some-epoch",
			Index: 999,
		},
	}
	err := actualLRP.Validate()
	Expect(err).NotTo(HaveOccurred())

	return actualLRP
}

func NewValidDesiredLRP(guid string) *models.DesiredLRP {
	myRouterJSON := json.RawMessage(`{"foo":"bar"}`)
	desiredLRP := &models.DesiredLRP{
		ProcessGuid:          guid,
		Domain:               "some-domain",
		RootFs:               "some:rootfs",
		Instances:            1,
		EnvironmentVariables: []*models.EnvironmentVariable{{Name: "FOO", Value: "bar"}},
		Setup:                models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
		Action:               models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
		StartTimeout:         15,
		Monitor: models.WrapAction(models.EmitProgressFor(
			models.Timeout(models.Try(models.Parallel(models.Serial(&models.RunAction{Path: "ls", User: "name"}))),
				10*time.Second,
			),
			"start-message",
			"success-message",
			"failure-message",
		)),
		DiskMb:      512,
		MemoryMb:    1024,
		CpuWeight:   42,
		Routes:      &models.Routes{"my-router": &myRouterJSON},
		LogSource:   "some-log-source",
		LogGuid:     "some-log-guid",
		MetricsGuid: "some-metrics-guid",
		Annotation:  "some-annotation",
		EgressRules: []*models.SecurityGroupRule{{
			Protocol:     models.TCPProtocol,
			Destinations: []string{"1.1.1.1/32", "2.2.2.2/32"},
			PortRange:    &models.PortRange{Start: 10, End: 16000},
		}},
	}
	err := desiredLRP.Validate()
	Expect(err).NotTo(HaveOccurred())

	return desiredLRP
}

func NewValidTask(guid string) *models.Task {
	task := &models.Task{
		TaskGuid: guid,
		Domain:   "some-domain",
		RootFs:   "docker:///docker.com/docker",
		EnvironmentVariables: []*models.EnvironmentVariable{
			{
				Name:  "ENV_VAR_NAME",
				Value: "an environmment value",
			},
		},
		Action: models.WrapAction(&models.DownloadAction{
			From:     "old_location",
			To:       "new_location",
			CacheKey: "the-cache-key",
			User:     "someone",
		}),
		MemoryMb:         256,
		DiskMb:           1024,
		CpuWeight:        42,
		Privileged:       true,
		LogGuid:          "123",
		LogSource:        "APP",
		MetricsGuid:      "456",
		CreatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 00, time.UTC).UnixNano(),
		UpdatedAt:        time.Date(2014, time.February, 25, 23, 46, 11, 10, time.UTC).UnixNano(),
		FirstCompletedAt: time.Date(2014, time.February, 25, 23, 46, 11, 30, time.UTC).UnixNano(),
		ResultFile:       "some-file.txt",
		State:            models.Task_Pending,
		CellId:           "cell",

		Result:        "turboencabulated",
		Failed:        true,
		FailureReason: "because i said so",

		EgressRules: []*models.SecurityGroupRule{
			{
				Protocol:     "tcp",
				Destinations: []string{"0.0.0.0/0"},
				PortRange: &models.PortRange{
					Start: 1,
					End:   1024,
				},
				Log: true,
			},
			{
				Protocol:     "udp",
				Destinations: []string{"8.8.0.0/16"},
				Ports:        []uint32{53},
			},
		},

		Annotation: `[{"anything": "you want!"}]... dude`,
		// TODO: UNCOMMENT ME ONCE YOU SWITCH TO PROTOBUFS
		//CompletionCallbackUrl: "http://user:password@a.b.c/d/e/f",
		CompletionCallbackUrl: "http://@a.b.c/d/e/f",
	}

	err := task.Validate()
	Expect(err).NotTo(HaveOccurred())

	return task
}
