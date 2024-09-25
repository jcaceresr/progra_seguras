package funciones

import (
	"log"
	"net/url"
)

func Crear_url(ruta string, query map[string]string) string {
	url, err := url.Parse("http://localhost:5000/api/" + ruta)
	if err != nil {
		log.Fatal("URL no válida")
	}

	if query == nil {
		return url.String()
	}

	values := url.Query()

	for key, value := range query {
		values.Add(key, value)
	}

	url.RawQuery = values.Encode()

	return url.String()
}
