package funciones

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func RealizarDeposito(numero_identificacion string) {
	fmt.Print("\nIngrese un monto: ")

	var montoStr string
	fmt.Scan(&montoStr)

	monto, err := strconv.ParseFloat(montoStr, 64)
	if err != nil {
		fmt.Println("Monto inválido. Intente nuevamente.")
		return
	}

	deposito := map[string]string{
		"nro_cliente": numero_identificacion,
		"monto":       strconv.FormatFloat(monto, 'f', 2, 64),
		"divisa":      "USD",
	}

	depositoJson, _ := json.Marshal(deposito)

	url := Crear_url("deposito", nil)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(depositoJson))
	if err != nil {
		log.Fatal("Error al realizar depósito:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, _ := client.Do(req)

	defer req.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("El depósito ha sido enviado correctamente")
	}

}
