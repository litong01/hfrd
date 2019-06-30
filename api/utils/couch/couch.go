package couch

import (
	"context"
	"encoding/json"
	"fmt"
	. "hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"

	"github.com/flimzy/kivik"       // Stable version of Kivik
	_ "github.com/go-kivik/couchdb" // The CouchDB driver
)

const (
	// DB name
	JOB_DB     = "hfrd_job"
	NETWORK_DB = "hfrd_network"

	SEPARATOR = "_"

	// Job DB Design Doc
	DESIGN_ID_JOB         = "_design/apiserver_v1"
	PENDING_JOBS_VIEW     = "pending_jobs" // For jobs with empty status and method of "POST" or "DELETE"
	PENDING_CREATE_BY_UID = "pending_create_by_uid"
	ALL_CERTS_JOB_VIEW    = "all_certs_jobs"

	// Network DB Design Doc
	DESIGN_ID_NETWORK = "_design/apiserver_v1"
	BY_SERVICEID      = "byServiceId"
	BY_USERID         = "byUserId"

	updateDocConflictErr = "Conflict: Document update conflict."
)

var (
	logger    = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_COUCH)
	couchUrl  = GetValue("couchUrl").(string)
	client    *kivik.Client
	jobDB     *kivik.DB
	networkDB *kivik.DB

	design_doc_job = map[string]interface{}{
		"_id": DESIGN_ID_JOB,
		"views": map[string]interface{}{
			PENDING_JOBS_VIEW: map[string]interface{}{
				"map": `function(doc) {
					if (doc.status === "" && doc.method &&
					(doc.method == "POST" || doc.method == "DELETE")) {
    						emit(doc._id, doc);
 					 }
				}`,
			},
			PENDING_CREATE_BY_UID: map[string]interface{}{
				"map": `function(doc) {
					if (doc.user && doc.status === "" && doc.method &&
					(doc.method == "POST")) {
    						emit(doc.user, doc);
 					 }
				}`,
			},
			ALL_CERTS_JOB_VIEW: map[string]interface{}{
				"map": `function(doc) {
					if (doc.name === "ibpcerts" && doc.method && doc.status != "" &&
					(doc.method == "POST" || doc.method == "DELETE")) {
    						emit(doc.user, doc);
 					 }
				}`,
			},
		},
	}

	design_doc_network = map[string]interface{}{
		"_id": DESIGN_ID_NETWORK,
		"views": map[string]interface{}{
			BY_SERVICEID: map[string]interface{}{
				"map": `function (doc) {
					if (doc.serviceId) {
						emit(doc.serviceId, doc);
					}
				}`,
			},
			BY_USERID: map[string]interface{}{
				"map": `function (doc) {
						if (doc.user && doc.deleted === false && doc.deleting === false) {
							var retDoc = {}
							var fields = ["networkId", "serviceId", "planName", "env", "jobName", "jobId"]
							for (var i in fields) {
								retDoc[fields[i]] = doc[fields[i]]
							}
							// IBP Enterprise Plan specific fields
							var epFields = ["location", "numOfOrgs", "numOfPeers", "ledgerType"]
							if (doc.planName === "ep") {
								for (var i in epFields) {
									retDoc[epFields[i]] = doc[epFields[i]]
								}
							}
							// convert unix timestamp into human readable date and time
							var dateTime = new Date(doc.createdAt * 1000);
							retDoc.createdAt = dateTime.toISOString()
							emit(doc.user, retDoc);
						}
					}`,
			},
		},
	}
)

