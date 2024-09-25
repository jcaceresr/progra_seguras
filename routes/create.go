package routes

import (
	getcollection "api/Collection"
	database "api/databases"
	model "api/model"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
)

// crea una esstructura depositMessage que sera utilizada mas adelante
type DepositMessage struct {
	Nro_cliente string `json:"nro_cliente"`
	Monto       string `json:"monto"`
	Divisa      string `json:"divisa"`
	Tipo        string `json:"tipo"`
}

type TransferenciaMessage struct {
	Nro_cliente_origen  string `json:"nro_cliente_origen"`
	Nro_cliente_destino string `json:"nro_cliente_destino"`
	Monto               string `json:"monto"`
	Divisa              string `json:"divisa"`
	Tipo                string `json:"tipo"`
}

type GiroMessage struct {
	Nro_cliente string `json:"nro_cliente"`
	Monto       string `json:"monto"`
	Divisa      string `json:"divisa"`
	Tipo        string `json:"tipo"`
}

func IniciarSesion(c *gin.Context) {
	var DB = database.ConnectDB()
	var clienteCollection = getcollection.GetCollectionClientes(DB, "Clientes")

	// Obtener el número de identificación y la contraseña del cuerpo de la solicitud
	type RequestBody struct {
		NumeroIdentificacion string `json:"numero_identificacion"`
		Contrasena           string `json:"contrasena"`
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
		c.JSON(http.StatusUnauthorized, gin.H{"estado": "no_exitoso"})
		return
	}

	// Validar la contraseña del cliente
	if reqBody.Contrasena != cliente.Contrasena {
		c.JSON(http.StatusUnauthorized, gin.H{"estado": "no_exitoso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"estado": "exitoso"})
}

func Deposito(c *gin.Context) {
	var DB = database.ConnectDB()
	var postCollection = getcollection.GetCollectionClientes(DB, "Clientes")
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener datos de la solicitud de depósito
	var depositoData model.Deposito
	if err := c.BindJSON(&depositoData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		log.Fatal(err)
		return
	}

	// Verificar si el cliente existe en la base de datos
	cliente := model.Cliente{}
	err := postCollection.FindOne(ctx, bson.M{"numero_identificacion": depositoData.Nro_cliente}).Decode(&cliente)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "cliente_no_encontrado"})
		return
	}

	// Verificar si la billetera del cliente existe y está activa
	billetera := model.Billetera{}
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": depositoData.Nro_cliente, "activo": true}).Decode(&billetera)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "billetera_no_encontrada"})
		//imprime si llega aca

		return
	}

	// Procesar el depósito en la billetera
	// Aquí realizar las operaciones necesarias para actualizar el saldo de la billetera, registrar el depósito, etc.

	// Enviar mensaje a RabbitMQ
	// Configurar la conexión a RabbitMQ

	rabbitmqURL := "amqp://guest:guest@localhost:5672/" // URL de conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Crear un canal en la conexión
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Declarar la cola en RabbitMQ
	queueName := "deposito_queue" // Nombre de la cola en RabbitMQ
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Crear el mensaje de depósito
	depositMessage := DepositMessage{
		Nro_cliente: depositoData.Nro_cliente,
		Monto:       depositoData.Monto,
		Divisa:      depositoData.Divisa,
		Tipo:        "Deposito",
	}

	messageBody, err := json.Marshal(depositMessage)
	if err != nil {
		log.Fatal(err)
	}

	// Publicar el mensaje en la cola
	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{"estado": "deposito_enviado"})
}

