package main

import (
	routes "api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/api/cliente", routes.GetCliente)

	router.POST("/api/inicio_sesion", routes.IniciarSesion)

	router.POST("/api/deposito", routes.Deposito)

	router.POST("/api/transferencia", routes.Transferencia)

	router.POST("/api/giro", routes.Giro)

	router.Run("localhost: 5000")
}
