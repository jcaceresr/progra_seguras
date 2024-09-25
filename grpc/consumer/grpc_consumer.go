package main

import (
	"context"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"

	getcollection "api/Collection"
	database "api/databases"
	pb "api/grpc" // Importa el paquete generado a partir del archivo .proto

	model "api/model"
	"strconv"

	"time"
)

// Implementa el servidor de transferencias
type transferServer struct {
	pb.UnimplementedTransferServiceServer
}

// Implementa el método SendTransfer del servicio TransferService
func (s *transferServer) SendTransfer(ctx context.Context, req *pb.TransferMessage) (*pb.TransferResponse, error) {
	// Realiza las operaciones necesarias con los datos de la transferencia recibida
	// ...

	// Crea un movimiento

	//obtener id de la billetera del cliente

	var DB = database.ConnectDB()
	var billeteraCollection = getcollection.GetCollectionBilleteras(DB, "Billeteras")

	ctx1, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var billeteraOrigen model.Billetera

	billeteraCollection.FindOne(ctx1, bson.M{"nro_cliente": req.NroClienteOrigen}).Decode(&billeteraOrigen)

	var movimientos model.Movimiento
	movimientos.Nro_cliente = req.NroClienteOrigen
	var monto = strconv.FormatFloat(float64(req.Monto), 'f', -1, 64)
	movimientos.Monto = monto
	movimientos.Divisa = req.Divisa
	movimientos.Tipo = req.TipoOperacion
	movimientos.Fecha_hora = time.Now().Format("2006-01-02 15:04:05")
	movimientos.ID_billetera = billeteraOrigen.Id

	// Crea un nuevo movimiento en la base de datos
	var DB2 = database.ConnectDB()
	var movimientosCollection = getcollection.GetCollectionMovimientos(DB2, "Movimientos")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	_, err2 := movimientosCollection.InsertOne(ctx2, movimientos)
	if err2 != nil {
		log.Fatal(err2)
	}

	// Devuelve una respuesta de ejemplo
	resp := &pb.TransferResponse{
		Status: "Deposito realizado correctamente",
	}
	return resp, nil
}

func main() {
	// Configura el servidor gRPC
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crea una nueva instancia del servidor de transferencias
	s := grpc.NewServer()

	// Registra el servidor de transferencias en el servidor gRPC
	pb.RegisterTransferServiceServer(s, &transferServer{})

	// Inicia el servidor gRPC
	log.Println("Servidor gRPC iniciado en el puerto 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
