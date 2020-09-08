package models

import (
	"fmt"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
)

// Public consts
const (
	FieldZIP          = "ZIP"
	FieldZIPMirrors   = "ZIPMirrors"
	FieldMD5          = "MD5"
	FieldVersionInfo  = "VersionInfo"
	FieldSourceReport = "SourceReport"
	FieldError        = "Error"
)

const (
	downloadTemplate = "https://downloads.sourceforge.net/project/opengapps/%s/%s/%s"
	mirrorsTemplate  = "https://sourceforge.net/settings/mirror_choices?projectname=opengapps&filename=%s/%s/%s"
	pkgTemplate      = "open_gapps-%s-%s-%s-%s"
	reportTemplate   = "sources_report-%s-%s-%s.txt"

	zipExtension = ".zip"
	md5Extension = ".zip.md5"
	logExtension = ".versionlog.txt"
)

// TemplateMap holds format templates
var TemplateMap = map[string]string{
	FieldZIP:          pkgTemplate + zipExtension,
	FieldZIPMirrors:   pkgTemplate + zipExtension,
	FieldMD5:          pkgTemplate + md5Extension,
	FieldVersionInfo:  pkgTemplate + logExtension,
	FieldSourceReport: reportTemplate,
}

// NewDownloadLink returns a download link per provided parameters
func NewDownloadLink(field, date string, p gapps.Platform, a gapps.Android, v gapps.Variant) string {
	switch field {
	case FieldZIPMirrors:
		filename := fmt.Sprintf(TemplateMap[field], p, a.HumanString(), v, date)
		return fmt.Sprintf(mirrorsTemplate, p, date, filename)
	case FieldSourceReport:
		filename := fmt.Sprintf(TemplateMap[field], p, a.HumanString(), date)
		return fmt.Sprintf(downloadTemplate, p, date, filename)
	default:
		filename := fmt.Sprintf(TemplateMap[field], p, a.HumanString(), v, date)
		return fmt.Sprintf(downloadTemplate, p, date, filename)
	}
}
