package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"github.com/opengapps/package-api/internal/pkg/link"
)

// Public consts
const (
	DateOnlyFormat    = "20060102"
	HumanDateTemplate = "%d %s %d"
)

// ListResponse is used for the /list endpoint
type ListResponse struct {
	ArchList map[string]ArchRecord `json:"archs,omitempty"`
	Error    string                `json:"error,omitempty"`

	mtx sync.RWMutex
}

// ArchRecord holds the list of gapps Variants per API and its date
type ArchRecord struct {
	APIList   map[string]APIRecord `json:"apis"`
	Date      string               `json:"date"`
	HumanDate string               `json:"human_date"`
}

// APIRecord holds all of the gapps Variants
type APIRecord struct {
	VariantList []APIVariant `json:"variants"`
}

// APIVariant describes gapps Variant
type APIVariant struct {
	Name         string `json:"name"`
	ZIP          string `json:"zip"`
	ZIPSize      int64  `json:"zip_size"`
	MD5          string `json:"md5"`
	VersionInfo  string `json:"version_info"`
	SourceReport string `json:"source_report"`
}

// ToJSON forms JSON body from a struct, ignoring Marshal error
func (r *ListResponse) ToJSON() []byte {
	r.mtx.RLock()
	body, _ := json.Marshal(r)
	r.mtx.RUnlock()
	return body
}

// AddPackage safely adds the package to the ListResponse
func (r *ListResponse) AddPackage(date string, p gapps.Platform, a gapps.Android, v gapps.Variant) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// parse date
	dt, err := time.Parse(DateOnlyFormat, date)
	if err != nil {
		return fmt.Errorf("unable to parse date: %w", err)
	}
	humandate := fmt.Sprintf(HumanDateTemplate, dt.Day(), dt.Month(), dt.Year())

	if r.ArchList == nil {
		r.ArchList = make(map[string]ArchRecord)
	}
	if _, ok := r.ArchList[p.String()]; !ok {
		r.ArchList[p.String()] = ArchRecord{
			Date:      date,
			HumanDate: humandate,
		}
	}
	archRecord := r.ArchList[p.String()]

	if archRecord.APIList == nil {
		archRecord.APIList = make(map[string]APIRecord)
	}
	if _, ok := archRecord.APIList[a.HumanString()]; !ok {
		archRecord.APIList[a.HumanString()] = APIRecord{}
	}
	apiRecord := archRecord.APIList[a.HumanString()]

	apiRecord.VariantList = append(apiRecord.VariantList, newAPIVariant(date, p, a, v))
	archRecord.APIList[a.HumanString()] = apiRecord
	r.ArchList[p.String()] = archRecord

	return nil
}

func newAPIVariant(date string, p gapps.Platform, a gapps.Android, v gapps.Variant) APIVariant {
	return APIVariant{
		Name:         v.String(),
		ZIP:          link.New(link.FieldZIP, date, p, a, v),
		MD5:          link.New(link.FieldMD5, date, p, a, v),
		VersionInfo:  link.New(link.FieldVersionInfo, date, p, a, v),
		SourceReport: link.New(link.FieldSourceReport, date, p, a, v),
	}
}
