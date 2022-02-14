package service

type Services struct {
	RepoSvc      RepoService
	AgentSvc     AgentService
	BackupSvc    BackupService
	BackupSubSvc BackupSubscriberService
	ForgetSvc    ForgetService
	JobSvc       JobService
}

func NewServices(repoSvc *RepoService, agentSvc *AgentService, backupSvc *BackupService, backupSubSvc *BackupSubscriberService, forgetSvc *ForgetService, jobSvc *JobService) *Services {
	return &Services{
		RepoSvc:      *repoSvc,
		AgentSvc:     *agentSvc,
		BackupSvc:    *backupSvc,
		BackupSubSvc: *backupSubSvc,
		ForgetSvc:    *forgetSvc,
		JobSvc:       *jobSvc,
	}
}
