package main

import (
	"flag"
	"fraktal/cli"
	"fraktal/crypto"
	"fraktal/vfs"

	"github.com/go-redis/redis/v8"
)

func main() {
	persist := flag.String("secret", "", "Secret key for decrypting remote persistance secrets")
	flag.Parse()

	var rdb *redis.Client

	if *persist != "" {
		encryptedAddr := "15b4c2a57754719b6ea532e18e84dfb8dec2242038cec22f40c0d5870bf18404cb2df7eac49112c9376bc52e465f"
		encryptedPwd := "fab110bd65f7f732f1545eb8c5cbca582f3fa189539fec67f109475562ec38bb30174bc40fececa1e01eb00a522c68b5c34ecdd309c887e982338d2d4c73192afa93f118a3b2d22703544d30c99f0db5eedd7bf6f7ae43ad14ca5685"

		addr := crypto.Decrypt(*persist, encryptedAddr)
		pwd := crypto.Decrypt(*persist, encryptedPwd)

		// Connect to remote redis server
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: pwd,
			DB:       0,
		})
	}

	vfs := vfs.NewVirtualFS(rdb)
	cli.StartCLI(vfs)
}
