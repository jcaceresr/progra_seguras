package funciones

import (
	model "api/model"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func RealizarTransferencia(numero_identificacion string) {
	fmt.Print("Ingrese la cuenta de destino: ")
	var cuentaDestino string
	fmt.Scan(&cuentaDestino)

	fmt.Printf("\nIngrese Monto: ")
	var monto float64
	fmt.Scan(&monto)

	_, err := Obtener_cliente(cuentaDestino)
	if err != nil {
		fmt.Println("Error al obtener cliente:", err)
		return
	}

	// Realizar la transferencia

	verificar, err := Post_transferencia(numero_identificacion, cuentaDestino, monto, "USD")

	if !verificar {
		fmt.Println("Error al realizar transferencia:", err)
		return
	}

	if err != nil {
		fmt.Println("Error al realizar transferencia:", err)
		return
	}

	fmt.Println("Transferencia enviada")

}

func Obtener_cliente(numero_identificacion string) (model.Cliente, error) {
	url := Crear_url("cliente", nil)

	// Crear el cuerpo de la solicitud con el número de identificación en formato JSON
	reqBody := struct {
		NumeroIdentificacion string `json:"numero_identificacion"`
	}{
		NumeroIdentificacion: numero_identificacion,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return model.Cliente{}, err
	}

	// Crear una solicitud HTTP con el método GET y el cuerpo de la solicitud
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return model.Cliente{}, err
	}

	// Establecer el encabezado Content-Type para indicar que se envía un cuerpo en formato JSON
	req.Header.Set("Content-Type", "application/json")

	// Realizar la solicitud HTTP y obtener la respuesta
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.Cliente{}, err
	}
	defer resp.Body.Close()

	var cliente model.Cliente
	err = json.NewDecoder(resp.Body).Decode(&cliente)
	if err != nil {
		return model.Cliente{}, err
	}

	return cliente, nil
}

func Post_transferencia(nro_cliente_origen string, nro_cliente_destino string, monto float64, divisa string) (bool, error) {
	url := Crear_url("transferencia", nil)

	transferencia := map[string]string{
		"nro_cliente_origen":  nro_cliente_origen,
		"nro_cliente_destino": nro_cliente_destino,
		"monto":               strconv.FormatFloat(monto, 'f', 2, 64),
		"divisa":              "USD",
	}

	transferenciaJson, _ := json.Marshal(transferencia)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(transferenciaJson))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, _ := client.Do(req)

	defer req.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil

}
