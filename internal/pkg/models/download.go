package models

import (
	"encoding/json"
	"sync"

	"github.com/opengapps/package-api/internal/pkg/link"
)

const fieldError = "Error"

// DownloadResponse is used for the /download endpoint
type DownloadResponse struct {
	ZIP          string `json:"zip,omitempty"`
	ZIPMirrors   string `json:"zip_mirrors,omitempty"`
	MD5          string `json:"md5,omitempty"`
	VersionInfo  string `json:"version_info,omitempty"`
	SourceReport string `json:"source_report,omitempty"`
	Error        string `json:"error,omitempty"`

	mtx sync.RWMutex
}

// ToJSON forms JSON body from a struct, ignoring Marshal error
func (r *DownloadResponse) ToJSON() []byte {
	r.mtx.RLock()
	var body []byte
	if r.HasCriticalError() {
		body, _ = json.Marshal(DownloadResponse{Error: r.Error})
	} else {
		body, _ = json.Marshal(r)
	}
	r.mtx.RUnlock()
	return body
}

// HasCriticalError reports if we had an error which lead to the missing link
func (r *DownloadResponse) HasCriticalError() bool {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	if r.Error != "" && (r.ZIP == "" || r.MD5 == "" || r.VersionInfo == "" || r.SourceReport == "") {
		return true
	}
	return false
}

// SetField safely sets the field value
func (r *DownloadResponse) SetField(field, value string) {
	r.mtx.Lock()
	switch field {
	case link.FieldZIP:
		r.ZIP = value
	case link.FieldZIPMirrors:
		r.ZIPMirrors = value
	case link.FieldMD5:
		r.MD5 = value
	case link.FieldVersionInfo:
		r.VersionInfo = value
	case link.FieldSourceReport:
		r.SourceReport = value
	case fieldError:
		r.Error = value
	default:
	}
	r.mtx.Unlock()
}
