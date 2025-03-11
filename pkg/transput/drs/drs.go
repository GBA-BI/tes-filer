package drs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GBA-BI/tes-filer/pkg/checker"
	md5checker "github.com/GBA-BI/tes-filer/pkg/checker/md5"
	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	transputhttp "github.com/GBA-BI/tes-filer/pkg/transput/http"
)

var drsObjectReg = regexp.MustCompile(`^[A-Za-z0-9\.\-_~]+$`)

type drsTransput struct {
	transput.DefaultTransput

	insecureDirDomain string
	aaiPassport       string

	client *http.Client
	logger log.Logger
}

func NewDRSTransput(cfg *Config, logger log.Logger) (transput.Transput, error) {
	drs := &drsTransput{
		client:            &http.Client{},
		insecureDirDomain: cfg.InsecureDirDomain,
		aaiPassport:       cfg.AAIPassport,
		logger:            logger,
	}

	return drs, nil
}

func (d *drsTransput) DownloadFile(ctx context.Context, local, remote string) error {
	parsedURL, err := url.Parse(remote)
	if err != nil {
		return err
	}
	objectID := filepath.Base(remote)
	if !drsObjectReg.MatchString(objectID) {
		return fmt.Errorf("invalid object id %q", objectID)
	}
	hostName := parsedURL.Hostname()
	if strings.Contains(hostName, ":") {
		return fmt.Errorf("Compact Identifier-based DRS URI not implemented")
	}
	requestURI := fmt.Sprintf("https://%s/ga4gh/drs/v1/objects/%s", hostName, objectID)
	if d.insecureDirDomain == hostName {
		requestURI = fmt.Sprintf("http://%s/ga4gh/drs/v1/objects/%s", hostName, objectID)
	}
	var req *http.Request
	if len(d.aaiPassport) != 0 {
		req, err = http.NewRequest(http.MethodPost, requestURI, nil)
		if err != nil {
			return err
		}
		req.Header.Set("passports", d.aaiPassport)
	} else {
		req, err = http.NewRequest(http.MethodGet, requestURI, nil)
		if err != nil {
			return err
		}
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("DRS GetObject got status code: %d", resp.StatusCode)
	}

	var drsResp GetObjectResponse
	err = json.NewDecoder(resp.Body).Decode(&drsResp)
	if err != nil {
		return err
	}

	if err := d.pickAvailableTransputAndDownload(ctx, drsResp.AccessMethods, hostName, objectID, local); err != nil {
		return err
	}

	stat, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("failed to stat download file of path %s: %w", local, err)
	}

	if uint64(stat.Size()) != uint64(drsResp.Size) {
		return fmt.Errorf("file size not match")
	}

	checker := d.pickAvailableChecker(drsResp.Checksums)
	// skip to do check sum if checker is nil
	if checker == nil {
		return nil
	}

	check, err := checker.Check(local)
	if err != nil {
		return fmt.Errorf("checker error:%w", err)
	}
	if !check {
		return fmt.Errorf("checksum not match")
	}

	return nil
}

func (d *drsTransput) pickAvailableTransputAndDownload(ctx context.Context, accessMethods []AccessMethod, hostName, objectID string, local string) error {
	if len(accessMethods) == 0 {
		return fmt.Errorf("no access_methods in the drs object")
	}

	for _, accessMethod := range accessMethods {
		accessType := accessMethod.Type
		if accessType == "https" {
			accessURL, err := d.getAccessURL(accessMethod, hostName, objectID)
			if err != nil {
				d.logger.Warnf("No available access url of http access_method")
				continue
			}
			trans, err := transputhttp.NewHTTPTransput(
				&transputhttp.Config{
					Headers: accessURL.Headers,
				},
			)
			if err != nil {
				return err
			}
			return trans.DownloadFile(ctx, local, accessURL.URL)
		}
	}

	return fmt.Errorf("can not get suitable transput of drs")
}

func (d *drsTransput) pickAvailableChecker(checksums []Checksum) checker.Checker {
	if len(checksums) == 0 {
		return nil
	}

	for _, checksumObj := range checksums {
		if checksumObj.Type == consts.CheckerTypeMD5 {
			return md5checker.NewMD5Checker(checksumObj.Checksum)
		}
	}

	d.logger.Warnf("no available checksum checker, skip")
	return nil
}

func (d *drsTransput) getAccessURL(am AccessMethod, hostName, objectID string) (AccessURL, error) {
	if am.AccessURL.URL != "" {
		return am.AccessURL, nil
	}

	accessID := am.AccessID

	requestURI := fmt.Sprintf("https://%s/ga4gh/drs/v1/objects/%s/access/%s", hostName, objectID, accessID)
	if d.insecureDirDomain == hostName {
		requestURI = fmt.Sprintf("http://%s/ga4gh/drs/v1/objects/%s/access/%s", hostName, objectID, accessID)
	}

	resp, err := http.Get(requestURI)
	if err != nil {
		d.logger.Errorf("DRS GetAccess error: %v", err)
		return AccessURL{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		d.logger.Errorf("DRS GetAccess got status code: %d", resp.StatusCode)
		return AccessURL{}, fmt.Errorf("DRS GetAccess got status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var getAccessResp GetAccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAccessResp); err != nil {
		return AccessURL{}, err
	}

	return getAccessResp.AccessURL, nil
}
