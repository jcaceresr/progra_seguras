package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	getcollection "api/Collection"
	database "api/databases"

	model "api/model"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"

	pb "api/grpc"
	// Importa el paquete generado por el compilador de protobuf
)

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

func consumeMessages(ch *amqp.Channel, queueName string) {
	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	for msg := range msgs {
		fmt.Println("Mensaje recibido:", string(msg.Body))
		mongo(queueName, msg.Body)
	}
}

func main() {
	rabbitmq()

	// Bloquear el hilo principal para que la función nunca termine
	select {}
}

func mongo(queueName string, msg []byte) {
	switch queueName {

	case "deposito_queue":
		deposito(msg)

	case "transferencia_queue":
		transferencia(msg)

	case "giro_queue":
		giro(msg)
	}
}

func rabbitmq() {
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
	deposito, err := ch.QueueDeclare(
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
	transferencia, err := ch.QueueDeclare(
		"transferencia_queue", // Nombre de la cola
		true,                  // Durabilidad de la cola
		false,                 // Eliminar cola cuando no hay consumidores
		false,                 // Exclusividad de cola
		false,                 // No esperar confirmación para mensajes
		nil,                   // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola de transferencias: %v", err)
	}

	// Declarar la cola de giros
	giro, err := ch.QueueDeclare(
		"giro_queue", // Nombre de la cola
		true,         // Durabilidad de la cola
		false,        // Eliminar cola cuando no hay consumidores
		false,        // Exclusividad de cola
		false,        // No esperar confirmación para mensajes
		nil,          // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola de giros: %v", err)
	}

	queueNames := []string{
		deposito.Name,
		transferencia.Name,
		giro.Name,
	}

	var wg sync.WaitGroup
	wg.Add(len(queueNames))

	for _, queueName := range queueNames {
		go func(name string) {
			defer wg.Done()
			consumeMessages(ch, name)
		}(queueName)
	}

	wg.Wait()

}
func deposito(msg []byte) {
	var depositoMessage DepositMessage
	err := json.Unmarshal(msg, &depositoMessage)
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

	// Conexión con el servidor gRPC
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer connection.Close()

	// Crear un cliente para el servicio de deposito
	client := pb.NewTransferServiceClient(connection)

	valor, _ := strconv.ParseFloat(deposito.Monto, 32)

	// Crear el mensaje de deposito
	transferencia := &pb.TransferMessage{
		NroClienteOrigen:  deposito.Nro_cliente,
		NroClienteDestino: deposito.Nro_cliente,
		Monto:             float32(valor),
		Divisa:            "USD",
		TipoOperacion:     "Deposito",
	}

	// Enviar la solicitud de deposito al servidor gRPC
	resp, err := client.SendTransfer(context.Background(), transferencia)
	if err != nil {
		log.Fatalf("Failed to send transfer request: %v", err)
	}

	log.Printf("Transfer request sent. Response received: %+v", resp)

}

func transferencia(msg []byte) {

	var transferenciaMessage TransferenciaMessage
	err := json.Unmarshal(msg, &transferenciaMessage)
	if err != nil {
		log.Println("Error al decodificar el mensaje de transferencia:", err)

	}

	var DB = database.ConnectDB()
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//crear transferencia
	var transferencia TransferenciaMessage
	transferencia.Nro_cliente_origen = transferenciaMessage.Nro_cliente_origen
	transferencia.Nro_cliente_destino = transferenciaMessage.Nro_cliente_destino
	transferencia.Monto = transferenciaMessage.Monto
	transferencia.Divisa = transferenciaMessage.Divisa
	transferencia.Tipo = transferenciaMessage.Tipo

	//buscar billetera origen

	var billeteraOrigen model.Billetera
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": transferencia.Nro_cliente_origen}).Decode(&billeteraOrigen)
	if err != nil {
		log.Fatal(err)
	}

	//buscar billetera destino
	var billeteraDestino model.Billetera
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": transferencia.Nro_cliente_destino}).Decode(&billeteraDestino)
	if err != nil {
		log.Fatal(err)
	}

	//actualizar billetera origen

	var billeteraUpdateOrigen model.Billetera
	billeteraUpdateOrigen.Nro_cliente = billeteraOrigen.Nro_cliente
	saldoOrigen, _ := strconv.ParseFloat(billeteraOrigen.Saldo, 64)
	montoOrigen, _ := strconv.ParseFloat(transferencia.Monto, 64)
	billeteraUpdateOrigen.Saldo = strconv.FormatFloat(saldoOrigen-montoOrigen, 'f', -1, 64)
	billeteraUpdateOrigen.Divisa = billeteraOrigen.Divisa
	billeteraUpdateOrigen.Activo = billeteraOrigen.Activo
	billeteraUpdateOrigen.Nombre = billeteraOrigen.Nombre

	//actualizar billetera origen
	updateOrigen := bson.M{
		"$set": billeteraUpdateOrigen,
	}
	_, err = billeteraCollection.UpdateOne(ctx, bson.M{"nro_cliente": billeteraUpdateOrigen.Nro_cliente}, updateOrigen)
	if err != nil {
		log.Fatal(err)
	}

	//actualizar billetera destino

	var billeteraUpdateDestino model.Billetera
	billeteraUpdateDestino.Nro_cliente = billeteraDestino.Nro_cliente
	saldoDestino, _ := strconv.ParseFloat(billeteraDestino.Saldo, 64)
	montoDestino, _ := strconv.ParseFloat(transferencia.Monto, 64)
	billeteraUpdateDestino.Saldo = strconv.FormatFloat(saldoDestino+montoDestino, 'f', -1, 64)
	billeteraUpdateDestino.Divisa = billeteraDestino.Divisa
	billeteraUpdateDestino.Activo = billeteraDestino.Activo
	billeteraUpdateDestino.Nombre = billeteraDestino.Nombre

	//actualizar billetera destino
	updateDestino := bson.M{
		"$set": billeteraUpdateDestino,
	}
	_, err = billeteraCollection.UpdateOne(ctx, bson.M{"nro_cliente": billeteraUpdateDestino.Nro_cliente}, updateDestino)
	if err != nil {
		log.Fatal(err)
	}

	// Conexión con el servidor gRPC
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer connection.Close()

	// Crear un cliente para el servicio de transferencia
	client := pb.NewTransferServiceClient(connection)

	valor, _ := strconv.ParseFloat(transferencia.Monto, 32)

	// Crear el mensaje de transferencia
	transferenciaGrpc := &pb.TransferMessage{
		NroClienteOrigen:  transferencia.Nro_cliente_origen,
		NroClienteDestino: transferencia.Nro_cliente_destino,
		Monto:             float32(valor),
		Divisa:            "USD",
		TipoOperacion:     "Transferencia",
	}

	// Enviar la solicitud de transferencia al servidor gRPC
	resp, err := client.SendTransfer(context.Background(), transferenciaGrpc)
	if err != nil {
		log.Fatalf("Failed to send transfer request: %v", err)
	}

	log.Printf("Transfer request sent. Response received: %+v", resp)
}

func giro(msg []byte) {

	var giroMessage GiroMessage
	err := json.Unmarshal(msg, &giroMessage)
	if err != nil {
		log.Println("Error al decodificar el mensaje de depósito:", err)
	}

	var DB = database.ConnectDB()
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//crear giro
	var giro GiroMessage
	giro.Nro_cliente = giroMessage.Nro_cliente
	giro.Monto = giroMessage.Monto
	giro.Divisa = giroMessage.Divisa
	giro.Tipo = giroMessage.Tipo

	//buscar billetera
	var billetera model.Billetera
	err = billeteraCollection.FindOne(ctx, bson.M{"nro_cliente": giro.Nro_cliente}).Decode(&billetera)
	if err != nil {
		log.Fatal(err)
	}

	//actualizar billetera
	var billeteraUpdate model.Billetera
	billeteraUpdate.Nro_cliente = billetera.Nro_cliente
	saldo, _ := strconv.ParseFloat(billetera.Saldo, 64)
	monto, _ := strconv.ParseFloat(giro.Monto, 64)
	billeteraUpdate.Saldo = strconv.FormatFloat(saldo-monto, 'f', -1, 64)
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

	// Conexión con el servidor gRPC
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer connection.Close()

	// Crear un cliente para el servicio de giro
	client := pb.NewTransferServiceClient(connection)

	valor, _ := strconv.ParseFloat(giro.Monto, 32)

	// Crear el mensaje de giro
	giroGrpc := &pb.TransferMessage{
		NroClienteOrigen:  giro.Nro_cliente,
		NroClienteDestino: giro.Nro_cliente,
		Monto:             float32(valor),
		Divisa:            "USD",
		TipoOperacion:     "Giro",
	}

	// Enviar la solicitud de giro al servidor gRPC
	resp, err := client.SendTransfer(context.Background(), giroGrpc)
	if err != nil {
		log.Fatalf("Failed to send transfer request: %v", err)
	}

	log.Printf("Transfer request sent. Response received: %+v", resp)

}
