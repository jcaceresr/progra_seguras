package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	getcollection "api/Collection"
	database "api/databases"

	model "api/model"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
)

type DepositMessage struct {
	Nro_cliente string `json:"nro_cliente"`
	Monto       string `json:"monto"`
	Divisa      string `json:"divisa"`
	Tipo        string `json:"tipo"`
}

type TransferenciaMessage struct {
	nro_cliente_origen  string `json:"nro_cliente_origen"`
	nro_cliente_destino string `json:"nro_cliente_destino"`
	Monto               string `json:"monto"`
	Divisa              string `json:"divisa"`
	Tipo                string `json:"tipo"`
}

type GiroMessage struct {
	nro_cliente string `json:"nro_cliente"`
	Monto       string `json:"monto"`
	Divisa      string `json:"divisa"`
	Tipo        string `json:"tipo"`
}

func main() {
	// Establecer conexión con RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("No se pudo conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Crear un canal
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("No se pudo abrir un canal: %v", err)
	}
	defer ch.Close()

	// Declarar la cola de depósitos
	depositosQueue, err := ch.QueueDeclare(
		"deposito_queue", // Nombre de la cola
		true,             // Durabilidad de la cola
		false,            // Eliminar cola cuando no hay consumidores
		false,            // Exclusividad de cola
		false,            // No esperar confirmación para mensajes
		nil,              // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola de depósitos: %v", err)
	}

	// Declarar la cola de transferencias
	transferenciasQueue, err := ch.QueueDeclare(
		"transferencias_queue", // Nombre de la cola
		false,                  // Durabilidad de la cola
		false,                  // Eliminar cola cuando no hay consumidores
		false,                  // Exclusividad de cola
		false,                  // No esperar confirmación para mensajes
		nil,                    // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola de transferencias: %v", err)
	}

	// Configurar el consumidor para la cola de depósitos
	depositosMsgs, err := ch.Consume(
		depositosQueue.Name, // Nombre de la cola
		"",                  // Etiqueta del consumidor
		true,                // Auto-acknowledge (confirmación automática)
		false,               // Exclusividad de consumidor
		false,               // No esperar confirmación para mensajes
		false,               // No-wait
		nil,                 // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo registrar el consumidor para la cola de depósitos: %v", err)
	}

	// Configurar el consumidor para la cola de transferencias
	transferenciasMsgs, err := ch.Consume(
		transferenciasQueue.Name, // Nombre de la cola
		"",                       // Etiqueta del consumidor
		true,                     // Auto-acknowledge (confirmación automática)
		false,                    // Exclusividad de consumidor
		false,                    // No esperar confirmación para mensajes
		false,                    // No-wait
		nil,                      // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo registrar el consumidor para la cola de transferencias: %v", err)
	}

	// Función para procesar mensajes de depósitos
	processDepositos := func(msgs <-chan amqp.Delivery) {
		for msg := range msgs {
			// Procesar el mensaje de depósito recibido
			fmt.Printf("Mensaje de depósito recibido: %s\n", msg.Body)

			var depositoMessage DepositMessage
			err := json.Unmarshal(msg.Body, &depositoMessage)
			if err != nil {
				log.Println("Error al decodificar el mensaje de depósito:", err)
			}

			var DB = database.ConnectDB()
			var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			//crear deposito
			var deposito DepositMessage
			deposito.Nro_cliente = depositoMessage.Nro_cliente
			deposito.Monto = depositoMessage.Monto
			deposito.Divisa = depositoMessage.Divisa
			deposito.Tipo = depositoMessage.Tipo

			//buscar billetera
			var billetera model.Billetera
			err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": deposito.Nro_cliente}).Decode(&billetera)
			if err != nil {
				log.Fatal(err)
			}

			//actualizar billetera
			var billeteraUpdate model.Billetera
			billeteraUpdate.Nro_cliente = billetera.Nro_cliente
			saldo, _ := strconv.ParseFloat(billetera.Saldo, 64)
			monto, _ := strconv.ParseFloat(deposito.Monto, 64)
			billeteraUpdate.Saldo = strconv.FormatFloat(saldo+monto, 'f', -1, 64)
			billeteraUpdate.Divisa = billetera.Divisa
			billeteraUpdate.Activo = billetera.Activo
			billeteraUpdate.Nombre = billetera.Nombre

			//actualizar billetera
			update := bson.M{
				"$set": billeteraUpdate,
			}
			_, err = billeteraCollection.UpdateOne(ctx, bson.M{"nro_cliente": billeteraUpdate.Nro_cliente}, update)
			if err != nil {
				log.Fatal(err)
			}

			// Realizar las operaciones necesarias con el mensaje de depósito (actualizar la base de datos, etc.)
			// ...

			// Confirmar el mensaje (acknowledge)
			msg.Ack(false)
		}
	}

	// Función para procesar mensajes de transferencias
	processTransferencias := func(msgs <-chan amqp.Delivery) {
		for msg := range msgs {
			// Procesar el mensaje de transferencia recibido
			fmt.Printf("Mensaje de transferencia recibido: %s\n", msg.Body)

			// Realizar las operaciones necesarias con el mensaje de transferencia (actualizar la base de datos, etc.)

			// ...

			// Confirmar el mensaje (acknowledge)
			msg.Ack(false)
		}
	}

	// Iniciar goroutines para procesar mensajes de depósitos y transferencias
	go processDepositos(depositosMsgs)
	go processTransferencias(transferenciasMsgs)

	// Mantener el programa en ejecución indefinidamente
	forever := make(chan bool)
	<-forever
}
