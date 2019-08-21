package packageapi

import (
	"fmt"
	"time"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
)

const (
	pkgTemplate          = "open_gapps-%s-%s-%s-%s"
	reportTemplate       = "sources_report-%s-%s-%s.txt"
	masterMirrorTemplate = "https://master.dl.sourceforge.net/project/opengapps/%s/%s/%s?ts=%d"
	mirrorsTemplate      = "https://sourceforge.net/settings/mirror_choices?projectname=opengapps&filename=%s/%s/%s"

	zipExtension = ".zip"
	md5Extension = ".zip.md5"
	logExtension = ".versionlog.txt"

	fieldZIP          = "ZIP"
	fieldZIPMirrors   = "ZIPMirrors"
	fieldMD5          = "MD5"
	fieldVersionInfo  = "VersionInfo"
	fieldSourceReport = "SourceReport"
	fieldError        = "Error"
)

var templateMap = map[string]string{
	fieldZIP:          pkgTemplate + zipExtension,
	fieldZIPMirrors:   pkgTemplate + zipExtension,
	fieldMD5:          pkgTemplate + md5Extension,
	fieldVersionInfo:  pkgTemplate + logExtension,
	fieldSourceReport: reportTemplate,
}

func formatLink(field, date string, p gapps.Platform, a gapps.Android, v gapps.Variant) string {
	now := time.Now().Unix()
	switch field {
	case fieldZIPMirrors:
		filename := fmt.Sprintf(templateMap[field], p, a.HumanString(), v, date)
		return fmt.Sprintf(mirrorsTemplate, p, date, filename)
	case fieldSourceReport:
		filename := fmt.Sprintf(templateMap[field], p, a.HumanString(), date)
		return fmt.Sprintf(masterMirrorTemplate, p, date, filename, now)
	default:
		filename := fmt.Sprintf(templateMap[field], p, a.HumanString(), v, date)
		return fmt.Sprintf(masterMirrorTemplate, p, date, filename, now)
	}
}
