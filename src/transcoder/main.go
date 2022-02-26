//main es el paquete donde está el archivo principal del programa.
package main

import (
	"fmt"
	"transcoder/codificacion"
	"transcoder/entrada"
	"transcoder/servidor_web"
)

func main() {
	ok, configuration := entrada.ObtenerArgumentos()
	main_channel := make(chan string)
	if ok {
		go codificacion.InicializarCodificaciones(configuration, main_channel)
		go servidor_web.InicializarServidorWeb(configuration, main_channel)
		fmt.Println(<-main_channel)
	}
}
