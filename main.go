package main

import (
	"github.com/BSFishy/lumos/router"
)

func main() {
	SetupLogger()
	SetupMqtt()

	r := router.NewRouter()

	r.ListenAndServe(":8080")
}
