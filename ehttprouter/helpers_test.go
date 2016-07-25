package ehttprouter

import (
	"encoding/json"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/creack/ehttp"
)

func getCallstack(skip int) (string, string, int) {
	var name string
	pc, file, line, ok := runtime.Caller(1 + skip)
	if !ok {
		name, file, line = "<unkown>", "<unknown>", -1
	} else {
		name = runtime.FuncForPC(pc).Name()
		name = path.Base(name)
		file = path.Base(file)
	}
	return name, file, line
}

func assertInt(t *testing.T, expect, got int) {
	_, file, line := getCallstack(1)
	if expect != got {
		t.Errorf("[%s:%d] Unexpected result.\nExpect:\t%d\nGot:\t%d\n", file, line, expect, got)
	}
}

func assertString(t *testing.T, expect, got string) {
	_, file, line := getCallstack(1)
	expect, got = strings.TrimSpace(expect), strings.TrimSpace(got)
	if expect != got {
		t.Errorf("[%s:%d] Unexpected result.\nExpect:\t%s\nGot:\t%s\n", file, line, expect, got)
	}
}

func assertJSONError(t *testing.T, expect, got string) {
	_, file, line := getCallstack(1)
	expect, got = strings.TrimSpace(expect), strings.TrimSpace(got)

	jErr := ehttp.JSONError{}
	if err := json.Unmarshal([]byte(got), &jErr); err != nil {
		t.Errorf("[%s:%d] Error parsing json error: %s\n", file, line, expect)
	}
	for _, errStr := range jErr.Errors {
		if errStr == expect {
			return
		}
	}
	t.Errorf("[%s:%d] Unexpected error.\nExpect:\t%s\nGot:\t%s\n", file, line, expect, got)
}
