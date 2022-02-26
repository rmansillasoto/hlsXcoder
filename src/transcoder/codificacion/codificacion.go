package codificacion

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"transcoder/configuracion"

	//"io/ioutil"  //used for translate body_response from titania after a post
	"crypto/tls"
	"encoding/json"
	"net/http"
)

// Stream es la estructura que almacena informacion de un stream de video, audio o subtítulos.
type Stream struct {
	id     string
	hexaId string
	//decimalId string //added for Hexadecimal PID conversion to Decimal PID
	format   string
	language string
	fps      string
	chroma   string //added for getting information about wether the stream is 4:2:0 or 4:2:2
}

// Programa es la estructura que almacena información del programa seleccionado.
type Programa struct {
	id         string
	nombre     string
	video      Stream
	audio      []Stream //modified for checking more than one stream
	subtitulos []Stream //modified for checking more than one stream
}

var programa Programa //programa contendrá la información del input introducido.
var configuration configuracion.Config

// calcularBitrateMaximo(bitrate string) string función que se encarga de generar el máximo bitrate para el argumento de la herramienta FFMpeg.
func calcularBitrateMaximo(bitrate string) string {
	num, err := strconv.Atoi(bitrate[0:2])
	var maxbitrate string
	if err != nil {
		maxbitrate = bitrate
	} else {
		maxbitrate = strconv.Itoa((7 * num) + (num * 100))
	}
	return maxbitrate
}

// generarArgumentosDVBSubs(codificacionPath string, bitrate string, codificacionNombre string) []string es la función que se encarga de generar
// los argumentos para quemar los subtitulos DVBSubs y generar streams de video con ellos integrados.
func generarArgumentosDVBSubs(codificacionPath string, bitrate string, codificacionNombre string) []string {
	var subtitles []string
	var resultado []string

	videoMapId := getMapId(programa.video)
	filterComplex := "[" + videoMapId + "]hwdownload,format=nv12"
	split := ",split=" + strconv.Itoa(len(programa.subtitulos))
	overlay := ""

	for index, elem := range programa.subtitulos {
		subtitleMapId := getMapId(elem)
		subsPath := codificacionPath + "/subtitulos/" + elem.language
		os.RemoveAll(subsPath)
		os.MkdirAll(subsPath, os.ModePerm)
		var bitrateCommand []string
		bitratePath := ""
		videoRef := "[base" + strconv.Itoa(index) + "]"
		//Added scale2ref filter before overlay as subtitles didnt scale when resolution != original size. With scale2ref subs will scale to match dar and video size, then overlay keeping margins and centering
		overlay += "[" + subtitleMapId + "]" + videoRef + "scale2ref=ih*mdar:ih[subtitle" + strconv.Itoa(index) + "][ref" + strconv.Itoa(index) + "];[ref" + strconv.Itoa(index) + "][subtitle" + strconv.Itoa(index) + "]overlay=(W-w)/2:(H-h)/2,hwupload_cuda[sub" + strconv.Itoa(index) + "]"

		//original overlay sentence -> overlay += videoRef + "[" + subtitleMapId + "]overlay,hwupload_cuda[sub" + strconv.Itoa(index) + "]"

		if index < len(programa.subtitulos)-1 {
			overlay += ";"
		}
		split += videoRef

		if bitrate != "" {
			bitratePath = bitrate + "-"
			bitrateCommand = []string{"-b:v", bitrate + "k", "-maxrate", calcularBitrateMaximo(bitrate) + "k"}
		}

		if configuration.Hwaccel {
			nvenc := "" //Added: Fixed BUG subtitles doesn´t use GPU when enabled, now every stream should have codificacionNombre+nvenc
			if configuration.Nvenc {
				nvenc = "_nvenc"
			}
			subtitles = append(subtitles, bitrateCommand...)
			subtitles = append(subtitles,
				"-map", "[sub"+strconv.Itoa(index)+"]", "-an", "-sn", "-c:v", codificacionNombre+nvenc, "-preset", "llhp", "-profile:v", "baseline", "-zerolatency", "1", "-cbr", "1", "-rc", "cbr", "-rc-lookahead", "50") //added +nvenc for better GPU use in subtitle streams
		} else {
			subtitles = append(subtitles,
				"-filter_complex", "["+videoMapId+"]["+subtitleMapId+"]overlay[v]")
			subtitles = append(subtitles,
				"-map", "[v]", "-an", "-sn", "-vcodec", codificacionNombre)
		}
		subtitles = append(subtitles, generarArgumentosHLS(subsPath+"/"+bitratePath+"subtitulo", subsPath+"/"+bitratePath+"subtitulo%01d", codificacionNombre)...)

	}

	if len(programa.subtitulos) > 0 && configuration.Hwaccel {
		resultado = append(resultado, "-filter_complex", filterComplex+split+";"+overlay)
	}
	resultado = append(resultado, subtitles...)

	return resultado
}

