package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Definición de las estructuras
type Cliente struct {
	Nombre                string `bson:"nombre"`
	Contrasena            string `bson:"contrasena"`
	Fecha_nacimiento      string `bson:"fecha_nacimiento"`
	Direccion             string `bson:"direccion"`
	Numero_identificacion string `bson:"numero_identificacion"`
	Email                 string `bson:"email"`
	Telefono              string `bson:"telefono"`
	Genero                string `bson:"genero"`
	Nacionalidad          string `bson:"nacionalidad"`
	Ocupacion             string `bson:"ocupacion"`
}

type Billetera struct {
	Id          string `bson:"_id,omitempty"`
	Nro_cliente string `bson:"nro_cliente"`
	Saldo       string `bson:"saldo"`
	Divisa      string `bson:"divisa"`
	Nombre      string `bson:"nombre"`
	Activo      bool   `bson:"activo"`
}

type Movimiento struct {
	Nro_cliente  string `bson:"nro_cliente"`
	Monto        string `bson:"monto"`
	Divisa       string `bson:"divisa"`
	Tipo         string `bson:"tipo"`
	Fecha_hora   string `bson:"fecha_hora"`
	ID_billetera string `bson:"id_billetera"`
}

// Conexión a la base de datos
func ConnectDB() *mongo.Client {
	Mongo_URL := "mongodb://127.0.0.1:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(Mongo_URL))

	if err != nil {
		log.Fatal("Error al crear el cliente de MongoDB:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error al conectar con MongoDB:", err)
	}

	// Verifica la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error al verificar la conexión con MongoDB:", err)
	}

	fmt.Println("Conectado a MongoDB")
	return client
}

// Funciones para obtener las colecciones
func GetClientesCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("myGoappDB").Collection("Clientes")
}

func GetBilleterasCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("myGoappDB").Collection("Billeteras")
}

func GetMovimientosCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("myGoappDB").Collection("Movimientos")
}

func main() {
	// Conectar a la base de datos
	client := ConnectDB()
	defer client.Disconnect(context.Background())

	// Insertar datos en la colección Clientes
	clienteCollection := GetClientesCollection(client)
	nuevoCliente := Cliente{
		Nombre:                "Juan Perez",
		Contrasena:            "secreta123",
		Fecha_nacimiento:      "1990-01-01",
		Direccion:             "Calle Falsa 123",
		Numero_identificacion: "11223344",
		Email:                 "juan.perez@example.com",
		Telefono:              "987654321",
		Genero:                "Masculino",
		Nacionalidad:          "Chileno",
		Ocupacion:             "Ingeniero",
	}
	clienteResult, err := clienteCollection.InsertOne(context.Background(), nuevoCliente)
	if err != nil {
		log.Fatal("Error al insertar el cliente:", err)
	}
	fmt.Printf("Cliente insertado con ID: %v\n", clienteResult.InsertedID)

	// Insertar datos en la colección Billeteras
	billeteraCollection := GetBilleterasCollection(client)
	nuevaBilletera := Billetera{
		Nro_cliente: "11223344",
		Saldo:       "1000",
		Divisa:      "CLP",
		Nombre:      "Cuenta Ahorro",
		Activo:      true,
	}
	billeteraResult, err := billeteraCollection.InsertOne(context.Background(), nuevaBilletera)
	if err != nil {
		log.Printf("Error al insertar la billetera: %v", err) // Agrega más detalle al log
	} else {
		fmt.Printf("Billetera insertada con ID: %v\n", billeteraResult.InsertedID)
	}

	// Insertar datos en la colección Movimientos
	movimientoCollection := GetMovimientosCollection(client)
	nuevoMovimiento := Movimiento{
		Nro_cliente:  "11223344",
		Monto:        "200",
		Divisa:       "CLP",
		Tipo:         "Deposito",
		Fecha_hora:   "2024-09-25T12:00:00Z",
		ID_billetera: fmt.Sprintf("%v", billeteraResult.InsertedID), // Relacionar con la billetera insertada
	}
	movimientoResult, err := movimientoCollection.InsertOne(context.Background(), nuevoMovimiento)
	if err != nil {
		log.Fatal("Error al insertar el movimiento:", err)
	}
	fmt.Printf("Movimiento insertado con ID: %v\n", movimientoResult.InsertedID)
}
