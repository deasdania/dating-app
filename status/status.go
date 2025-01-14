package status

import (
	"crypto/sha256"
	"fmt"
)

type DatingStatusCode string

const (
	EmptyCode               DatingStatusCode = ""
	genericIntegrationsCode string           = "DAPPSXXX5000"
	datingStatusCodeLen     int              = 11
)

var (
	datingMapInverse     = make(map[string]DatingStatusCode, len(datingMap))
	invalidArgsStatCodes []string
)

func init() {
	// as each description must be unique as per specs, then we are safe to assume there will be no collisions
	for k, v := range datingMap {
		hashedDesc := fmt.Sprintf("%x", sha256.Sum256([]byte(v.StatusDesc))) // prevent long irregular strings being set as a key to the map
		datingMapInverse[hashedDesc] = k
	}

	invalidArgsStatCodes = []string{
		"DAPPSXXX5005",
		"DAPPSXXX5006",
		"DAPPSXXX5025",
	}
}

// Error gives a specific http error for validation or authentication errors.
type Error struct {
	Err  error
	Code DatingStatusCode
}

// Error fulfils the error interface.
func (e Error) Error() string {
	return e.Err.Error() // retain err wording on response to retain backwards compatibility
}

func ResponseFromCode(code DatingStatusCode) StatusResponse {
	var res StatusResponse
	if len(code) == datingStatusCodeLen {
		if res2, ok := datingMap[code]; ok {
			res = res2
			res.StatusCode = string(code)
			return res
		}
	}
	res = datingMap[SystemErrCode_Generic]
	res.StatusCode = string(SystemErrCode_Generic)
	return res
}
