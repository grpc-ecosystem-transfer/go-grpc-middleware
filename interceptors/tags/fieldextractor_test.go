// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tags_test

import (
	"testing"
	"time"

	pb_gogotestproto "github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/gogotestproto"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/v2/grpctesting/testproto"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeGenRequestLogFieldExtractor_ManualIsDeclared(t *testing.T) {
	req := &pb_testproto.PingRequest{Value: "my_value"}
	valMap := tags.CodeGenRequestFieldExtractor("", req)
	require.Len(t, valMap, 1, "PingRequest should have a ExtractLogFields method declared in test.manual_extractfields.pb")
	require.EqualValues(t, valMap, map[string]string{"value": "my_value"})
}

func TestTaggedRequestFiledExtractor_PingRequest(t *testing.T) {
	req := &pb_gogotestproto.PingRequest{
		Ping: &pb_gogotestproto.Ping{
			Id: &pb_gogotestproto.PingId{
				Id: 1337, // logfield is ping_id
			},
			Value: "something",
		},
		Meta: &pb_gogotestproto.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "1337", valMap["ping_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}

func TestTaggedRequestFiledExtractor_PongRequest(t *testing.T) {
	req := &pb_gogotestproto.PongRequest{
		Pong: &pb_gogotestproto.Pong{
			Id: "some_id",
		},
		Meta: &pb_gogotestproto.Metadata{
			Tags: []string{"tagone", "tagtwo"}, // logfield is meta_tags
		},
	}
	valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
	assert.EqualValues(t, "some_id", valMap["pong_id"])
	assert.EqualValues(t, "[tagone tagtwo]", valMap["meta_tags"])
}

// Test to ensure TagBasedRequestFieldExtractor does not panic when encountering private struct members such as
// when using gogoproto.stdtime which results in a time.Time that has private struct members
func TestTaggedRequestFiledExtractor_GogoTime(t *testing.T) {
	ts := time.Date(2010, 01, 01, 0, 0, 0, 0, time.UTC)
	req := &pb_gogotestproto.GoGoProtoStdTime{
		Timestamp: &ts,
	}
	assert.NotPanics(t, func() {
		valMap := tags.TagBasedRequestFieldExtractor("log_field")("", req)
		assert.Empty(t, valMap)
	})
}
