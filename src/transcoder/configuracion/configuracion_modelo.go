
// Paquete donde se almacenarán todas las estructuras y funciones de configuracion procesadas tras la entrada.
package configuracion


//Config es la estructura que contiene toda la información de configuración de la aplicación.
//Contiene todos aquellos parámetros introducidos por el terminal así como el programa seleccionado para realizar el encoding.
type Config struct {
	Input              string
	Output             string
	Ip                 string
	Port			   string  //añadido para pasar al archivo que generamos para titania con las url el Target Port
	Resolution         string
	MediaId            string
	Url_subpath        string
	Audio_codificacion string
	Audio_bitrate      string
	Hls_fragments      int
	Hls_time           int
	Id_programa        string
	Disable_streaming  bool
	Re                 bool
	Hwaccel            bool
	Nvenc              bool
	Sn                 bool
	An                 bool
	Log                bool
	Sub_lang           map[string]struct{}
	Audio_lang         map[string]struct{}
	Streams            map[string]struct{}
	Encodings          []Codificacion
	Wn                 bool
	Videoplayer_path   string
	Vumeter            bool
	Test               bool
	Deint              bool
	Rf                 bool
	Tid				   string  //añadido para pasar un Table ID de Titania.
	Server			   string  //para tener la ip del server sin el puerto
}

//Codificacion es la estructura que almacena la codificación deseada para un stream y sus bitrates.
type Codificacion struct {
	Nombre   string   // nombre incluye información del nombre de la codificación.
	Bitrates []string // bitrates es un array que almacena los bitrates introducidos por consola de la aplicacion.
}
