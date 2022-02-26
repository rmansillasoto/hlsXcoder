package servidor_web

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strings"
	//"crypto/tls"   //testing https
)


// formatoDePagina es el objeto que se encargará de mandar los datos a la página de videoreproductor generada.
// Se manda información sobre la ruta HLS.
type formatoDePagina struct {
	HlsUrl string
	Vumeter string
	RWidth string
	RHeight string
}



// manejadorPaginaWebVideoPlayer(writer http.ResponseWriter, request *http.Request) es la función del servidor web que se encarga de presentar
// la página del reproductor web.
func manejadorPaginaWebVideoPlayer(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(request)
	mediaId, mediaOk := vars["mediaId"]
	_, htmlOk := vars["html"]
	resource, resourcesOk := vars["resources"]
	path, _ := vars["path"]
	if path != "" {
		path = "/" + path
	}
	webpagePath := configuration.Videoplayer_path
	if mediaOk {
		if htmlOk {
			temp,err:=template.ParseFiles(webpagePath + "/index.html")
			if(err!=nil){
				fmt.Println("No se han encontrado los archivos web del reproductor en: "+webpagePath)
				writer.WriteHeader(http.StatusNotFound)
			}else{
				tmpl := template.Must(temp,err)
				vumeter:="none";
				if(configuration.Vumeter){
					vumeter="block";
				}
				RWidth:="auto"
				RHeight:="auto"
				if(configuration.Resolution!="" && configuration.Rf){
					arrayResolucion:=strings.Split(configuration.Resolution,"x")
					RWidth=arrayResolucion[0]
					RHeight=arrayResolucion[1]
				}
				//Cambiado a https por el tema de seguridad
				data := formatoDePagina{HlsUrl: "https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + mediaId + path + "/hls.m3u8",Vumeter:vumeter,RWidth:RWidth,RHeight:RHeight}
				//old HTTP
				//data := formatoDePagina{HlsUrl: "http://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + mediaId + path + "/hls.m3u8",Vumeter:vumeter,RWidth:RWidth,RHeight:RHeight}
				tmpl.Execute(writer, data)
			}

		} else if resourcesOk {
			servirFichero(writer, request, webpagePath, resource, 0)
		}
	}
}

func anhadirRutasReproductor(router *mux.Router, path string)  {
	router.HandleFunc(path+"/{html:(?:index.html)}", manejadorPaginaWebVideoPlayer).Methods("GET")
	router.HandleFunc(path+"/{resources:(?:.{2,100}[.](?:js|css|png))}", manejadorPaginaWebVideoPlayer).Methods("GET")
}

