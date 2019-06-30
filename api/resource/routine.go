package resource

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"hfrd/api/utils"
	"hfrd/api/utils/couch"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"
	"io"
	"time"
)

// functions in this file should run in new goroutine

const (
	interval        = 10 * time.Second // poll job status interval
	serviceJsonPath = "workdir/results/service.json"
	networkJsonPath = "workdir/results/network.json"
)

var routineLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_ROUTINE)

func init() {
	if rows, err := couch.GetPendingJobs(); err != nil {
		routineLogger.Errorf("Error get pending jobs: %s", err)
	} else {
		var createNum int
		for rows.Next() {
			var job couch.Job
			if err := rows.ScanDoc(&job); err != nil {
				routineLogger.Errorf("Error scan job doc: %s", err)
			} else {
				routineLogger.Debugf("pending job: %s", utils.PrettyStructString(job))
				if job.Method == jenkins.POST {
					createNum++
				}
				go waitJob(&job)
			}
		}
		routineLogger.Debugf("Found %d pending jobs in total, %d service creation job(s)",
			rows.TotalRows(), createNum)
		rows.Close()
	}
}

func saveJob(job *couch.Job) {
	rev, err := couch.ForceSaveJob(job)
	if err != nil {
		routineLogger.Errorf("Error saving job: %s", err)
		return
	}
	// find network by serviceId and mark it as deleting
	if job.Method == jenkins.DELETE && len(job.ServiceId) > 0 {
		if rows, err := couch.GetNetworkByServiceId(job.ServiceId); err != nil {
			routineLogger.Errorf("Error get network by serviceid[%s]:%s", job.ServiceId, err)
		} else {
			var network couch.Network
			var rowNum int
			for rows.Next() {
				rowNum++
				if err := rows.ScanDoc(&network); err != nil {
					routineLogger.Errorf("Error scan network doc: %s", err)
				} else if network.Deleted == false {
					network.Deleting = true
					if _, err := couch.ForceSaveNetwork(network); err != nil {
						routineLogger.Errorf("Error saving network into DB: %s\nnetwork:%s",
							err, utils.PrettyStructString(network))
					} else {
						routineLogger.Debugf("Successfully marked network:%s as deleting",
							network.NetworkId)
					}
				}
			}
			if rowNum != 1 {
				// for Bluemix starter plan and POK on-prem cluster, service id should be 1:1
				// mapped to networkid
				routineLogger.Warningf("%d network found with service id: %s",
					rowNum, job.ServiceId)
			}
		}
	}
	job.Rev = rev
	waitJob(job)
	routineLogger.Debugf("Exiting goroutine for job %s", utils.PrettyStructString(job))
}

