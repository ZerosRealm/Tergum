package manager

import (
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/restic"
)

func (man *Manager) Backup(job string, repo *entity.Repo, backup *entity.Backup) {
	man.log.WithFields("function", "backup", "job", job).Info("Starting job")
	out, err := man.restic.Backup(repo.Repo, backup.Source, repo.Password, backup.Exclude, job, repo.Settings...)
	if err != nil {
		man.jobErrors <- jobError{JobID: job, Error: err, Msg: out}
		man.log.WithFields("function", "backup", "job", job, "output", string(out)).Error("restic backup error:", err)
		return
	}

	man.log.WithFields("function", "backup", "job", job).Debug("output:", string(out))
}

func (man *Manager) Restore(job string, repo *entity.Repo, snapshot, target string, include, exclude []string) {
	man.log.WithFields("function", "restore", "job", job).Info("Starting job")

	out, err := man.restic.Restore(repo.Repo, repo.Password, snapshot, target, include, exclude, repo.Settings...)
	if err != nil {
		man.jobErrors <- jobError{JobID: job, Error: err, Msg: out}
		man.log.WithFields("function", "restore", "job", job, "output", string(out)).Error("restic backup error:", err)
		return
	}

	man.log.WithFields("function", "restore", "job", job).Debug("output:", string(out))

	man.jobMutex.Lock()
	defer man.jobMutex.Unlock()

	delete(man.jobs, job)
}

func (man *Manager) Stop(job string) {
	man.log.WithFields("function", "stop", "job", job).Info("Stopping job")

	resticJob, ok := man.jobs[job]
	if !ok {
		man.log.WithFields("function", "stop", "job", job).Info("Job not found")
		return
	}

	man.jobMutex.Lock()
	defer man.jobMutex.Unlock()

	resticJob.Cancel()
	delete(man.jobs, job)
}

func (man *Manager) GetSnapshots(repo *entity.Repo) ([]*restic.Snapshot, error) {
	man.log.WithFields("function", "getSnapshots").Info("Starting request")
	snapshots, err := man.restic.Snapshots(repo.Repo, repo.Password, repo.Settings...)
	if err != nil {
		return nil, err
	}
	for i, snapshot := range snapshots {
		if len(snapshot.Tags) == 0 {
			snapshots[i].Tags = make([]string, 0)
		}
	}

	if snapshots == nil {
		snapshots = make([]*restic.Snapshot, 0)
	}

	return snapshots, nil
}

func (man *Manager) Forget(repo *entity.Repo, snapshots []string, options *restic.ForgetOptions) ([]byte, error) {
	man.log.WithFields("function", "forget").Info("Starting request")
	out, err := man.restic.Forget(repo.Repo, repo.Password, snapshots, options, repo.Settings...)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (man *Manager) List(repo *entity.Repo, snapshot string) ([]*restic.FileNode, error) {
	man.log.WithFields("function", "list").Info("Starting request")
	nodes, err := man.restic.List(repo.Repo, repo.Password, snapshot, repo.Settings...)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
