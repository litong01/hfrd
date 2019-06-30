package filter

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"hfrd/api/resource"
	"io"
	"net/http"
	"strings"
)

const CERTS_MAX_SIZE = 100 * 1024 * 1024 // Limit the test plan yaml to be <= 100MB

func ValidateCertsTar(c *gin.Context) {
	if c.Query("rerun") == "1" {
		// bypass certs validation if this is a rerun
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
		return
	}
	files := form.File["cert"]
	if len(files) != 1 || files[0].Size == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No cert tar file"})
		return
	}
	file := files[0]
	logger.Debugf("certs tar file name: %s", file.Filename)
	if file.Size > CERTS_MAX_SIZE {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("certs tar file size %d should be smaller than %d",
				file.Size, CERTS_MAX_SIZE)})
		return
	}
	var r io.Reader
	r, err = file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "open certs tar file failed"})
		return
	}
	if strings.HasSuffix(file.Filename, ".tgz") || strings.HasSuffix(file.Filename, ".gz") {
		if r, err = gzip.NewReader(r); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("Error reading gz file: %s", err)})
			return
		}
	}
	certsTarReader := tar.NewReader(r)
	var v1Support, v2Support bool
	for {
		hdr, err := certsTarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("Read certs tar error:%s", err)})
			return
		}
		if strings.HasPrefix(hdr.Name, "./keyfiles") || strings.HasPrefix(hdr.Name, "keyfiles") {
			// this is the certs tar gz directory structure that we already support
			v1Support = true
		} else if strings.HasPrefix(hdr.Name, "./connection-profiles") ||
			strings.HasPrefix(hdr.Name, "connection-profiles") {
			v2Support = true
		}

	}
	if v1Support {
		c.Set(resource.CERT_VERSION_KEY, resource.CERT_V1)
	} else if v2Support {
		c.Set(resource.CERT_VERSION_KEY, resource.CERT_V2)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unsupported folder structure in certs tar"})
	}
}
