package packageapi

import (
	"encoding/json"
	"sync"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
)

// ListResponse is used for the /list endpoint
type ListResponse struct {
	ArchList map[string]ArchRecord `json:"archs,omitempty"`
	Error    string                `json:"error,omitempty"`

	mtx sync.RWMutex
}

// ArchRecord holds the list of gapps Variants per API and its date
type ArchRecord struct {
	APIList map[string]APIRecord `json:"apis"`
	Date    string               `json:"date"`
}

// APIRecord holds all of the gapps Variants
type APIRecord struct {
	VariantList []string `json:"variants"`
}

// ToJSON forms JSON body from a struct, ignoring Marshal error
func (r *ListResponse) ToJSON() []byte {
	r.mtx.RLock()
	body, _ := json.Marshal(r)
	r.mtx.RUnlock()
	return body
}

// AddPackage safely adds the package to the ListResponse
func (r *ListResponse) AddPackage(date string, p gapps.Platform, a gapps.Android, v gapps.Variant) {
	r.mtx.Lock()

	if r.ArchList == nil {
		r.ArchList = make(map[string]ArchRecord)
	}
	if _, ok := r.ArchList[p.String()]; !ok {
		r.ArchList[p.String()] = ArchRecord{Date: date}
	}
	archRecord := r.ArchList[p.String()]

	if archRecord.APIList == nil {
		archRecord.APIList = make(map[string]APIRecord)
	}
	if _, ok := archRecord.APIList[a.HumanString()]; !ok {
		archRecord.APIList[a.HumanString()] = APIRecord{}
	}
	apiRecord := archRecord.APIList[a.HumanString()]

	apiRecord.VariantList = append(apiRecord.VariantList, v.String())
	archRecord.APIList[a.HumanString()] = apiRecord
	r.ArchList[p.String()] = archRecord

	r.mtx.Unlock()
}
