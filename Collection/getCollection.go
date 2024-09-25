package getcollection

import (
	"go.mongodb.org/mongo-driver/mongo"
)

/*
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Posts") ///////////// Editar aqui segun nuettra base de datos
	return collection
}

func GetCollectionVuelos(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Vuelos") ///////////// Editar aqui segun nuettra base de datos
	return collection
}

func GetCollectionReservas(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Reservas") ///////////// Editar aqui segun nuettra base de datos
	return collection
}*/

func GetCollectionClientes(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Clientes")
	return collection
}

func GetCollectionBilleteras(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Billeteras")
	return collection
}

func GetCollectionMovimientos(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("myGoappDB").Collection("Movimientos")
	return collection
}
