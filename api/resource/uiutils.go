package resource

import (
	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"os"
	"sort"
)

type HFRDItem struct {
	Id        string `json:"id"`
	Name 	  string `json:"name"`
	Cdate     string `json:"cdate"`
	Queueid   string `json:"queueid"`
	Jobid     string `json:"jobid"`
	Status    string `json:"status"`
	Chartpath string `json:"chartpath,omitempty"`
}

type HFRDItemList struct {
	Items       []HFRDItem `json:"items"`
	Uid         string     `json:"uid"`
	ConsoleBase string     `json:"consolebase"`
	ApacheBase  string     `json:"apachebase"`
}

type ItemList []HFRDItem

func (I ItemList) Len() int {
	return len(I)
}

func (I ItemList) Less(i, j int) bool {
	return I[i].Cdate > I[j].Cdate
}

func (I ItemList) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

func SortItems(theList []HFRDItem) {
	sort.Sort(ItemList(theList))
}

var uiutilsLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_UIUTILS)

func GetList(c *gin.Context, suffix string) ([]string, error) {
	hfrdItemList, err := GetItems(c, suffix)
	if err != nil {
		return []string{}, err
	}

	items := hfrdItemList.Items
	var theList []string
	for _, item := range items {
		theList = append(theList, item.Id)
	}
	return theList, nil
}

func GetItems(c *gin.Context, suffix string) (HFRDItemList, error) {
	uid := c.Param("uid")
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/"
	pattern := rootPath + suffix
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return HFRDItemList{}, err
	}
	var theList HFRDItemList
	theList.Uid = uid
	baseUrl := utils.GetValue("jenkins.baseUrl").(string)
	u, err := url.Parse(baseUrl)
	if err != nil {
		theList.ConsoleBase = ""
	} else {
		u.User = nil
		theList.ConsoleBase = u.String()
	}
	theList.ApacheBase = utils.GetValue("apacheBaseUrl").(string)
	for _, path := range paths {
		//Get item submit date
		var cDate string
		file, err := os.Stat(path + "/" + "queueid")
		if err != nil {
			cDate = "Unavailable"
		} else {
			cDate = file.ModTime().String()
		}
		_, networkid := filepath.Split(path)
		var jobid, status string
		queueid, err := ioutil.ReadFile(rootPath + networkid + "/" + QUEUEID)
		if err != nil {
			queueid = []byte{}
			jobid = ""
			status = "Pending"
		} else {
			var jobname string
			if strings.HasSuffix(networkid, "-t") {
				jobname = jenkins.MODULETEST
			} else if strings.HasSuffix(networkid, "-n") {
				jobname = jenkins.K8SNETWORK
			} else if strings.HasSuffix(networkid, "-i"){
				jobname = jenkins.NETWORK_ICP
			}
			jobid, status, err = jks.GetJobIdAndStatus(string(queueid), jobname)
			// The job might be in queue and not scheduled yet
			if err != nil {
				status = "Pending"
			} else if len(jobid) > 0 && len(status) == 0 {
				status = "INPROGRESS"
			}
		}
		chartpath, err := ioutil.ReadFile(rootPath + networkid + "/" + CHARTPATH)
		if err != nil {
			chartpath = []byte{}
		}
		// TODO: Get network name for ibp4icp network name

		item := HFRDItem{networkid,"hfrd" ,cDate, string(queueid), jobid,
			status, string(chartpath)}
		theList.Items = append(theList.Items, item)
	}
	SortItems(theList.Items)
	return theList, nil
}
