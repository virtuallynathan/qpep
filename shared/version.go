package shared

import "fmt"

const (
	MAJOR_VERSION = 0
	MINOR_VERSION = 0
	PATCH_VERSION = 1
)

var strVersion string

func init() {
	strVersion = fmt.Sprintf("%d.%d.%d", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)
}

func Version() string {
	return strVersion
}
