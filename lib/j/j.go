package j

import (
	"encoding/json"
	"os"

	"github.com/saushew/users-balance-service/app/apiserver"
)

// ParseFile ...
func ParseFile(path string, config *apiserver.Config) error {

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, config); err != nil {
		return err
	}
	
	return nil
}