func Transferencia(c *gin.Context) {

	var DB = database.ConnectDB()
	var postCollection = getcollection.GetCollectionClientes(DB, "Clientes")
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener datos de la solicitud de transferencia
	var transferenciaData model.Transferencia
	if err := c.BindJSON(&transferenciaData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		log.Fatal(err)
		return
	}

	// Verificar si el cliente origen existe en la base de datos
	clienteOrigen := model.Cliente{}
	err := postCollection.FindOne(ctx, bson.M{"numero_identificacion": transferenciaData.Nro_cliente_origen}).Decode(&clienteOrigen)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "cliente_origen_no_encontrado"})
		return
	}

	// Verificar si el cliente destino existe en la base de datos
	clienteDestino := model.Cliente{}
	err = postCollection.FindOne(ctx, bson.M{"numero_identificacion": transferenciaData.Nro_cliente_destino}).Decode(&clienteDestino)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "cliente_destino_no_encontrado"})
		return
	}

	// Verificar si la billetera del cliente origen existe y está activa
	billeteraOrigen := model.Billetera{}
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": transferenciaData.Nro_cliente_origen, "activo": true}).Decode(&billeteraOrigen)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "billetera_origen_no_encontrada"})
		return
	}

	// Verificar si la billetera del cliente destino existe y está activa
	billeteraDestino := model.Billetera{}
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": transferenciaData.Nro_cliente_destino, "activo": true}).Decode(&billeteraDestino)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "billetera_destino_no_encontrada"})
		return
	}

	// Procesar la transferencia
	// Aquí realizar las operaciones necesarias para actualizar el saldo de las billeteras, registrar la transferencia, etc.

	// Enviar mensaje a RabbitMQ
	// Configurar la conexión a RabbitMQ
	rabbitmqURL := "amqp://guest:guest@localhost:5672/" // URL de conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Crear un canal en la conexión
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Declarar la cola en RabbitMQ
	queueName := "transferencia_queue" // Nombre de la cola en RabbitMQ
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Crear el mensaje de transferencia
	transferenciaMessage := TransferenciaMessage{
		Nro_cliente_origen:  transferenciaData.Nro_cliente_origen,
		Nro_cliente_destino: transferenciaData.Nro_cliente_destino,
		Monto:               transferenciaData.Monto,
		Divisa:              transferenciaData.Divisa,
		Tipo:                "Transferencia",
	}

	messageBody, err := json.Marshal(transferenciaMessage)
	if err != nil {
		log.Fatal(err)
	}

	// Publicar el mensaje en la cola
	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {

		log.Fatal(err)
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{"estado": "transferencia_enviada"})

}

func Giro(c *gin.Context) {

	var DB = database.ConnectDB()
	var postCollection = getcollection.GetCollectionClientes(DB, "Clientes")
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener datos de la solicitud de giro
	var giroData model.GiroRequest
	if err := c.BindJSON(&giroData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		log.Fatal(err)
		return
	}

	// Verificar si el cliente existe en la base de datos
	cliente := model.Cliente{}
	err := postCollection.FindOne(ctx, bson.M{"numero_identificacion": giroData.NroCliente}).Decode(&cliente)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "cliente_no_encontrado"})
		return
	}

	// Verificar si la billetera del cliente existe y está activa
	billetera := model.Billetera{}
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": giroData.NroCliente, "activo": true}).Decode(&billetera)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"estado": "billetera_no_encontrada"})
		return
	}

	// Verificar si el saldo es suficiente para realizar el giro
	var monto, _ = strconv.ParseFloat(giroData.Monto, 64)
	var saldo, _ = strconv.ParseFloat(billetera.Saldo, 64)

	if monto > saldo {
		c.JSON(http.StatusNotFound, gin.H{"estado": "saldo_insuficiente"})
		return
	}

	// Enviar mensaje a RabbitMQ
	// Configurar la conexión a RabbitMQ
	rabbitmqURL := "amqp://guest:guest@localhost:5672/" // URL de conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Crear un canal en la conexión
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// Declarar la cola en RabbitMQ
	queueName := "giro_queue" // Nombre de la cola en RabbitMQ
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Crear el mensaje de giro
	giroMessage := GiroMessage{
		Nro_cliente: giroData.NroCliente,
		Monto:       giroData.Monto,
		Divisa:      giroData.Divisa,
		Tipo:        "Giro",
	}

	messageBody, err := json.Marshal(giroMessage)
	if err != nil {
		log.Fatal(err)
	}

	// Publicar el mensaje en la cola
	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{"estado": "giro_enviado"})

}
