package routes

import (
	getcollection "api/Collection"
	database "api/databases"
	model "api/model"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetCliente(c *gin.Context) {
	var DB = database.ConnectDB()
	var clienteCollection = getcollection.GetCollectionClientes(DB, "Clientes")

	// Obtener el número de identificación del cliente del cuerpo de la solicitud
	type RequestBody struct {
		NumeroIdentificacion string `json:"numero_identificacion"`
	}

	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Buscar el cliente en la base de datos por número de identificación
	var cliente model.Cliente
	err := clienteCollection.FindOne(context.TODO(), bson.M{"numero_identificacion": reqBody.NumeroIdentificacion}).Decode(&cliente)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "cliente_no_encontrado"})
		return
	}

	// Eliminar la contraseña del cliente de la respuesta
	cliente.Contrasena = ""

	c.JSON(http.StatusOK, cliente)
}
