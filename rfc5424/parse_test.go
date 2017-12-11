package rfc5424

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestParser_Empty(c *C) {
	res, err := ParseStructuredData("-")
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{})
}

func (s *TestSuite) TestParser_SingleCustomParam(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"iut":         "3",
			"eventSource": "Application",
			"eventID":     "1011",
		},
	})
}

func (s *TestSuite) TestParser_MultipleCustomParam(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32472 iut="3"][exampleSDID@32473 eventID="1011"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32472": StructuredDataParam{
			"iut": "3",
		},
		"exampleSDID@32473": StructuredDataParam{
			"eventID": "1011",
		},
	})
}

func (s *TestSuite) TestParser_SpaceInParamValue(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 eventSource="My Application"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"eventSource": "My Application",
		},
	})
}

func (s *TestSuite) TestParser_EscapedSymbolInParamValue1(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 eventSource="My\"Application"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"eventSource": "My\"Application",
		},
	})
}

func (s *TestSuite) TestParser_EscapedSymbolInParamValue2(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 eventSource="My\]Application"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"eventSource": "My]Application",
		},
	})
}

func (s *TestSuite) TestParser_EscapedSymbolInParamValue3(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 eventSource="My\\Application"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"eventSource": "My\\Application",
		},
	})
}

func (s *TestSuite) TestParser_BackslashInParamValue(c *C) {
	res, err := ParseStructuredData(`[exampleSDID@32473 eventSource="My\[Application"]`)
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, StructuredData{
		"exampleSDID@32473": StructuredDataParam{
			"eventSource": "My\\[Application",
		},
	})
}

func (s *TestSuite) TestParser_ParamWithNoValue(c *C) {
	_, err := ParseStructuredData(`[exampleSDID@32473 eventSource]`)
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestParser_NoParam(c *C) {
	_, err := ParseStructuredData(`[exampleSDID@32473]`)
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestParser_NoID(c *C) {
	_, err := ParseStructuredData(`[eventSource="My Application"]`)
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestParser_NoID2(c *C) {
	_, err := ParseStructuredData(`[ eventSource="My Application"]`)
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestParser_NoSquareBracket(c *C) {
	_, err := ParseStructuredData(`eventSource="My Application"]`)
	c.Assert(err, NotNil)
}
