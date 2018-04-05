package cmd

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestParser_Plain(c *C) {
	res := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"s2": "v2",
	}, map[string]interface{}{
		"r1": "s2",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "v2",
	})
}

func (s *TestSuite) TestParser_PlainNotExist(c *C) {
	res := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"s2": "v2",
	}, map[string]interface{}{
		"r1": "s3",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{})
}

func (s *TestSuite) TestParser_PlainTypecast(c *C) {
	res := getDocumentFromLogParts(map[string]interface{}{
		"s1": "42",
		"s2": "v2",
	}, map[string]interface{}{
		"r1": map[interface{}]interface{}{
			"type":  "int",
			"field": "s1",
		},
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": int64(42),
	})
}

func (s *TestSuite) TestParser_NestedTypecast(c *C) {
	res := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"sx": map[string]interface{}{
			"sy": map[string]interface{}{
				"sz": "42",
			},
		},
		"s2": "v2",
	}, map[string]interface{}{
		"r1": map[interface{}]interface{}{
			"type":  "int",
			"field": "sx.sy.sz",
		},
		"r2": "s2",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": int64(42),
		"r2": "v2",
	})
}

func (s *TestSuite) TestParser_Nested(c *C) {
	res := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"sx": map[string]interface{}{
			"sy": map[string]interface{}{
				"sz": "42",
			},
		},
		"s2": "v2",
	}, map[string]interface{}{
		"r1": "sx.sy.sz",
		"r2": "s2",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "42",
		"r2": "v2",
	})
}

func (s *TestSuite) TestParser_Timestamp(c *C) {
	timestamp, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05.000003+07:00")
	res := getDocumentFromLogParts(map[string]interface{}{
		"timestamp": timestamp,
		"structured_data": map[string]interface{}{
			"meta": map[string]interface{}{
				"sequenceId": "2147483647",
			},
		},
	}, map[string]interface{}{
		"r1": "timestampRFC3339",
		"r2": "timestampUnixNano",
		"r3": "timestampUnix",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "2006-01-02T08:04:05.000003Z",
		"r2": int64(1136189045000003000),
		"r3": int64(1136189045),
	})
}
