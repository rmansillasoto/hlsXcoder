package servidor_web

import (
	"fmt"
	//"os"
	//"log"
	"github.com/gorilla/mux"
	"mime"
	//"net"
	"net/http"
	"path/filepath"
	"transcoder/configuracion"
	//"github.com/kabukky/httpscerts"  //testing https autocerts
	//"crypto/tls"  //testing https
)

var configuration configuracion.Config

// anhadirRutas(router *mux.Router, path string) se encarga de asignar una función que manejará las peticiones recibidas
// a través de las rutas especificadas por una expresión regular.
func anhadirRutas(router *mux.Router, path string) {
	finalPath := "/" + configuration.Url_subpath
	mediaPath := "/{mediaId:[0-9|a-z|-]+}"
	finalPath += mediaPath + path

	anhadirRutasHls(router, finalPath)
	anhadirRutasReproductor(router,finalPath)
}


// handlers se encarga de llamar a anhadirRutas para generar las rutas URL
// para las distintas codificaciones que se introduzcan como parámetro.
func handlers(encodings []configuracion.Codificacion) *mux.Router {
	router := mux.NewRouter()
	if len(encodings) > 0 {
		for _, codificacion := range encodings {
			anhadirRutas(router, "/{path:"+codificacion.Nombre+"}")
		}
	} else {
		anhadirRutas(router, "")

	}
	return router
}

// servirFichero función del servidor web que se encarga de servir ficheros. Estos ficheros pueden ser
// fragmentos del protocolo HLS o archivos de recurso de la página web del reproductor.
func servirFichero(writer http.ResponseWriter, request *http.Request, mediaBase string, filename string, tipoStream int) {
	var extension = filepath.Ext(filename)
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, filename)

	var contentType string
	var tipo string
	if tipoStream == 1 {
		tipo = "audio"
	} else {
		tipo = "video"
	}
	contentType = mime.TypeByExtension(extension)
	if configuration.Log {
		fmt.Println(extension, contentType, mediaFile)

	}
	if contentType == "" {
		switch extension {
		case ".m3u8":
			contentType = "application/x-mpegURL"
			break
		case ".ts":
			contentType = tipo + "/MP2T"
			break
		case ".mp4":
			contentType = tipo + "/mp4"
			break
		case ".vtt":
			contentType = "text/vtt"
			break
		case ".m4s":
			contentType = "text/plain"
			break
		case ".js":
			mediaFile = mediaBase + "/js/" + filename
			contentType = "text/javascript"
			break
		case ".css":
			mediaFile = mediaBase + "/styles/" + filename
			contentType = "text/css"
			break
		case ".png":
			mediaFile = mediaBase + "/img/" + filename
			contentType = "image/png"
			break

		}
	}
	writer.Header().Set("Content-Type", contentType)
	http.ServeFile(writer, request, mediaFile)
}
//Versión del WebServer en HTTP
// arrancarServidorWeb()  se encarga de arrancar el servidor web para disponer las conexiones HTTP para HLS.
//func arrancarServidorWeb(channel chan string){
	//http.Handle("/", handlers(configuration.Encodings))
	//listener, err := net.Listen("tcp", configuration.Ip)

	//resultado:="Servidor web se ha cerrado"
	//if err != nil {
		//resultado=err.Error()+"\n"+resultado
	//} else {
		//fmt.Println("\nServidor iniciado en", "http://"+configuration.Ip)
		//http.Serve(listener, nil)
	//}
	//channel<-resultado
//}

//PRUEBAS SERVER HTTPS para evitar fallo de seguridad en Titania

//with overon certs
func arrancarServidorWeb(channel chan string){
	http.Handle("/", handlers(configuration.Encodings))
	err := http.ListenAndServeTLS(":"+ configuration.Port,"overon_es.pem", "overon_es.key", nil)
	resultado:="Servidor web se ha cerrado"
	if err != nil {
		resultado=err.Error()+"\n"+resultado
		}else {
		fmt.Println("\nServidor iniciado en", "https://"+configuration.Ip)
		}
	channel<-resultado
}
//Fin PRUEBAS HTTPS

// InicializarServidorWeb()  se encarga de inicializar y configurar el servidor web.
func InicializarServidorWeb(configuracion configuracion.Config,main_channel chan string) {
	configuration = configuracion
	if(!configuration.Wn && !configuration.Test) {
		channel := make(chan string)
		go arrancarServidorWeb(channel)
		fmt.Println(<-channel)
		main_channel<-"Programa finalizado"
	}
}
