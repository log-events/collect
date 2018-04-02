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
	res, id := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"s2": "v2",
	}, map[string]interface{}{
		"r1": "s2",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "v2",
	})
	c.Assert(id, Equals, "")
}

func (s *TestSuite) TestParser_PlainNotExist(c *C) {
	res, _ := getDocumentFromLogParts(map[string]interface{}{
		"s1": "v1",
		"s2": "v2",
	}, map[string]interface{}{
		"r1": "s3",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{})
}

func (s *TestSuite) TestParser_PlainTypecast(c *C) {
	res, _ := getDocumentFromLogParts(map[string]interface{}{
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
	res, _ := getDocumentFromLogParts(map[string]interface{}{
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
	res, _ := getDocumentFromLogParts(map[string]interface{}{
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

func (s *TestSuite) TestParser_IDGenerationWithoutSeq(c *C) {
	timestamp, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	res, id := getDocumentFromLogParts(map[string]interface{}{
		"s1":        "v1",
		"timestamp": timestamp,
		"s2":        "v2",
	}, map[string]interface{}{
		"r1": "s1",
		"r2": "s2",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "v1",
		"r2": "v2",
	})
	c.Assert(id, Equals, "11361890450000000000000000")
}

func (s *TestSuite) TestParser_IDGenerationWithSeq(c *C) {
	timestamp, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05.000003+07:00")
	res, id := getDocumentFromLogParts(map[string]interface{}{
		"timestamp": timestamp,
		"structured_data": map[string]interface{}{
			"meta": map[string]interface{}{
				"sequenceId": "2147483647",
			},
		},
	}, map[string]interface{}{
		"r1": "timestamp",
		"r2": "id",
	})
	c.Assert(res, DeepEquals, map[string]interface{}{
		"r1": "2006-01-02T08:04:05.000003Z",
		"r2": "11361890450000032147483647",
	})
	c.Assert(id, Equals, "11361890450000032147483647")
}
