package main

import (
	_ "net/http/pprof"

	"github.com/MTudorrrr/distribution/registry"
	_ "github.com/MTudorrrr/distribution/registry/auth/htpasswd"
	_ "github.com/MTudorrrr/distribution/registry/auth/silly"
	_ "github.com/MTudorrrr/distribution/registry/auth/token"
	_ "github.com/MTudorrrr/distribution/registry/proxy"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/azure"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/filesystem"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/gcs"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/inmemory"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/middleware/redirect"
	_ "github.com/MTudorrrr/distribution/registry/storage/driver/s3-aws"
)

func main() {
	// NOTE(milosgajdos): if the only two commands registered
	// with registry.RootCmd fail they will halt the program
	// execution and  exit the program with non-zero exit code.
	// nolint:errcheck
	registry.RootCmd.Execute()
}
