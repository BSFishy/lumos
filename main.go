package main

import (
	"github.com/BSFishy/lumos/router"
)

func main() {
	SetupLogger()

	client := SetupMqtt()
	setupGroups(client)

	r := router.NewRouter()

	r.ListenAndServe(":8080")
}
