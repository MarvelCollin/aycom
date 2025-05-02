//go:build tools

// This file ensures build and codegen tools and indirect dependencies are tracked in go.mod.
package tools

import (
	_ "github.com/google/uuid"
	_ "golang.org/x/crypto"
	_ "google.golang.org/grpc"
	_ "google.golang.org/protobuf"
	_ "gorm.io/gorm"
)
