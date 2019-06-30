package filter

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"net/http"
)

const TESTPLAN_MAX_SIZE = 20 * 1024 * 1024 // Limit the test plan yaml to be <= 20MB

type testplan struct {
	Tests []map[string]interface{} // required field
}

func ValidateTestPlan(c *gin.Context) {
	if c.Query("rerun") == "1" {
		// TODO: validate rerun test plan yaml file
		// bypass test plan validation if this is a rerun for now
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
		return
	}
	files := form.File["plan"]
	if len(files) != 1 || files[0].Size == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No test plan file"})
		return
	}
	for _, file := range files {
		if file.Size > TESTPLAN_MAX_SIZE {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("testplan yaml file size %d should be smaller than %d",
					file.Size, TESTPLAN_MAX_SIZE)})
			return
		}
		f, err := file.Open()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "open test plan file failed"})
			return
		}
		buf := make([]byte, file.Size)
		num, err := f.Read(buf)
		if err != nil || int64(num) != file.Size {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Read test plan error"})
			return
		}
		tp := testplan{}
		if err := yaml.Unmarshal(buf, &tp); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unmarshall test plan error"})
			return
		}
		if tp.Tests == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no test defined in testplan"})
			return
		}
	}
}