// func obtenerStreams() map[string]*Programa es la función que se encarga de recoger y guardar los streams obtenidos del comando ffprobe
// para su posterior transcodificación.

//FUNCION OBTENER STREAMS PARA POST TITANIA

func obtenerStreamsPost() map[string]*Programa {
	sub_lang_ok := len(configuration.Sub_lang) > 0
	audio_lang_ok := len(configuration.Audio_lang) > 0
	streams_ok := len(configuration.Streams) > 0

	var sid string
	var name string
	var video string
	var audio string
	var subtitulo string
	var vid2 []string
	var aud2 []string
	var sub2 []string
	var ffprobeError bool = false
	var error string

	//Post Configuration
	var api_url string = "https://" + configuration.Server + "/api/custom/update_ffprobe"
	var user string = "titania"
	var pass string = "titanios"

	type urlData struct {
		Id         string
		VideoInput string
		Ffprobe    string
		Error      bool
	}

	udp := configuration.Input
	cmd := exec.Command("ffprobe", "-hide_banner", udp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		ffprobeError = true
		error = string(out)
		//POST data to JSON
		data := urlData{
			Id:         configuration.Tid,
			VideoInput: configuration.Input,
			Ffprobe:    error,
			Error:      ffprobeError,
		}
		postData, err1 := json.Marshal(data)

		//POST Request
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}
		req, err1 := http.NewRequest("POST", api_url, bytes.NewBuffer(postData))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.SetBasicAuth(user, pass)
		resp, err1 := client.Do(req)
		if err1 != nil {
			log.Fatal(err1)
		}
		resp.Body.Close()
		/*Printing POST RESULTS for tracing
		fmt.Printf("\n\nPost_Url: %s", api_url) //post url api
		fmt.Printf("\nPost_Status: %s", resp.Status) //post status response
		fmt.Printf("\njson: %s", string(postData)) //post data sent

		//Translate Body Response from Titania
		response, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Println(err2)
			log.Fatal(err2)
		}
		fmt.Printf("\nPost_Response: %v", string(response))
		resp.Body.Close()*/

		log.Fatalf("FFprobe ha fallado: %s\n", out)
	}
	programas := make(map[string]*Programa)

	streamReg := regexp.MustCompile(`Stream #0:([0-9]+)(?:\[((?:0[xX])?[[:xdigit:]]+)\])?(?:\(([a-zA-Z]{3}(?:,[a-zA-Z]{3})*?)\))?: ([a-zA-Z]+): ([a-z0-9_]+)`)
	fpsReg := regexp.MustCompile(`,[ ]*([0-9.]+) fps[ ]*,`)
	progReg := regexp.MustCompile(`Program ([0-9]+)`)
	chromaReg := regexp.MustCompile(`,[ ]*yuv([0-9]+)p[ ]*`) //added for checking video format 422 o 420
	serviceName := regexp.MustCompile(`service_name[ ]*: (.+)`)
	lineas := strings.Split(string(out), "\n")
	programId := ""
	for i := 0; i < len(lineas); i++ {
		lineaActual := lineas[i]
		if progReg.MatchString(lineaActual) && !streams_ok {
			if programId != "" {
				fmt.Printf("")
			}
			fmt.Printf("\n")
			programId = progReg.FindStringSubmatch(lineaActual)[1]
			if configuration.Id_programa == "" || configuration.Id_programa == programId {
				programa.id = programId
				if !configuration.Test {
					fmt.Printf("Servicios encontrados:\n")
				}
			}
			sid = programId
			fmt.Printf("\n\tSID: %s", programId)
			programas[programId] = &Programa{id: programId}
		} else if serviceName.MatchString(lineaActual) && !streams_ok {
			if programId != "" {
				programas[programId].nombre = serviceName.FindStringSubmatch(lineaActual)[1]
				name = programas[programId].nombre
				fmt.Printf("\n\tName: %s", programas[programId].nombre)
			}
		} else if streamReg.MatchString(lineaActual) {
			if programId == "" {
				programId = "-1"
				programas[programId] = &Programa{nombre: "No name", id: programId}
				sid = programas[programId].id
				name = programas[programId].nombre
				fmt.Printf("\n\tSID: %s\n\tName: %s", programId, programas[programId].nombre)
				if configuration.Id_programa == "" {
					programa.id = programId
				}
			}
			programa := programas[programId]
			stream := streamReg.FindStringSubmatch(lineaActual)
			streamElement := Stream{id: stream[1], hexaId: stream[2], language: stream[3], format: stream[5]}

			if streams_ok {
				_, stream_ok := configuration.Streams[streamElement.hexaId]
				if !stream_ok {
					continue
				}
			}
			switch stream[4] {
			case "Video":
				streamElement.fps = fpsReg.FindStringSubmatch(lineaActual)[1]
				streamElement.chroma = chromaReg.FindStringSubmatch(lineaActual)[1]
				programa.video = streamElement
				hexStream := streamElement.hexaId
				decimal := hex2int(hexStream)
				decimalId := strconv.FormatUint(decimal, 10)
				pids := []string{streamElement.hexaId, decimalId}
				stream[2] = strings.Join(pids, "_")
				fmt.Printf("\n\t %s %s %s %s %s %s %s", stream[1], stream[2], stream[3], stream[4], stream[5], streamElement.fps, streamElement.chroma)
				videoArray := []string{stream[1], stream[2], stream[3], stream[4], stream[5], streamElement.fps, streamElement.chroma}
				vid1 := strings.Join(videoArray, "_")
				vid2 = append(vid2, vid1)

				break
			case "Audio":
				if !configuration.An {
					hexStream := streamElement.hexaId
					decimal := hex2int(hexStream)
					decimalId := strconv.FormatUint(decimal, 10)

					if len(streamElement.language) <= 0 {
						//streamElement.language = decimalId
						stream[3] = "none"
						b := "none"
						streamElement.language = fmt.Sprintf("%s_%s", b, decimalId)
					} else {
						b := streamElement.language
						streamElement.language = fmt.Sprintf("%s_%s", b, decimalId)
					}

					ok := true
					if audio_lang_ok {
						_, ok = configuration.Audio_lang[streamElement.language]
					}
					//a := 2
					for _, audio := range programa.audio {
						if streamElement.language == audio.language {
							//streamElement.language = decimalId
							b := streamElement.language
							streamElement.language = fmt.Sprintf("%s_%s", b, decimalId)
							//a++
						}
					}
					if ok {
						programa.audio = append(programa.audio, streamElement)
						pids := []string{streamElement.hexaId, decimalId}
						stream[2] = strings.Join(pids, "_")
						fmt.Printf("\n\t %s %s %s %s %s", stream[1], stream[2], stream[3], stream[4], stream[5])
						audioArray := []string{stream[1], stream[2], stream[3], stream[4], stream[5]}
						aud1 := strings.Join(audioArray, "_")
						aud2 = append(aud2, aud1)
					}
				}
				break
			case "Subtitle":
				if !configuration.Sn {
					if len(streamElement.language) <= 0 {
						streamElement.language = "default"
					}
					ok := true
					if sub_lang_ok {
						_, ok = configuration.Sub_lang[streamElement.language]
					}
					for _, audio := range programa.subtitulos {
						if streamElement.language == audio.language {
							ok = false
						}
					}
					if ok {
						programa.subtitulos = append(programa.subtitulos, streamElement)
						hexStream := streamElement.hexaId
						decimal := hex2int(hexStream)
						decimalId := strconv.FormatUint(decimal, 10)
						pids := []string{streamElement.hexaId, decimalId}
						stream[2] = strings.Join(pids, "_")
						fmt.Printf("\n\t %s %s %s %s %s", stream[1], stream[2], stream[3], stream[4], stream[5])
						subtituloArray := []string{stream[1], stream[2], stream[3], stream[4], stream[5]}
						sub1 := strings.Join(subtituloArray, "_")
						sub2 = append(sub2, sub1)
					}
				}
				break
			}
		}
	}

	video = strings.Join(vid2, " / ")
	audio = strings.Join(aud2, " / ")
	subtitulo = strings.Join(sub2, " / ")

	//POST data to JSON
	data := urlData{
		Id:         configuration.Tid,
		VideoInput: configuration.Input,
		Ffprobe:    "SID:" + sid + ",ServiceName:" + name + ",VideoStreams:" + video + ",AudioStreams:" + audio + ",SubtitleStreams:" + subtitulo + "",
		Error:      ffprobeError,
	}
	postData, err := json.Marshal(data)

	//POST Request
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", api_url, bytes.NewBuffer(postData))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(user, pass)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	/*Printing POST RESULTS for tracing
	fmt.Printf("\n\nPost_Url: %s", api_url) //post url api
	fmt.Printf("\nPost_Status: %s", resp.Status) //post status response
	fmt.Printf("\njson: %s", string(postData)) //post data sent

	//Translate Body Response from Titania
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fmt.Printf("\nPost_Response: %v", string(response))
	resp.Body.Close()*/

	if programaHash, ok := programas[programa.id]; ok {
		programa = *programaHash
	} else {
		log.Fatal("Error no existe el PID")
	}

	return programas
}

