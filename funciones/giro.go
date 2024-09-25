package funciones

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func RealizarGiro(numero_identificacion string) {
	fmt.Print("\nIngrese un monto: ")

	var montoStr string
	fmt.Scan(&montoStr)

	monto, err := strconv.ParseFloat(montoStr, 64)
	if err != nil {
		fmt.Println("Monto inválido. Intente nuevamente.")
		return
	}

	//verificar si tiene saldo suficiente

	giro := map[string]string{
		"nro_cliente": numero_identificacion,
		"monto":       strconv.FormatFloat(monto, 'f', 2, 64),
		"divisa":      "USD",
	}

	giroJson, _ := json.Marshal(giro)

	url := Crear_url("giro", nil)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(giroJson))
	if err != nil {
		log.Fatal("Error al realizar giro:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, _ := client.Do(req)

	defer req.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Giro enviado")
	}

	fmt.Print("El giro ha sido enviado correctamente\n")

}
