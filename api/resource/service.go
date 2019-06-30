package resource

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"hfrd/api/utils"
	"hfrd/api/utils/couch"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var (
	jks         = jenkins.NewJenkins()
	logger      = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_RESOURCE)
	contentRepo = utils.GetValue("contentRepo").(string)
)

type ServiceList []Service

func (I ServiceList) Len() int {
	return len(I)
}

func (I ServiceList) Less(i, j int) bool {
	return I[i].CreatedAt > I[j].CreatedAt
}

func (I ServiceList) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func SortServices(theList []Service) {
	sort.Sort(ServiceList(theList))
}

func ServiceGetByRequestId(c *gin.Context) {
	requestid := c.Query(REQUEST_ID_REQ)
	env := c.DefaultQuery(ENV, DEFAULT_ENV)
	switch env {
	case CM:
		serveJobTarByQueueId(c, requestid, jenkins.NETWORK_CM)
	default:
		serveJobTarByQueueId(c, requestid, jenkins.NETWORK)
	}
}

func ServiceGetByServiceId(c *gin.Context) {
	env := c.DefaultQuery(ENV, DEFAULT_ENV)
	serviceid := c.Param(jenkins.SERVICE_ID)
	var params = map[string]string{ENV: env, jenkins.SERVICE_ID: serviceid, jenkins.METHOD: jenkins.GET}
	var jobName string
	switch env {
	case CM:
		jobName = jenkins.NETWORK_CM
	default:
		jobName = jenkins.NETWORK
	}
	var queueid, err = jks.TriggerJob(jobName, params)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	}
	c.Header(REQUEST_ID_RES, queueid)
	time.Sleep(3000 * time.Millisecond)

	serveJobTarByQueueId(c, queueid, jenkins.NETWORK)
}

func ServiceDeleteByRequestId(c *gin.Context) {
	env := c.DefaultQuery(ENV, DEFAULT_ENV)
	requestid := c.Query(REQUEST_ID_REQ)
	var jobid, result, err = jks.GetJobIdByQueueId(requestid, jenkins.NETWORK)
	if err != nil {
		logger.Warningf("Can not get job status with requestid: %s. status: \"%s\". Error: %s",
			requestid, result, err)
		c.String(http.StatusNotFound, fmt.Sprintf("Can not get job status with requestid: %s",
			requestid))
		return
	}
	if result == jenkins.FAIL {
		c.String(http.StatusBadRequest, fmt.Sprintf("The job with id: %s and requestid: %s failed",
			jobid, requestid))
		return
	}
	if result != jenkins.SUCCESS {
		c.String(http.StatusAccepted, fmt.Sprintf("The job with id: %s and requestid: %s is not completed yet",
			jobid, requestid))
		return
	}
	data, err := jks.GetServiceJson(jenkins.NETWORK, jobid)
	if err != nil {
		logger.Warningf("Can not get the service json file: %s", err)
		c.String(http.StatusBadRequest, "Can not get the service json file")
		return
	}
	var thejson map[string]interface{}
	if err := json.Unmarshal(data, &thejson); err != nil {
		logger.Warningf("Can not unmarshall the service json: %s", err)
		c.String(http.StatusBadRequest, "Can not unmarshall the service json")
		return
	}
	serviceid := thejson[jenkins.SERVICE_ID].(string)
	deleteService(c, env, serviceid)
}

func ServiceDeleteByServiceId(c *gin.Context) {
	env := c.DefaultQuery(ENV, DEFAULT_ENV)
	serviceid := c.Param(jenkins.SERVICE_ID)
	deleteService(c, env, serviceid)
}

