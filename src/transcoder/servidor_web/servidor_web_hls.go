package servidor_web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// manejadorHLS(writer http.ResponseWriter, request *http.Request) es la función del servidor web que se encarga de manejar las peticiones HTTP
// recibidas a través de las URL dedicadas al protocolo HLS. Como respuesta envía los fragmentos HLS.
func manejadorHLS(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(request)

	mediaId, mediaOk := vars["mediaId"]
	if !mediaOk {
		writer.WriteHeader(http.StatusNotFound)
	}

	path, pathOk := vars["path"]
	filename, filenameOk := vars["filename"]
	subName, subOk := vars["subName"]
	audioName, audioOk := vars["audioName"]

	if mediaOk && filenameOk {
		mediaBase := configuration.Output + "/" + mediaId + "/hls"
		if pathOk {
			mediaBase = fmt.Sprintf("%s/%s", mediaBase, path)
		}
		if audioOk {
			servirFichero(writer, request, mediaBase+"/audio/"+audioName, filename, 1)

		} else if subOk {
			servirFichero(writer, request, mediaBase+"/subtitulos/"+subName, filename, 1)

		} else {
			servirFichero(writer, request, mediaBase, filename, 0)
		}

	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func anhadirRutasHls(router *mux.Router, path string) {
	// maxcharsize es el número máximo de caracteres a introducir. Esto está así por temas de seguridad.
	maxcharsize := "20"
	router.HandleFunc(path+"/{filename:[a-z|0-9]{1,"+maxcharsize+"}.m3u8}", manejadorHLS).Methods("GET")
	router.HandleFunc(path+"/{filename:[a-z|0-9]{0,"+maxcharsize+"}(?:-)?segment-[0-9]{1,"+maxcharsize+"}.(?:ts|m4s)}", manejadorHLS).Methods("GET")
	router.HandleFunc(path+"/{filename:(?:init.mp4)}", manejadorHLS).Methods("GET")
	router.HandleFunc(path+"/subtitulo/{subName:(?:[a-z,]{2,"+maxcharsize+"})}/{filename:(?:(?:[0-9]{2,15}-)?subtitulo(?:_vtt)?.m3u8)}", manejadorHLS).Methods("GET")
	router.HandleFunc(path+"/subtitulo/{subName:(?:[a-z,]{2,"+maxcharsize+"})}/{filename:(?:(?:[0-9]{2,15}-)?[a-z,]{2,"+maxcharsize+"})[0-9]{1,"+maxcharsize+"}.(?:ts|m4s|vtt)}", manejadorHLS).Methods("GET")
	router.HandleFunc(path+"/audio/{audioName:(?:[a-z|_|0-9]{2,"+maxcharsize+"})}/{filename:(?:audio.m3u8)}", manejadorHLS).Methods("GET")                              //changed from [a-z,] to [a-z|0-9] to allow numbers in audio name
	router.HandleFunc(path+"/audio/{audioName:(?:[a-z|_|0-9]{2,"+maxcharsize+"})}/{filename:(?:audio)[0-9]{1,"+maxcharsize+"}[.].{1,10}}", manejadorHLS).Methods("GET") //changed from [a-z,] to [a-z|0-9] to allow numbers in audio name
	router.HandleFunc(path+"/audio/{audioName:(?:[a-z|_|0-9]{2,"+maxcharsize+"})}/{filename:(?:init.mp4)}", manejadorHLS).Methods("GET")                                //changed from [a-z,] to [a-z|0-9] to allow numbers in audio name
}