func init() {
	var err error
	client, err = kivik.New(context.TODO(), "couch", couchUrl)
	if err != nil {
		panic(err)
	}

	// Create necessary DBs if not exist
	dbs := []string{JOB_DB, NETWORK_DB}
	for _, db := range dbs {
		exist, err := client.DBExists(context.TODO(), db)
		if err != nil {
			logger.Errorf("Error checking existance of %s: %s", db, err)
			panic(err)
		}
		if !exist {
			// create DB
			if err := client.CreateDB(context.TODO(), db); err != nil {
				logger.Errorf("Error creating db %s: %s", db, err)
				panic(err)
			} else {
				logger.Debugf("Successfully created db: %s", db)
			}

		} else {
			logger.Debugf("db %s already exists", db)
		}
	}
	if jobDB, err = client.DB(context.TODO(), JOB_DB); err != nil {
		logger.Errorf("Error getting job DB: %s", err)
		panic(err)
	}

	if err := createDesignDoc(jobDB, design_doc_job); err != nil {
		logger.Errorf("Error creating design doc for Job DB: %s", err)
		panic(err)
	}
	if networkDB, err = client.DB(context.TODO(), NETWORK_DB); err != nil {
		logger.Errorf("Error getting network DB: %s", err)
		panic(err)
	}
	if err := createDesignDoc(networkDB, design_doc_network); err != nil {
		logger.Errorf("Error creating design doc for Network DB: %s", err)
		panic(err)
	}
	logger.Debugf("Couchdb is initialized successfully!")
}

type Job struct {
	ID         string `json:"_id"`
	Rev        string `json:"_rev,omitempty"`
	Name       string `json:"name"`
	QueueId    string `json:"queueId"`
	JobId      string `json:"jobId"`
	Status     string `json:"status"`
	ServiceId  string `json:"serviceId"` // only available for service deletion job
	Method     string `json:"method"`    // POST or DELETE
	User       string `json:"user"`
	Email      string `json:"email"`
	PlanName   string `json:"planName"`   // ONLY available for service creation job
	Env        string `json:"env"`        // ONLY available for service creation job
	Location   string `json:"location"`   // location id for IBP Enterprise Plan
	NumOfOrgs  int    `json:"numOfOrgs"`  // number of organizations
	NumOfPeers int    `json:"numOfPeers"` // number of peers
	LedgerType string `json:"ledgerType"` // ledger type
}

type Network struct {
	ID        string                 `json:"_id"`
	Rev       string                 `json:"_rev,omitempty"`
	NetworkId string                 `json:"networkId"`
	ServiceId string                 `json:"serviceId"`
	User      string                 `json:"user"`
	Email     string                 `json:"email"`
	Network   map[string]interface{} `json:"network"` // contain all Orgs' key and secret
	// Deleting: false, Deleted: false -> network is active from hfrd perspective
	// Deleting: true, Deleted: false -> deleting, job not completed yet
	// Deleting: false, Deleted: true -> blockchain network is completely deleted
	// Deleting: true, Deleted: true -> should NOT happen
	Deleting   bool   `json:"deleting"`
	Deleted    bool   `json:"deleted"`
	PlanName   string `json:"planName"`
	Env        string `json:"env"`
	CreatedAt  int64  `json:"createdAt"` // Unix timestamp in seconds
	JobId      string `json:"jobId"`
	JobName    string `json:"jobName"`
	Location   string `json:"location"`   // location id for IBP Enterprise Plan
	NumOfOrgs  int    `json:"numOfOrgs"`  // number of organizations
	NumOfPeers int    `json:"numOfPeers"` // number of peers
	LedgerType string `json:"ledgerType"` // ledger type

}

// save doc again with rev conflict
func ForceSaveJob(job *Job) (string, error) {
	rev, err := jobDB.Put(context.TODO(), job.ID, job)
	if err == nil {
		return rev, err
	} else if err.Error() == updateDocConflictErr {
		// rev conflict
		if row, err := jobDB.Get(context.TODO(), job.ID); err != nil {
			logger.Error("Error getting job doc %s: %s", job.ID, err)
		} else {
			var doc Job
			if err = row.ScanDoc(&doc); err != nil {
				return rev, err
			}
			job.Rev = doc.Rev
			// store the document again with existing rev
			if rev, err = jobDB.Put(context.TODO(), job.ID, job); err != nil {
				return rev, err
			} else {
				logger.Debugf("save job %s successfully", job.ID)
				return rev, nil
			}
		}
	}
	return rev, err
}

