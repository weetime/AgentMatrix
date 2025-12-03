package kit

import (
	"strings"

	"github.com/google/uuid"
)

func GeneratorUUID(arg ...string) uuid.UUID {
	if len(arg) == 0 {
		return uuid.New()
	}
	return uuid.NewMD5(uuid.NameSpaceDNS, []byte(strings.Join(arg, "")))
}
