package agents

import "encoding/json"

func unmarshalCapabilities(data string) ([]string, error) {
	var caps []string
	if err := json.Unmarshal([]byte(data), &caps); err != nil {
		return nil, err
	}
	return caps, nil
}
