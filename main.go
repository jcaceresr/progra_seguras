package main

import (
	funciones "api/funciones"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Bienvenido a TrustBank!")

	for {
		fmt.Print("1. Iniciar sesión\n")
		fmt.Print("2. Salir\n")
		fmt.Print("Ingrese una opción: ")

		var opcion int
		_, err := fmt.Scanf("%d", &opcion)
		if err != nil {
			fmt.Print("\nOpción inválida. Intente nuevamente.\n")
			continue
		}

		// Consumir el carácter de nueva línea
		fmt.Scanln()

		switch opcion {
		case 1:
			loginExitoso, numero_identificacion := funciones.IniciarSesion()
			if loginExitoso {
				realizarOperaciones(numero_identificacion)
			}
		case 2:
			fmt.Println("Gracias por usar TrustBank!")
			os.Exit(0)
		default:
			fmt.Print("Opción inválida. Intente nuevamente.\n")
		}
	}
}

// menu de operaciones

func realizarOperaciones(numero_identificacion string) {

	for {

		fmt.Print("\n1. Realizar depósito")
		fmt.Print("\n2. Realizar transferencia")
		fmt.Print("\n3. Realizar giro")
		fmt.Print("\n4. Salir")

		fmt.Print("\nIngrese una opción: ")

		var opcionStr string
		fmt.Scan(&opcionStr)
		opcionStr = strings.TrimSpace(opcionStr)

		opcion, err := strconv.Atoi(opcionStr)
		if err != nil {
			fmt.Println("Opción inválida. Intente nuevamente.")
			continue
		}

		switch opcion {
		case 1:
			funciones.RealizarDeposito(numero_identificacion)
		case 2:
			funciones.RealizarTransferencia(numero_identificacion)
		case 3:
			funciones.RealizarGiro(numero_identificacion)
		case 4:
			fmt.Println("Gracias por usar TrustBank!")
			os.Exit(0)
		default:
			fmt.Println("Opción inválida. Intente nuevamente.")
		}
	}
}
