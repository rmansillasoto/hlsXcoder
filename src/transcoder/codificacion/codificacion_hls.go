package codificacion

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

// generarArgumentosHLS es función que se encarga de generar los argumentos para codificar a HLS mediante
// el comando FFmpeg.
func generarArgumentosHLS(m3u8listpath string, segmentpath string, encoding string) []string {
	//fpsInt, _ := strconv.Atoi(programa.video.fps)
	//cambiado para v1.4.1 -> cambiado el fps de Int a float64 para contemplar decimales (NTSC dropped frame), redondear resultado y pasarlo a Int de nuevo
	fpsFloat, _ := strconv.ParseFloat(programa.video.fps, 64)
	fpsRound := math.Round(fpsFloat)
	fpsInt := int(fpsRound)
	hlsconfig := []string{"-flags", "+cgop", "-g", strconv.Itoa(fpsInt * configuration.Hls_time), "-hls_time", strconv.Itoa(configuration.Hls_time)}
	if configuration.Hls_fragments > 0 && !configuration.Disable_streaming {
		hlsconfig = append(hlsconfig, "-hls_list_size", strconv.Itoa(configuration.Hls_fragments), "-hls_flags", "delete_segments", "-hls_allow_cache", "0")
	} else {
		hlsconfig = append(hlsconfig, "-hls_list_size", "0")
	}
	output := []string{"-f", "hls", m3u8listpath + ".m3u8"}
	switch encoding {
	case "hevc":
		hlsconfig = append(hlsconfig, "-hls_segment_type", "fmp4", "-hls_segment_filename", segmentpath+".m4s")
		return append(hlsconfig, output...)
		break
		//case "h264":
		//	hlsconfig = append(hlsconfig, "-hls_segment_type", "fmp4", "-hls_segment_filename", segmentpath+".m4s")
		//	return append(hlsconfig, output...)
		//	break
	}
	hlsconfig = append(hlsconfig, "-hls_segment_filename", segmentpath+".ts")
	return append(hlsconfig, output...)
}

// generarArchivoM3U8(filePath string, programa Programa, bitrates []string) función que se encarga de crear el fichero M3U8 para la lista HLS.
func generarArchivoM3U8(filePath string, programa Programa, bitrates []string) {

	os.RemoveAll(filePath)
	os.MkdirAll(filePath, os.ModePerm)

	file, fileErr := os.Create(filePath + "/hls.m3u8")
	if fileErr != nil {
		log.Fatal("El archivo no se puede crear:", fileErr)
		os.Exit(1)
	}
	defer file.Close()
	teletextSubtitle := false
	var audio, subtitle, video string
	audio = ""
	subtitle = ""
	video = ""

	var buffer bytes.Buffer
	buffer.WriteString("#EXTM3U\n")
	for _, elem := range programa.audio {
		audio = ",AUDIO=\"audio\""
		buffer.WriteString("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE=\"" + elem.language + "\",NAME=\"" + elem.language + "\",URI=\"audio/" + elem.language + "/audio.m3u8\"\n") //cambiado elem.language x elem.hexaId
	}

	if len(programa.subtitulos) > 0 && programa.subtitulos[0].format == "dvb_teletext" {
		teletextSubtitle = true
	}

	if teletextSubtitle {
		for _, elem := range programa.subtitulos {
			subtitle = ",SUBTITLES=\"subs\""
			buffer.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"" + elem.language + "\",DEFAULT=NO,FORCED=NO,URI=\"subtitulo/" + elem.language + "/subtitulo_vtt.m3u8\",LANGUAGE=\"" + elem.language + "\"\n")
			//buffer.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"" + elem.language + "\",DEFAULT=NO,FORCED=NO,URI=default.m3u8,LANGUAGE=\"" + elem.language + "\"\n")
		}
	} else {
		for _, elem := range programa.subtitulos {
			video = ",VIDEO=\"videoSubs\""
			buffer.WriteString("#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"videoSubs\",NAME=\"" + elem.language + "\",DEFAULT=NO,FORCED=NO,URI=\"subtitulo/" + elem.language + "/subtitulo.m3u8\",LANGUAGE=\"" + elem.language + "\"\n")
		}
	}

	generarM3U8Videos := func(bitrate string, nombreM3U8 string) {
		nombreVideoM3U8 := "default"
		if bitrate == "" {
			bitrate = "1000"
		} else {
			nombreVideoM3U8 = bitrate
		}
		buffer.WriteString(
			"#EXT-X-STREAM-INF:BANDWIDTH=" + bitrate + "000" + subtitle + audio + video + " \n" +
				nombreVideoM3U8 + ".m3u8\n")
		if !teletextSubtitle {
			for _, elem := range programa.subtitulos {
				subtitle = ",SUBTITLES=\"subs\""
				buffer.WriteString(
					"#EXT-X-STREAM-INF:NAME=" + elem.language + "-" + bitrate + ",BANDWIDTH=" + bitrate + "000" + subtitle + audio + " \n " +
						"subtitulo/" + elem.language + "/" + nombreM3U8 + "\n")

				//buffer.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"" + elem.language + "\",DEFAULT=NO,FORCED=NO,URI=default.m3u8,LANGUAGE=\"" + elem.language + "\"\n")
			}
		}
	}

	videoDefault := true
	for _, bitrate := range bitrates {
		videoDefault = false
		generarM3U8Videos(bitrate, bitrate+"-subtitulo.m3u8")
	}
	if videoDefault {
		generarM3U8Videos("", "subtitulo.m3u8")

	}
	fmt.Fprintln(file, buffer.String())

}
