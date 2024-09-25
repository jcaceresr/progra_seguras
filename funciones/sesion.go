package funciones

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Verificar_sesion(numero_identificacion string, contrasena string) bool {
	sesion := map[string]string{
		"numero_identificacion": numero_identificacion,
		"contrasena":            contrasena,
	}

	sesionJson, err := json.Marshal(sesion)
	if err != nil {
		log.Fatal("Error al verificar sesion:", err)
	}

	url := Crear_url("inicio_sesion", nil)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(sesionJson))
	if err != nil {
		log.Fatal("Error al crear solicitud:", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error al realizar la solicitud:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}

func IniciarSesion() (bool, string) {
	fmt.Print("Ingrese su número de identificación: ")
	var numero_identificacion string
	fmt.Scan(&numero_identificacion)

	fmt.Print("Ingrese su contraseña: ")
	var contraseña string
	fmt.Scan(&contraseña)

	verificacion := Verificar_sesion(numero_identificacion, contraseña)

	if verificacion {
		fmt.Println("¡Inicio de sesión exitoso!")
		return true, numero_identificacion
	} else {
		fmt.Println("Número de identificación o contraseña incorrecta")
		return false, ""
	}
}