// getMapId(stream Stream) string obtiene el id de map para el comando FFmpeg. Este id será o el identificador decimal o el hexadecimal dependiendo de si
// existe uno u otro.
func getMapId(stream Stream) string {
	if stream.hexaId == "" {
		return "0:" + stream.id
	} else if stream.id == "" {
		return "0:v:0"
	} else {
		return "i:" + stream.hexaId
	}

}

// inicializarFfmpegEncoding(channel chan string, codificacionIndex int) es la función que se encargará de realizar la codificación. Generará los argumentos
// para el comando FFmpeg y lo iniciará.
func inicializarFfmpegEncoding(channel chan string, codificacionIndex int) {

	customEncoding := codificacionIndex >= 0
	var codificacion configuracion.Codificacion
	if customEncoding {
		codificacion = configuration.Encodings[codificacionIndex]
	}
	obtenerStreamsPost()

	input := configuration.Input
	lenSubtitles := len(programa.subtitulos) > 0
	teletextSubtitles := false
	path := configuration.Output + "/" + configuration.MediaId + "/hls"

	//var commandsPreInput, commandsPostInput, commandsVideo, commandsAudio, commandsSubtitles []string

	var commandArgs []string
	codificacionPath := path

	videoFormat := programa.video.format

	switch programa.video.format {
	case "mpeg2video":
		videoFormat = "mpeg2"
		break
	}

	videoMap := getMapId(programa.video)

	if configuration.Re {
		commandArgs = append(commandArgs, "-re")
	}

	//modified for adjusting dvb_teletext canvas size when resize is enabled, else....native resolution.
	if lenSubtitles && programa.subtitulos[0].format == "dvb_teletext" {
		if configuration.Resolution != "" {
			commandArgs = append(commandArgs, "-txt_format", "text", "-txt_duration", "5000", "-canvas_size", configuration.Resolution) // "-fix_sub_duration", "-txt_page", "801", "-txt_page", "803"
			teletextSubtitles = true
		} else if configuration.Resolution == "" {
			commandArgs = append(commandArgs, "-txt_format", "text", "-txt_duration", "5000")
			teletextSubtitles = true
		}
	}

	var videoCommands []string
	var preInputCommands []string
	var postInputCommands []string

	//added for receiving RTP,RTSP,RTMP,HTTP streams... A veces, rtp streamos con Data Stream incorrect order, no luck yet.
	if strings.Contains(input, "rtp") {
		preInputCommands = append(preInputCommands, "-probesize", "6M", "-analyzeduration", "6M")
	} else if strings.Contains(input, "udp") {
		preInputCommands = append(preInputCommands, "-probesize", "3M", "-analyzeduration", "3M")
	} else if strings.Contains(input, "rtsp") {
		preInputCommands = append(preInputCommands, "-rtsp_transport", "tcp")
		videoCommands = append(videoCommands, "-movflags", "frag_keyframe+empty_moov")
	} else if strings.Contains(input, "rtmp") {
		preInputCommands = append(preInputCommands, "-f", "flv")
	} else if strings.Contains(input, "http") {
		preInputCommands = append(preInputCommands, "-re")
	}

	if customEncoding {
		codificacionPath += "/" + codificacion.Nombre

		nvenc := ""
		if configuration.Nvenc {
			//preInputCommands = append(preInputCommands,
			//	"-c:v", ""+videoFormat+"_cuvid")
			nvenc = "_nvenc"
			videoCommands = append(videoCommands, "-map_chapters", "-1", "-c:v", codificacion.Nombre+nvenc, "-preset", "llhp", "-zerolatency", "1", "-cbr", "1", "-rc", "cbr", "-rc-lookahead", "50", "-profile:v", "baseline") // "-profile:v", "main", "-level", "4",
		} else {
			videoCommands = append(videoCommands, "-map_chapters", "-1", "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-profile:v", "baseline", "-level:v", "3.2")
		}
		// Added "-map_chapters", "-1" para filtrar los streams de data tipo cuetones en VANC.
		//videoCommands = append(videoCommands, "-c:v", codificacion.Nombre+nvenc)
	} else {
		videoCommands = append(videoCommands, "-map_chapters", "-1", "-c:v", "copy")
		codificacion.Nombre = videoFormat
	}
	generarArchivoM3U8(codificacionPath, programa, codificacion.Bitrates)

	//Modificado para diferenciar el argumento de -resize SI hay CUVID como preInput o -s si NO hay NVENC

	if configuration.Resolution != "" {
		if configuration.Hwaccel {
			preInputCommands = append(preInputCommands, "-resize", configuration.Resolution)
		} else {
			videoCommands = append(videoCommands, "-s", configuration.Resolution)
		}
	}

	//Modificado para poder hacer codec copy si decodificamos cuvid
	if configuration.Hwaccel {
		//modified as Hwaccel flag doesn´t seem to work as flag nvenc activated cuvid decoding as well (deactivate for 4:2:2 decoding)
		preInputCommands = append(preInputCommands, "-hwaccel", "cuvid", "-c:v", ""+videoFormat+"_cuvid")
		if codificacion.Nombre == "copy" {
			videoCommands = append(videoCommands, "-vf", "hwdownload,format=nv12")
		}
	} else {
		videoCommands = append(videoCommands, "-pix_fmt", "yuv420p") //Added for change format from 4:2:2 feeds to 4:2:0
	}

	//modificado para diferenciar el argumento de -deint && dss SI hay NVENC como preInput o -vf yadif si NO hay NVENC

	if configuration.Deint {
		if configuration.Hwaccel {
			preInputCommands = append(preInputCommands, "-deint", "2", "-drop_second_field", "1")
		} else {
			videoCommands = append(videoCommands, "-vf", "yadif")
		}
	}

	preInputCommands = append(preInputCommands, "-thread_queue_size", "128", "-rtbufsize", "128", "-i", input) //"-fflags", "nobuffer"

	singleVideoMap := true
	for _, bitrate := range codificacion.Bitrates {
		singleVideoMap = false
		maxbitrate := calcularBitrateMaximo(bitrate)
		videoCommands = append(videoCommands,
			"-map", videoMap, "-b:v", bitrate+"k", "-maxrate", maxbitrate+"k", "-an", "-sn")
		videoCommands = append(videoCommands, generarArgumentosHLS(codificacionPath+"/"+bitrate, codificacionPath+"/"+bitrate+"-segment-%03d", codificacion.Nombre)...)
		if !teletextSubtitles {
			videoCommands = append(videoCommands, generarArgumentosDVBSubs(codificacionPath, bitrate, codificacion.Nombre)...)
		}
	}
	if singleVideoMap {
		videoCommands = append(videoCommands, "-map", videoMap)
		videoCommands = append(videoCommands, generarArgumentosHLS(codificacionPath+"/"+"default", codificacionPath+"/"+"segment-%03d", codificacion.Nombre)...)
		if !teletextSubtitles {
			videoCommands = append(videoCommands, generarArgumentosDVBSubs(codificacionPath, "", codificacion.Nombre)...)
		}
	}

	var audioCommands []string
	for _, elem := range programa.audio {

		codificacionAudioPath := codificacionPath + "/audio/" + elem.language
		os.RemoveAll(codificacionAudioPath)
		os.MkdirAll(codificacionAudioPath, os.ModePerm)
		audioCommands = append(audioCommands, "-ac", "2")
		//audioFormat:=elem.format
		if configuration.Audio_codificacion != "" {
			if configuration.Audio_bitrate != "" {
				audioCommands = append(audioCommands, "-b:a", configuration.Audio_bitrate+"k")
			}
			audioCommands = append(audioCommands, "-c:a", configuration.Audio_codificacion)
			//audioFormat=configuration.audio_codificacion
		} else {
			audioCommands = append(audioCommands, "-c:a", "copy")
		}
		audioCommands = append(audioCommands, "-map", getMapId(elem))
		if configuration.Audio_codificacion != "" {
			if customEncoding {
				audioCommands = append(audioCommands, "-af", "aresample=async=1:first_pts=0")

			} else {
				//commandArgs = append(commandArgs, "-vsync","2")

				audioCommands = append(audioCommands, "-af", "aresample=async=1:first_pts=0")
			}
		}
		//commandArgs = append(commandArgs,
		//	"-ac", "2", "-b:a", "192k", "-c:a", "aac", "-map", getMapId(elem), "-af", "aresample=async=1:first_pts=0", "-map", getMapId(programa.video), "-vn")
		audioCommands = append(audioCommands, "-map", getMapId(programa.video), "-vn")
		audioCommands = append(audioCommands, generarArgumentosHLS(codificacionAudioPath+"/audio", codificacionAudioPath+"/audio%01d", codificacion.Nombre)...)

	}

	var subsCommands []string
	if teletextSubtitles {
		for _, elem := range programa.subtitulos {
			subsPath := codificacionPath + "/subtitulos/" + elem.language
			os.RemoveAll(subsPath)
			os.MkdirAll(subsPath, os.ModePerm)
			//subsCommands = append(subsCommands,"-map", getMapId(programa.audio[0]), "-map", getMapId(elem), "-map", getMapId(programa.video), "-filter_complex", "["+ getMapId(programa.video) +"]hwdownload,format=nv12;["+ getMapId(elem) +"]["+ getMapId(programa.video) +"]scale2ref=ih*mdar:ih [sub][ref],[ref][sub]overlay=(W-w)/2:(H-h)/2,hwupload_cuda", "-vn", "-scodec", "webvtt")  //, ["+  getMapId(programa.video) +"]overlay"
			//subsCommands = append(subsCommands,"-map", getMapId(programa.audio[0]), "-map", getMapId(elem), "-map", getMapId(programa.video), "-filter_complex", "["+ getMapId(programa.video) +"]hwdownload,format=nv12;["+ getMapId(elem) +"]scale="+ configuration.Resolution +",["+ getMapId(programa.video) +"]overlay,hwupload_cuda", "-vn", "-scodec", "webvtt")  //, ["+  getMapId(programa.video) +"]overlay"
			//subsCommands = append(subsCommands,"-map", getMapId(programa.audio[0]), "-map", getMapId(elem), "-map", getMapId(programa.video), "-filter_complex", "["+ getMapId(elem) +"]scale="+ configuration.Resolution +"[sub]", "-map", "[sub]", "-vn", "-scodec", "webvtt")  //, ["+  getMapId(programa.video) +"]overlay"
			subsCommands = append(subsCommands, "-map_chapters", "-1", "-map", getMapId(programa.audio[0]), "-map", getMapId(elem), "-map", getMapId(programa.video), "-vn", "-scodec", "webvtt")
			subsCommands = append(subsCommands, generarArgumentosHLS(subsPath+"/subtitulo", subsPath+"/reference%01d", codificacion.Nombre)...)
		}
	}

	commandArgs = append(commandArgs, preInputCommands...)
	//dvbSystem.preinputCommands=commandArgs
	commandArgs = append(commandArgs, postInputCommands...)
	commandArgs = append(commandArgs, videoCommands...)
	commandArgs = append(commandArgs, audioCommands...)
	commandArgs = append(commandArgs, subsCommands...)

	fmt.Printf("\nArgumentos ffmpeg: %s", commandArgs)
	fmt.Printf("\n")
	fmt.Printf("\n\nRecibiendo: %s", input)

	codiUrl := ""
	if customEncoding {
		codiUrl = codificacion.Nombre + "/"
	}
	if !configuration.Wn { //Manipulación de la ULR del stream de entrada para adaptarla al formato del archivo .json de salida. Estructurado para todo tipo de streaming 10/02/2020

		//Post Configuration
		var api_url string = "https://" + configuration.Server + "/api/custom/update_url"
		var user string = "titania"
		var pass string = "titanios"
		type urlData struct {
			Id       string
			Url_hls  string
			Url_m3u8 string
		}
		//POST data to JSON
		data := urlData{
			Id:       configuration.Tid,
			Url_hls:  "https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html",
			Url_m3u8: "https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8",
		}
		postData, err := json.Marshal(data)

		//POST Request
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("POST", api_url, bytes.NewBuffer(postData))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.SetBasicAuth(user, pass)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		//Printing POST RESULTS for tracing
				fmt.Printf("\n\nPost_Url: %s", api_url) //post url api
				fmt.Printf("\nPost_Status: %s", resp.Status) //post status response
				fmt.Printf("\njson: %s", string(postData)) //post data sent

				/*Translate Body Response from Titania
				response, err := ioutil.ReadAll(resp.Body)
		 		if err != nil {
		 			fmt.Println(err)
		 			log.Fatal(err)
		 		}
		 		fmt.Printf("\nPost_Response: %v", string(response))
				resp.Body.Close()*/

		//modificadas urls a https para seguridad en titania

		fmt.Printf("\n\nhttps://%s/%s/%s/%shls.m3u8", configuration.Ip, configuration.Url_subpath, configuration.MediaId, codiUrl)
		fmt.Printf("\nhttps://%s/%s/%s/%sindex.html", configuration.Ip, configuration.Url_subpath, configuration.MediaId, codiUrl)
		fmt.Printf("\n")
	} else {
		fmt.Printf("\nGenerando fragmentos en %s", configuration.Output)
	}
	cmd := exec.Command("ffmpeg", commandArgs...)
	resultado := "La codificación " + codificacion.Nombre + " ha finalizado."
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		resultado = err.Error() + ": " + stderr.String() + "\n" + resultado
	}
	channel <- resultado
}

