package jenkins

type Jenkins interface {
	// Input jobname and jobid string
	// Returns console text as []byte and error
	GetConsoleText(jobname, jobid string) ([]byte, error)
	// Input jobname and jobid string
	// Returns the content of workdir/results/service.json as []byte and error
	GetServiceJson(jobname, jobid string) ([]byte, error)
	// Input jobname and jobid string
	// Returns the content of workdir/results/package.tar as []byte and error
	ServeTar(jobname string, jobid string) ([]byte, error)
	// Input jobname string and parameters map. The params map will be passed
	// to the job--Build with Parameters
	// Returns queue id as string and error
	TriggerJob(jobname string, params map[string]string) (string, error)
	// Input queueid and jobname string
	// Return jobid, jobstatus string and error
	GetJobIdByQueueId(queueid string, jobname string) (string, string, error)
	// Input queueid and jobname string
	// Return jobid, jobstatus string and error only based on jenkins job status
	GetJobIdAndStatus(queueid string, jobname string) (string, string, error)
}