func ForceSaveNetwork(network Network) (string, error) {
	rev, err := networkDB.Put(context.TODO(), network.ID, network)
	if err == nil {
		return rev, err
	} else if err.Error() == updateDocConflictErr {
		// rev conflict
		if row, err := networkDB.Get(context.TODO(), network.ID); err != nil {
			logger.Error("Error getting network doc %s: %s", network.ID, err)
		} else {
			var doc Network
			if err = row.ScanDoc(&doc); err != nil {
				return rev, err
			}
			network.Rev = doc.Rev
			// store the document again with existing rev
			if rev, err = networkDB.Put(context.TODO(), network.ID, network); err != nil {
				return rev, err
			} else {
				logger.Debugf("save/update network %s successfully", network.ID)
			}
		}
	}
	return rev, err
}

// return all network creation job with empty status
func GetPendingJobs() (*kivik.Rows, error) {
	return jobDB.Query(context.TODO(), DESIGN_ID_JOB, PENDING_JOBS_VIEW, kivik.Options{"include_docs": true})
}

func GetNetworkByServiceId(serviceId string) (*kivik.Rows, error) {
	key, _ := json.Marshal([]string{serviceId})
	return networkDB.Query(context.TODO(), DESIGN_ID_NETWORK, BY_SERVICEID,
		kivik.Options{"include_docs": true, "keys": string(key)})
}

func GetNetworkByUserId(uid string) (*kivik.Rows, error) {
	key, _ := json.Marshal([]string{uid})
	return networkDB.Query(context.TODO(), DESIGN_ID_NETWORK, BY_USERID,
		kivik.Options{"keys": string(key)})
}

func GetPendingCreateByUserId(uid string) (*kivik.Rows, error) {
	key, _ := json.Marshal([]string{uid})
	return jobDB.Query(context.TODO(), DESIGN_ID_JOB, PENDING_CREATE_BY_UID,
		kivik.Options{"keys": string(key)})
}

func GetPendingCertsByUserId(uid string) (*kivik.Rows, error) {
	key, _ := json.Marshal([]string{uid})
	return jobDB.Query(context.TODO(), DESIGN_ID_JOB, PENDING_CREATE_BY_UID,
		kivik.Options{"keys": string(key)})
}

func GetCertsJobByUserId(uid string) (*kivik.Rows, error) {
	key, _ := json.Marshal([]string{uid})
	return jobDB.Query(context.TODO(), DESIGN_ID_JOB, ALL_CERTS_JOB_VIEW,
		kivik.Options{"keys": string(key)})
}

func DeleteCertsJobById(id string, ver string) (string, error) {
	return jobDB.Delete(context.TODO(), id, ver,
		kivik.Options{})
}

func createDesignDoc(db *kivik.DB, designDoc map[string]interface{}) error {
	var designDocId string
	if _, ok := designDoc["_id"]; !ok {
		return fmt.Errorf("_id does not exist in designDoc: %s", PrettyStructString(designDoc))
	} else if designDocId, ok = designDoc["_id"].(string); !ok || len(designDocId) == 0 {
		return fmt.Errorf("Invalid designDoc _id: %s", PrettyStructString(designDoc))
	}
	_, err := db.Put(context.TODO(), designDocId, designDoc)
	// design doc already exists?
	if err != nil && err.Error() == updateDocConflictErr {
		if row, err := db.Get(context.TODO(), designDocId); err != nil {
			logger.Error("Error getting design doc %s: %s", designDocId, err)
			return err
		} else {
			var doc map[string]interface{}
			if err = row.ScanDoc(&doc); err != nil {
				return err
			}
			designDoc["_rev"] = doc["_rev"]
			// store  the document again with existing rev
			if rev, err := db.Put(context.TODO(), designDocId, designDoc); err != nil {
				return err
			} else {
				logger.Debugf("Updated design doc to rev: %s", rev)
			}
		}
		return nil
	} else if err == nil {
		logger.Debugf("Successfully created design doc %s for %s DB", designDocId, db.Name())
	}
	return err
}