// inicializarCodificaciones() ejecuta la función de ejecución de codificación dependiendo de los argumentos introducidos.
func InicializarCodificaciones(configuracion configuracion.Config, main_channel chan string) {
	configuration = configuracion
	channel := make(chan string)
	def := true
	encodings_number := 0
	if configuration.Test {
		obtenerStreamsPost() //saca streams por post
		//obtenerStreams() //saca streams por consola
		fmt.Println("\n")
	} else {
		for index, _ := range configuration.Encodings {
			def = false
			go inicializarFfmpegEncoding(channel, index)
			encodings_number++
		}
		if def {
			go inicializarFfmpegEncoding(channel, -1)
			encodings_number++
		}
		for i := 0; i < encodings_number; i++ {
			fmt.Println(<-channel)
		}
	}
	main_channel <- "Programa finalizado"
}

//Function added for hex to int to string conversion and show stream pids in decimal instead of hexadecimal
func hex2int(hexStr string) uint64 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 64)
	return uint64(result)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Write urls in jason format to file Deprecated as POST MEthod is working, need to check out why some url replacement doesnt work
/*
	var input_url string = input //cambiado udp_url a input_url para no confundir, ya que aceptamos más inputs (rtp,srt)
	if strings.Contains(input, "rtp") {
		rtp_1 := strings.NewReplacer("//", "", ":", "_",".","_")
		rtp_2 := rtp_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022Id\u0022:\u0022"+ configuration.Tid +"\u0022,\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ rtp_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "srt") {
		srt_1 := strings.NewReplacer("//", "", ":", "_",".","_")
		srt_2 := srt_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ srt_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "udp") {
		udp_1 := strings.NewReplacer("//", "", ":", "_",".","_","?fifo_size=557753&overrun_nonfatal=1","")
		udp_2 := udp_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ udp_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "rtsp") {
		rtsp_1 := strings.NewReplacer("//", "", ":", "_",".","_","/","_","-","_")
		rtsp_2 := rtsp_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ rtsp_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "rtmp") {
		rtmp_1 := strings.NewReplacer("//", "", ":", "_",".","_","/","_","-","_")
		rtmp_2 := rtmp_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ rtmp_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "https") {
		https_1 := strings.NewReplacer("//", "", ":", "_",".","_","/","_","-","_")
		https_2 := https_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ https_2 +".json", Output_Data, 0644)
		check(err)
	}else if strings.Contains(input, "http") {
		http_1 := strings.NewReplacer("//", "", ":", "_",".","_","/","_","-","_")
		http_2 := http_1.Replace(input_url)
		Output_Data := []byte("{\n\u0022m3u8_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "hls.m3u8\u0022,\n\u0022player_url\u0022:\u0022https://" + configuration.Ip + "/" + configuration.Url_subpath + "/" + configuration.MediaId + "/" + codiUrl + "index.html\u0022\n}")
		err := ioutil.WriteFile("/output_url/"+ configuration.Port +"_"+ http_2 +".json", Output_Data, 0644)
		check(err)
}*/

