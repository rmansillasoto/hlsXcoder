package entrada

import (
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"transcoder/configuracion"
)

// arrayFlags es un tipo de dato que contendrá información de las codificaciones.
// Almacena los valores de encoding:bitrates introducido por el input
// Esto se ha hecho para poder introducir como argumentos varias veces -c -c -c como argumentos.
type arrayFlags []string

// se sobreescribe la función para devolver un string de arrayFlags.
func (i *arrayFlags) String() string {
	return fmt.Sprintf("%s", *i)
}

// se sobreescribe para añadir un nuevo valor a arrayFlags.
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// generarIdentificador(s string) string que genera un identificador de codificación en base al nombre del input introducido.
func generarIdentificador(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func transformarEntrada(input *string) {
	rYoutube := regexp.MustCompile(`https://www.youtube.com/`)
	rUdp := regexp.MustCompile(`udp://([a-z|A-Z]+|([0-9]{1,3}\.){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)
	rRtp := regexp.MustCompile(`rtp://([a-z|A-Z]+|([0-9]{1,3}\.){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)      //added for rtp v1.4.3
	rSrt := regexp.MustCompile(`srt://([a-z|A-Z]+|([0-9]{1,3}\.){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)      //added for srt v1.4.4
	rRtsp := regexp.MustCompile(`rtsp://([a-z|A-Z]+|([0-9]{1,3}\./){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)   //added for rstp v1.4.4
	rRtmp := regexp.MustCompile(`rtmp://([a-z|A-Z]+|([0-9]{1,3}\./){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)   //added for rtmp v1.4.4
	rHttp := regexp.MustCompile(`http://([a-z|A-Z]+|([0-9]{1,3}\./){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`)   //added for hls (http) v1.4.4
	rHttps := regexp.MustCompile(`https://([a-z|A-Z]+|([0-9]{1,3}\./){3}[0-9]{1,3}):[0-9]{1,6}(/.+)*`) //added for hls (https) v1.4.4

	if rUdp.MatchString(*input) {
		*input = *input + "?fifo_size=557753" //modified fifo_size -> fifo_size en Mb, por ejemplo 50Mbps -> 50*1024*1024/188 = 278876 y cambiado el overrun a 0 para forzsr reboot de fuente.
	} else if rYoutube.MatchString(*input) {
		fmt.Println("Detectada URL de youtube. Intentando obtener video...")
		cmd := exec.Command("youtube-dl", "-g", "-f", "96", "--hls-prefer-ffmpeg", "--restrict-filenames", *input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Youtube-dl ha fallado con el siguiente error %s\n", err)
		} else {
			*input = strings.TrimSuffix(string(out), "\n")

			fmt.Println("URL obtenida: " + *input)
		}
	} else if rRtp.MatchString(*input) { //añadido regexp para inputs tipo RTP v1.4.3
		*input = *input
	} else if rSrt.MatchString(*input) { //añadido regexp para inputs tipo SRT v1.4.4
		*input = *input
	} else if rRtsp.MatchString(*input) { //añadido regexp para inputs tipo RTSP v1.4.4
		*input = *input
	} else if rRtmp.MatchString(*input) { //añadido regexp para inputs tipo RTMP v1.4.4
		*input = *input
	} else if rHttp.MatchString(*input) { //añadido regexp para inputs tipo HTTP v1.4.4
		*input = *input
	} else if rHttps.MatchString(*input) { //añadido regexp para inputs tipo HTTP v1.4.4
		*input = *input
	}
}

// ObtenerArgumentos() bool Comprueba, transforma y almacena los argumentos que se introducen a través de los parámetros del programa.
// En caso de que uno falle se devuelve false y la ejecución del programa se cancela.
func ObtenerArgumentos() (bool, configuracion.Config) {
	ok := true

	var codificacionesFlag arrayFlags
	var codificaciones []configuracion.Codificacion
	nvenc := false

	input := flag.String("i", "", "Introduce un fichero o dirección de entrada")
	output := flag.String("o", "./media", "Introduce el directorio donde se generará la salida")
	videoplayer_path := flag.String("videoplayer_path", "./www", "Introduce la ruta del directorio donde se ubican los archivos del videoreproductor web")
	ip := flag.String("ip", "localhost", "Introduce la ip donde iniciar el servidor HLS")
	port := flag.Int("port", 8080, "Introduce el puerto donde se va a iniciar el servidor HLS")
	resolution := flag.String("r", "", "Introduce la resolución del video. Por defecto el propio del video")
	sid := flag.String("sid", "", "Introduce el Sid del programa (Por defecto se utiliza el primer programa)") //cambiado de pid a sid, descripción más logica
	maxfragments := flag.Int("hls_list_size", 30, "Introduce el número máximo de fragmentos.")
	disable_streaming := flag.Bool("ds", false, "Deshabilita el modo streaming")
	url_subpath := flag.String("url_subpath", "streaming", "Introducir un subpath a la URL.")
	hls_time := flag.Int("hls_time", 10, "Introduce el tamaño (En segundos) de los fragmentos HLS")
	re := flag.Bool("re", false, "El video se codificará al ratio de frames nativo")
	sn := flag.Bool("sn", false, "Desactivar subtitulos")
	an := flag.Bool("an", false, "Desactivar audio")
	vumeter := flag.Bool("vumeter", false, "Activar vumetros")
	test := flag.Bool("test", false, "Visualizar streams")
	wn := flag.Bool("wn", false, "Desactivar servidor web (Solo genera los fragmentos)")
	sub_lang := flag.String("sub_lang", "", "Selecciona que subtitulos codificar.Ej: -sub_lang esp,eng")
	audio_lang := flag.String("audio_lang", "", "Selecciona que audios codificar.Ej: -audio_lang esp,eng")
	streams := flag.String("streams", "", "Selecciona que streams va a procesar el programa.Ej: -streams 0x14d,0x35f")
	hwaccel := flag.Bool("hwaccel", false, "Activar aceleración por GPU, bueno para unos casos, malo para otros.(Si el stream tiene cambios de Ratio de aspecto no usar, ya que se caerá el reproductor)")
	logger := flag.Bool("log", false, "Activar mensajes de peticiones HTTP recibidas")
	deint := flag.Bool("deint", false, "Activar el desentralazado")
	rf := flag.Bool("rf", false, "Forzar la resolucion del reproductor web al tamaño del video")
	codificacion_bitrate_audio := flag.String("ca", "", "Introducir la codificacion del audio deseada")
	tid := flag.String("tid", "", "Table ID para Titania")
	flag.Var(&codificacionesFlag, "cv", "Introducir codificacion o codificaciones de video junto con bitrates tal que -cv h264 -cv h265:5000,2800")
	flag.Parse()

	rCodificaciones := regexp.MustCompile(`([^:]{1,10})(?::([0-9]{1,5}(?:,[0-9]{1,5})*))?`)

	if len(*input) <= 0 {
		fmt.Println("No se ha introducido ninguna entrada, prueba a mirar los argumentos con el parámetro -h")
		ok = false
	}

	for _, codificacionFlag := range codificacionesFlag {
		if !rCodificaciones.MatchString(codificacionFlag) {

			//log.Fatal("Se ha introducido mal algun argumento para codificación. Prueba a mirar los argumentos con el parámetro -h")
			fmt.Println("Se ha introducido el argumento para codificación prueba a mirar los argumentos con el parámetro -h")
			ok = false
			break
		} else {
			var codificacion configuracion.Codificacion
			submatches := rCodificaciones.FindStringSubmatch(codificacionFlag)

			if strings.Contains(submatches[1], "h265") {
				submatches[1] = strings.Replace(submatches[1], "h265", "hevc", -1)
				codificacion.Nombre = submatches[1]
			} else {
				codificacion.Nombre = submatches[1]

			}

			if strings.Contains(codificacion.Nombre, "_nvenc") {
				codificacion.Nombre = strings.Replace(codificacion.Nombre, "_nvenc", "", -1)
				nvenc = true
			}

			for _, submatch := range strings.Split(submatches[2], ",") {
				if submatch != "" {
					codificacion.Bitrates = append(codificacion.Bitrates, submatch)
				}
			}
			codificaciones = append(codificaciones, codificacion)
		}
	}

	audio_codificacion := ""
	audio_bitrate := ""
	if *codificacion_bitrate_audio != "" {
		audioCodArray := strings.Split(*codificacion_bitrate_audio, ":")
		audio_codificacion = audioCodArray[0]
		if len(audioCodArray) > 1 {
			audio_bitrate = audioCodArray[1]
		}
	}

	subs_lang_map := make(map[string]struct{})
	audio_lang_map := make(map[string]struct{})
	streams_map := make(map[string]struct{})
	for _, elem := range strings.Split(*sub_lang, ",") {
		if elem != "" {
			subs_lang_map[elem] = struct{}{}
		}
	}
	for _, elem := range strings.Split(*audio_lang, ",") {
		if elem != "" {
			audio_lang_map[elem] = struct{}{}
		}
	}
	for _, elem := range strings.Split(*streams, ",") {
		if elem != "" {
			streams_map[elem] = struct{}{}
		}
	}

	transformarEntrada(input)

	return ok, configuracion.Config{
		Ip:                 *ip + ":" + strconv.Itoa(*port),
		Port:               strconv.Itoa(*port), //añadido para poder pasar el puerto al nombre del archivo que mira titania
		Input:              *input,
		Output:             *output,
		Videoplayer_path:   *videoplayer_path,
		Resolution:         *resolution,
		Encodings:          codificaciones,
		Hls_fragments:      *maxfragments,
		Hls_time:           *hls_time,
		Id_programa:        *sid, //cambiado de pid a sid, descripción más logica
		Tid:                *tid,
		Audio_bitrate:      audio_bitrate,
		Sub_lang:           subs_lang_map,
		Audio_lang:         audio_lang_map,
		Streams:            streams_map,
		Sn:                 *sn,
		An:                 *an,
		Wn:                 *wn,
		Re:                 *re,
		Vumeter:            *vumeter,
		Deint:              *deint,
		Rf:                 *rf,
		Test:               *test,
		Nvenc:              nvenc,
		Hwaccel:            *hwaccel,
		Disable_streaming:  *disable_streaming,
		Log:                *logger,
		Audio_codificacion: audio_codificacion,
		Url_subpath:        *url_subpath,
		Server:             *ip, //para tener la ip del server sin el puerto
		MediaId:            generarIdentificador(*input)}
}
