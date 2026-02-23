package mapconverter

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func ConvertMap(in map[string]interface{}) (map[string]*anypb.Any, error) {
	out := make(map[string]*anypb.Any, len(in))
	for k, v := range in {
		msg, ok := v.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("Unable to convert value %q to proto.Message", v)
		}
		anyVal, err := anypb.New(msg)
		if err != nil {
			return nil, err
		}
		out[k] = anyVal
	}
	return out, nil
}
