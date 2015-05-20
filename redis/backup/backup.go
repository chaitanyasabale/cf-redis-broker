package backup

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pivotal-cf/cf-redis-broker/recovery"
	"github.com/pivotal-cf/cf-redis-broker/recovery/task"
	redis "github.com/pivotal-cf/cf-redis-broker/redis/client"
	"github.com/pivotal-golang/lager"
)

func Backup(client redis.Client, logger lager.Logger) error {
	snapshot := Snapshot{client}
	img, err := snapshot.Create()
	if err != nil {
		fmt.Println("Snapshot failed: ", err.Error())
	}

	originalPath := img.Path()
	tmpSnapshotPath := uuid.New()
	artifactPath := "archive.tgz"

	img, err = recovery.NewPipeline(
		"redis-backup",
		logger,
		task.NewRename(tmpSnapshotPath),
		NewPackager(artifactPath),
		task.NewGeneric("Generic 1"),
		task.NewS3Upload("bucket-name", "endpoint", "key", "secret"),
		task.NewGeneric("Generic 2"),
	).Run(img)

	recovery.NewPipeline(
		"cleanup",
		logger,
		NewCleanup(
			originalPath,
			tmpSnapshotPath,
			artifactPath,
		),
	).Run(img)

	return err
}
