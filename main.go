package main

import (
	"gin-gorm-OJ/router"
)

func main() {
	r := router.Router()

	r.Run(":8080")
}