func ServicePost(c *gin.Context) {
	p, exists := c.Get("plan")
	plan, ok := p.(Plan)
	if !exists || !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No valid plan provided"})
		return
	}
	var jobName string
	var params map[string]string
	switch plan.Env {
	case CM:
		jobName = jenkins.NETWORK_CM
		if plan.Name == "" {
			plan.Name = ENTERPRISE
		}
		// Check network configurations
		if plan.Config.NumOfOrgs < 1 {
			plan.Config.NumOfOrgs = DEFAULT_NUM_OF_ORGS
		}
		if plan.Config.NumOfPeers < 1 {
			plan.Config.NumOfPeers = DEFAULT_NUM_OF_PEERS
		}
		switch plan.Config.LedgerType {
		case LEVELDB, COUCHDB:
		// do nothing
		case "":
			// set default ledger type
			plan.Config.LedgerType = LEVELDB
		default:
			// unsupported ledger type
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"message": fmt.Sprintf("Unsupported ledger type: %s. We only support %s",
					plan.Config.LedgerType, []string{LEVELDB, COUCHDB})})
			return
		}
		params = map[string]string{ENV: plan.Env, jenkins.METHOD: jenkins.POST,
			jenkins.SERVICE_ID: "", jenkins.LOCATION: plan.Location,
			NUM_OF_ORGS_KEY:  strconv.Itoa(plan.Config.NumOfOrgs),
			NUM_OF_PEERS_KEY: strconv.Itoa(plan.Config.NumOfPeers),
			LEDGER_TYPE_KEY:  plan.Config.LedgerType}
	case BX_STAGING, BX_PROD:
		jobName = jenkins.NETWORK
		params = map[string]string{ENV: plan.Env, jenkins.METHOD: jenkins.POST,
			jenkins.SERVICE_ID: ""}
		if plan.Name == "" {
			plan.Name = STARTER
		} else if plan.Name == ENTERPRISE {
			params[ENV] = plan.Env + "-" + ENTERPRISE
			params[LOCATION] = plan.Location
			// Todo: support multiple orgs for Bluemix IBP Enterprise Plan
			if plan.Config.NumOfOrgs > 2 || plan.Config.NumOfOrgs <= 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{"message": "Currently we only support no more than 2 Orgs for bx EP"})
				return
			}
			params[NUM_OF_ORGS_KEY] = strconv.Itoa(plan.Config.NumOfOrgs)
			if plan.Config.NumOfPeers < 1 {
				plan.Config.NumOfPeers = DEFAULT_NUM_OF_PEERS
			}
			params[NUM_OF_PEERS_KEY] = strconv.Itoa(plan.Config.NumOfPeers)
			switch plan.Config.LedgerType {
			case LEVELDB, COUCHDB:
			// do nothing
			case "":
				// set default ledger type
				plan.Config.LedgerType = LEVELDB
			default:
				// unsupported ledger type
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{"message": fmt.Sprintf("Unsupported ledger type: %s. We only support %s",
						plan.Config.LedgerType, []string{LEVELDB, COUCHDB})})
				return
			}
			params[LEDGER_TYPE_KEY] = plan.Config.LedgerType
		}
	default:
		// should not go here because we have checked env
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": "Unsupported environment " + plan.Env})
		return
	}
	// check whether we support the plan name
	if !contains(SUPPORTED_ENV[plan.Env], plan.Name) {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("Unsupported plan name %s, we only support %s for %s",
				plan.Name, SUPPORTED_ENV[plan.Env], plan.Env)})
		return
	}
	params["contentrepo"] = contentRepo
	params["uid"] = c.Param("uid")
	logger.Debugf("Plan:%s", utils.PrettyStructString(plan))
	logger.Debugf("Job params:%s", utils.PrettyMapString(params))
	logger.Debugf("jobname: %s", jobName)
	queueid, err := jks.TriggerJob(jobName, params)
	if err != nil {
		logger.Errorf("Error triggering jenkins job: %s", err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	// skip location parameter for starter plan
	if plan.Name == STARTER {
		plan.Location = ""
	}
	// Job triggered in Jenkins server
	// start a go routine to store the job into DB and poll status of job
	job := couch.Job{QueueId: queueid, Name: jobName, ID: jobName + couch.SEPARATOR + queueid,
		Method: jenkins.POST, User: c.Param("uid"), PlanName: plan.Name, Env: plan.Env, Location: plan.Location}
	if plan.Name == ENTERPRISE {
		job.NumOfOrgs = plan.Config.NumOfOrgs
		job.NumOfPeers = plan.Config.NumOfPeers
		job.LedgerType = plan.Config.LedgerType
	} else if plan.Name == STARTER {
		// IBP Starter Plan topology
		job.NumOfOrgs = 2
		job.NumOfPeers = 1
		job.LedgerType = COUCHDB
	}
	go saveJob(&job)
	c.Header(REQUEST_ID_RES, queueid)
	c.String(http.StatusAccepted, "Your request has been accepted in "+plan.Env+" environment.")
}

func GetAllServices(c *gin.Context) {
	uid := c.Param("uid")
	services, err := GetServicesByUid(uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unable to get your network"})
		return
	} else {
		c.AbortWithStatusJSON(http.StatusOK, services)
		return
	}
}

// get all pending create IBP network job by user id
func GetPendingCreate(c *gin.Context) {
	uid := c.Param("uid")
	jobs, err := GetPendingCreateByUid(uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unable to get your pending jobs"})
		return
	} else {
		c.AbortWithStatusJSON(http.StatusOK, jobs)
		return
	}
}

func GetServicesByUid(uid string) ([]Service, error) {
	services := []Service{}
	if rows, err := couch.GetNetworkByUserId(uid); err != nil {
		logger.Errorf("Error getting network by userid: %s", err)
		return services, err
	} else {
		for rows.Next() {
			var service Service
			if err := rows.ScanValue(&service); err != nil {
				logger.Errorf("Error scan network doc: %s", err)
			} else {
				// successfully scanned network
				services = append(services, service)
			}
		}
		SortServices(services)
		return services, nil
	}
}

func GetPendingCreateByUid(uid string) ([]couch.Job, error) {
	jobs := []couch.Job{}
	if rows, err := couch.GetPendingCreateByUserId(uid); err != nil {
		logger.Errorf("Error getting pending network create job by userid: %s", err)
		return jobs, err
	} else {
		for rows.Next() {
			var job couch.Job
			if err := rows.ScanValue(&job); err != nil {
				logger.Errorf("Error scan Job doc: %s", err)
			} else {
				// successfully scanned job
				jobs = append(jobs, job)
			}
		}
		return jobs, nil
	}
}

// get all pending getting IBP network certs job by user id
func GetCertsPendingJobs(c *gin.Context) {
	uid := c.Param("uid")
	jobs, err := GetPendingCertsByUid(uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unable to get your pending jobs"})
		return
	} else {
		c.AbortWithStatusJSON(http.StatusOK, jobs)
		return
	}
}
func GetPendingCertsByUid(uid string) ([]couch.Job, error) {
	jobs := []couch.Job{}
	if rows, err := couch.GetPendingCertsByUserId(uid); err != nil {
		logger.Errorf("Error getting pending network certs job by userid: %s", err)
		return jobs, err
	} else {
		for rows.Next() {
			var job couch.Job
			if err := rows.ScanValue(&job); err != nil {
				logger.Errorf("Error scan Job doc: %s", err)
			} else {
				// successfully scanned job
				jobs = append(jobs, job)
			}
		}
		return jobs, nil
	}
}

func GetAllCertsJob4IBP(c *gin.Context) {
	uid := c.Param("uid")
	jobs, err := GetCertsJobByUid(uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unable to get your pending jobs"})
		return
	} else {
		c.AbortWithStatusJSON(http.StatusOK, jobs)
		return
	}
}

func GetCertsJobByUid(uid string) ([]couch.Job, error) {
	jobs := []couch.Job{}
	if rows, err := couch.GetCertsJobByUserId(uid); err != nil {
		logger.Errorf("Error getting all network certs job by userid: %s", err)
		return jobs, err
	} else {
		for rows.Next() {
			var job couch.Job
			if err := rows.ScanValue(&job); err != nil {
				logger.Errorf("Error scan Job doc: %s", err)
			} else {
				// successfully scanned job
				jobs = append(jobs, job)
			}
		}
		return jobs, nil
	}
}

// TODO: This should only be called by testing files...
func SetJenkins(jenkins jenkins.Jenkins) {
	jks = jenkins
}