//Cambiada por obtenerStreamsPost
/*func obtenerStreams() map[string]*Programa {
	sub_lang_ok := len(configuration.Sub_lang) > 0
	audio_lang_ok := len(configuration.Audio_lang) > 0
	streams_ok := len(configuration.Streams) > 0

	udp := configuration.Input
	cmd := exec.Command("ffprobe","-hide_banner", udp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("FFprobe ha fallado: %s\n", out)
	}
	programas := make(map[string]*Programa)

	streamReg := regexp.MustCompile(`Stream #0:([0-9]+)(?:\[((?:0[xX])?[[:xdigit:]]+)\])?(?:\(([a-zA-Z]{3}(?:,[a-zA-Z]{3})*?)\))?: ([a-zA-Z]+): ([a-z0-9_]+)`)
	fpsReg := regexp.MustCompile(`,[ ]*([0-9.]+) fps[ ]*,`)
	progReg := regexp.MustCompile(`Program ([0-9]+)`)
	chromaReg := regexp.MustCompile(`,[ ]*yuv([0-9]+)p[ ]*`) //added for checking video format 422 o 420
	serviceName := regexp.MustCompile(`service_name[ ]*: (.+)`)
	lineas := strings.Split(string(out), "\n")
	programId := ""
	for i := 0; i < len(lineas); i++ {
		lineaActual := lineas[i]
		if progReg.MatchString(lineaActual) && !streams_ok {
			if programId != "" {
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
			programId = progReg.FindStringSubmatch(lineaActual)[1]
			if configuration.Id_programa == "" || configuration.Id_programa == programId {
				programa.id = programId
				if !configuration.Test {
					fmt.Printf("Servicios encontrados:\n")
				}
			}
			fmt.Printf("\n\tSID: %s",programId)
			programas[programId] = &Programa{id: programId}
		} else if serviceName.MatchString(lineaActual) && !streams_ok {
			if programId != "" {
				programas[programId].nombre = serviceName.FindStringSubmatch(lineaActual)[1]
				fmt.Printf("\n\tName: %s",programas[programId].nombre)
			}
		} else if streamReg.MatchString(lineaActual) {
			if programId == "" {
				programId = "-1"
				programas[programId] = &Programa{nombre: "No name", id: programId}
				fmt.Printf("\n\tSID: %s\n\tName: %s",programId, programas[programId].nombre)
				if configuration.Id_programa == "" {
					programa.id = programId
				}
			}
			programa := programas[programId]
			stream := streamReg.FindStringSubmatch(lineaActual)
			streamElement := Stream{id: stream[1], hexaId: stream[2], language: stream[3], format: stream[5]}

			if streams_ok {
				_, stream_ok := configuration.Streams[streamElement.hexaId]
				if !stream_ok {
					continue
				}
			}

			switch stream[4] {
			case "Video":
				streamElement.fps = fpsReg.FindStringSubmatch(lineaActual)[1]
				streamElement.chroma = chromaReg.FindStringSubmatch(lineaActual)[1]
				programa.video = streamElement
				//added lasts %s + streamElement.fps and %s + streamElement.chroma
				fmt.Printf("\n\t %s %s %s %s %s %s %s",stream[1],stream[2],stream[3],stream[4],stream[5],streamElement.fps,streamElement.chroma)
				break
			case "Audio":
				if !configuration.An {
					if len(streamElement.language) <= 0 {
						streamElement.language = "default"
					}
					ok := true
					if audio_lang_ok {
						_, ok = configuration.Audio_lang[streamElement.language]
					}
					for _, audio := range programa.audio {
						if streamElement.language == audio.language {
							ok = false
						}
					}
					if ok {
						programa.audio = append(programa.audio, streamElement)
						fmt.Printf("\n\t %s %s %s %s %s",stream[1],stream[2],stream[3],stream[4],stream[5])
					}
				}
				break
			case "Subtitle":
				if !configuration.Sn {
					if len(streamElement.language) <= 0 {
						streamElement.language = "default"
					}
					ok := true
					if sub_lang_ok {
						_, ok = configuration.Sub_lang[streamElement.language]
					}
					for _, audio := range programa.subtitulos {
						if streamElement.language == audio.language {
							ok = false
						}
					}
					if ok {
						programa.subtitulos = append(programa.subtitulos, streamElement)
						fmt.Printf("\n\t %s %s %s %s %s",stream[1],stream[2],stream[3],stream[4],stream[5])
					}
				}
				break
			}
		}
	}
	fmt.Printf("\n")
	if programaHash, ok := programas[programa.id]; ok {
		programa = *programaHash
	} else {
		log.Fatal("Error no existe el PID")
	}

	return programas
}*/
