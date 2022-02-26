options = {
  preload: 'auto',
  autoplay: true,
  fluid: false,
  controlBar: {
    volumePanel: { inline: false },
    children: [
      //'hdButton',
      'playToggle',
      'volumePanel',
      'liveDisplay',
      'ProgressControl',
      'volumeButton',
      'CustomControlSpacer',
      'subtitlesButton',
      'audioTrackButton',
      'fullscreenToggle'
    ]
  }
}
var player = videojs('videoplayer', options);

//var player = videojs('videoplayer',options); 

function supportsVideoType(type) {
  let video;
  let formats = {
    ogg: 'video/ogg; codecs="theora"',
    h264: 'video/mp4; codecs="avc1.42E01E"',
    webm: 'video/webm; codecs="vp8, vorbis"',
    vp9: 'video/webm; codecs="vp9"',
    hls: 'application/x-mpegURL; codecs="avc1.42E01E"',
    h265: 'video/hevc; codecs="hevc, aac"'
  };

  if (!video) {
    video = document.createElement('video')
  }

  return video.canPlayType(formats[type] || type);
}

var context;
var audioBuffer;
var sourceNode;
var splitter;
var analyser, analyser2;
var javascriptNode;

var canvas = document.getElementById('vumeter');
var ctx = canvas.getContext('2d');

var gradient = ctx.createLinearGradient(0,0,0,300);
gradient.addColorStop(1,'#000000');
gradient.addColorStop(0.75,'#ff0000');
gradient.addColorStop(0.25,'#ffff00');
gradient.addColorStop(0,'#ffffff');

//console.log(supportsVideoType('webm'))
player.on('loadedmetadata',function(){
  context=new AudioContext();
  setupAudioNodes()
  player.controlBar.addChild('hdButton',{},4)
})
player.ready();

function setupAudioNodes() {

  javascriptNode = context.createScriptProcessor(2048, 1, 1);
  javascriptNode.connect(context.destination);

  analyser = context.createAnalyser();
  analyser.smoothingTimeConstant = 0.3;
  analyser.fftSize = 1024;
  analyser2 = context.createAnalyser();
  analyser2.smoothingTimeConstant = 0.0;
  analyser2.fftSize = 1024;

  sourceNode = context.createMediaElementSource(document.getElementById("videoplayer_html5_api"));
  sourceNode.connect(context.destination)

  splitter = context.createChannelSplitter();
  sourceNode.connect(splitter);

  splitter.connect(analyser, 0, 0);
  splitter.connect(analyser2, 1, 0);

  analyser.connect(javascriptNode);
  sourceNode.connect(context.destination);

  javascriptNode.onaudioprocess = function () {
    var array = new Uint8Array(analyser.frequencyBinCount);
    analyser.getByteFrequencyData(array);
    var average = getAverageVolume(array);
    var array2 = new Uint8Array(analyser2.frequencyBinCount);
    analyser2.getByteFrequencyData(array2);
    var average2 = getAverageVolume(array2);
    ctx.clearRect(0, 0, 60, 130);
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 130 - average, 25, 130);
    ctx.fillRect(30, 130 - average2, 25, 130);
  }

  function getAverageVolume(array) {
    var values = 0;
    var average;
    var length = array.length;
    for (var i = 0; i < length; i++) {
      values += array[i];
    }
    average = values / length;
    return average;
  }
}
