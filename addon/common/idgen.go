// pkg/idgen/idgen.go
package common

import "github.com/gofrs/uuid"

func NewV7() string {
	id, _ := uuid.NewV7()
	return id.String() // 36 ký tự
}