func waitJob(job *couch.Job) {
	var status string
	var jobId string
	var err error
	for {
		time.Sleep(interval)
		// job still in progress?
		jobId, status, err = jks.GetJobIdByQueueId(job.QueueId, job.Name)
		if err == nil && status == jenkins.INPROGRESS {
			routineLogger.Debug("[%s:%s]job is still in progress", job.Name, job.QueueId)
			if job.JobId == "" {
				job.JobId = jobId
				rev, err := couch.ForceSaveJob(job)
				if err != nil {
					routineLogger.Errorf("[%s:%s]Error updating job id in DB: %s", err)
				} else {
					job.Rev = rev
				}
			}
			continue
		}
		// job error
		if err != nil {
			routineLogger.Errorf("[%s:%s]Error getting job status: %s",
				job.Name, job.QueueId, err)
			break
		}
		job.Status = status
		job.JobId = jobId
		// update job status in DB
		if _, err := couch.ForceSaveJob(job); err != nil {
			routineLogger.Errorf("[%s:%s]Error updating job status in DB: %s", job.Name, job.QueueId, err)
		}
		// get network info and store in DB if job succeeds
		if status == jenkins.SUCCESS && job.Method == jenkins.POST {
			pkgBytes, err := jks.ServeTar(job.Name, job.JobId)
			if err != nil {
				routineLogger.Errorf("[%s:%s]Error getting job package tar data: %s",
					job.Name, job.JobId, err)
				break
			}
			buf := bytes.NewBuffer(pkgBytes)
			tr := tar.NewReader(buf)
			var network couch.Network
			network.CreatedAt = time.Now().Unix()
			network.User = job.User
			network.PlanName = job.PlanName
			network.Env = job.Env
			network.Location = job.Location
			network.NumOfOrgs = job.NumOfOrgs
			network.NumOfPeers = job.NumOfPeers
			network.LedgerType = job.LedgerType
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break // End of archive
				}
				if err != nil {
					routineLogger.Errorf("[%s:%s]Error reading tar hdr: %s",
						job.Name, job.QueueId, err)
					break
				}
				switch hdr.Name {
				case serviceJsonPath:
					serviceJsonBytes := make([]byte, hdr.Size)
					var content map[string]interface{}
					length, err := tr.Read(serviceJsonBytes)
					routineLogger.Debugf("[%s:%s]service.json length: %d, hdr.Size:%d",
						job.Name, job.QueueId, length, hdr.Size)
					if (err == nil || err == io.EOF) && length == int(hdr.Size) {
						// successfully read service json bytes
						if err := json.Unmarshal(serviceJsonBytes, &content); err != nil {
							routineLogger.Errorf("[%s:%s]Error unmarshalling service json: %s",
								job.Name, job.QueueId, err)
							break
						}
						if serviceid, ok := content["serviceid"]; ok {
							// assure serviceid is string
							if serviceIdStr, ok := serviceid.(string); ok {
								network.ServiceId = serviceIdStr
							} else {
								routineLogger.Errorf("[%s:%s]serviceid is not string: %v",
									job.Name, job.QueueId, serviceid)
							}

						} else {
							routineLogger.Errorf("[%s:%s]No serviceid found in service.json: %s",
								job.Name, job.QueueId, string(serviceJsonBytes))
						}

					} else {
						routineLogger.Errorf("[%s:%s]Error reading service json bytes. "+
							"length: %d, error: %s", job.Name, job.QueueId, length, err)
					}
				case networkJsonPath:
					networkJsonBytes := make([]byte, hdr.Size)
					var content map[string]interface{}
					length, err := tr.Read(networkJsonBytes)
					routineLogger.Debugf("[%s:%s] network.json length: %d, hdr.Size:%d",
						job.Name, job.QueueId, length, hdr.Size)
					if (err == nil || err == io.EOF) && length == int(hdr.Size) {
						// successfully read network json bytes
						if err := json.Unmarshal(networkJsonBytes, &content); err != nil {
							routineLogger.Errorf("[%s:%s]Error unmarshalling network json: %s",
								job.Name, job.QueueId, err)
							break
						}
						network.Network = content
						var org1 map[string]interface{}
						if _, ok := content["PeerOrg1"]; ok { // Enterprise Plan
							if _, ok = content["PeerOrg1"].(map[string]interface{}); ok {
								org1 = content["PeerOrg1"].(map[string]interface{})
							}
						} else if _, ok := content["org1"]; ok { // Starter Plan
							if _, ok = content["org1"].(map[string]interface{}); ok {
								org1 = content["org1"].(map[string]interface{})
							}
						} else {
							routineLogger.Errorf("[%s:%s]No org found in network.json: %s",
								job.Name, job.QueueId, string(networkJsonBytes))
						}
						if networkId, ok := org1["network_id"]; ok {
							// assure networkId is string
							if networkIdStr, ok := networkId.(string); ok && len(networkIdStr) > 0 {
								network.NetworkId = networkIdStr
							} else {
								routineLogger.Errorf("[%s:%s]Invalid network id: %s",
									job.Name, job.QueueId, networkId)
							}
						} else {
							routineLogger.Errorf("[%s:%s]No network_id found in %v", org1)
						}
					} else {
						routineLogger.Errorf("[%s:%s]Error reading network json bytes. "+
							"length: %d, error: %s", job.Name, job.QueueId, length, err)
					}
				default:
					// ignore
				} // end of switch
			} // end of for loop
			// only update job.env for ibpcerts job -- resuse env field to save networkId
			if (job.Name == jenkins.IBPCERTS) && len(network.NetworkId) > 0 {
				job.Env = network.NetworkId
				if _, err := couch.ForceSaveJob(job); err != nil {
					routineLogger.Errorf("[%s:%s]Error updating job status in DB: %s", job.Name, job.QueueId, err)
				}
			}
			if len(network.NetworkId) > 0 && len(network.ServiceId) > 0 {
				network.ID = network.NetworkId
				network.JobId = job.JobId
				network.JobName = job.Name
				if _, err := couch.ForceSaveNetwork(network); err == nil {
					routineLogger.Debugf("Successfully stored network %s into DB", network.ID)
				} else {
					routineLogger.Errorf("[%s:%s] Error saving network to DB: %s",
						job.Name, job.QueueId, err)
				}
			} else {
				routineLogger.Errorf("[%s:%s]Will not save the network into DB: %v",
					job.Name, job.QueueId, network)
			}
		} else if status == jenkins.SUCCESS && job.Method == jenkins.DELETE && len(job.ServiceId) > 0 {
			// mark the network as deleted
			routineLogger.Debugf("Delete network job[%s:%s] succeeded, will mark network as deleted",
				job.Name, job.QueueId)
			if rows, err := couch.GetNetworkByServiceId(job.ServiceId); err != nil {
				routineLogger.Errorf("Error get network by serviceid[%s]:%s", job.ServiceId, err)
			} else {
				var network couch.Network
				var rowNum int
				for rows.Next() {
					rowNum++
					if err := rows.ScanDoc(&network); err != nil {
						routineLogger.Errorf("Error scan network doc: %s", err)
					} else {
						network.Deleted = true
						network.Deleting = false
						if _, err := couch.ForceSaveNetwork(network); err != nil {
							routineLogger.Errorf("Error saving network into DB: %s\nnetwork:%s",
								err, utils.PrettyStructString(network))
						} else {
							routineLogger.Debugf("Successfully marked network:%s as deleted",
								network.NetworkId)
						}
					}
				}
				if rowNum != 1 {
					// for Bluemix starter plan and POK on-prem cluster, service id should be 1:1
					// mapped to networkid
					routineLogger.Warningf("%d network found with service id: %s",
						rowNum, job.ServiceId)
				}
			}
		}
		break
	}
}
