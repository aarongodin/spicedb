package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type caveatContext map[string]any

func (cc *caveatContext) Scan(val any) error {
	v, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", v)
	}
	return json.Unmarshal(v, &cc)
}

func (cc *caveatContext) Value() (driver.Value, error) {
	return json.Marshal(&cc)
}
