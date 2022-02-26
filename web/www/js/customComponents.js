videojs.selectAuto=function(){

    player.tech_.hls.representations().forEach(element => {
        let sub = element.id.split("/")[0]
        if (sub !== "subtitulo") {
            element.enabled(true)
        } else {
            element.enabled(false)
        }
    });
}


var MenuItem = videojs.getComponent('MenuItem');
customMenuItem= videojs.extend(MenuItem, {

    constructor: function (player, options, onClickListener) {
        this.onClickListener = onClickListener;
        MenuItem.call(this, player, options);
        this.on('click', this.onClick);
        this.on('touchstart', this.onClick);

    },
    onClick: function () {
        this.onClickListener(this);
        var selected = this.options_.el.id;
        player.tech_.hls.representations().forEach(element => {
            if (selected === "auto") {
                videojs.selectAuto()
            } else {
                if (element.id === selected) {
                    element.enabled(true)
                } else {
                    element.enabled(false)
                }
            }
        });
    }
});
videojs.registerComponent("CustomMenuItem", customMenuItem);


var Button = videojs.getComponent("MenuButton");
hdButton = videojs.extend(Button, {
    constructor: function (player, options) {
        Button.call(this, player, options);
        this.controlText('Quality');
        var staticLabel = document.createElement('span');
        staticLabel.classList.add('vjs-levels-button-staticlabel');
        this.el().appendChild(staticLabel);
        videojs.selectAuto()
    },

    createItems: function () {

        var component = this;
        var player = component.player();
        var levels = player.tech_.hls.representations();
        var item;
        var menuItems = [];

        if (!levels.length || levels.length <= 1) {
            return [];
        }
        levels = [{
            name: 'Auto',
            index: -1,
            id: "auto"
        }].concat(levels);


        var onClickUnselectOthers = function (clickedItem) {
            menuItems.forEach(function (item) {
                if (item.el().classList.contains('vjs-selected')) {
                    item.el().classList.remove('vjs-selected');
                }
            });
            clickedItem.el().classList.add('vjs-selected');
        };

        return levels.map(function (level, index) {
            var levelName;

            if (level.name) {
                levelName = level.name;
            } else if (level.id) {
                levelName = level.id;
                let temp = level.id.split(".")[0]
                let tempArray = temp.split(/[/-]+/)
                if (tempArray.length <= 1) {
                    levelName = temp
                    if (!isNaN(temp)) {
                        levelName = levelName + "kb"
                    }
                } else {
                    if (tempArray[0] === "subtitulo") {
                        levelName = "sub/" + tempArray[1]
                        if (!isNaN(tempArray[2])) {
                            levelName = levelName + "/" + tempArray[2] + 'kb'
                        }
                    } else {
                        levelName = temp
                    }

                }
            } else {
                //levelName = Math.round(level.bitrate / 1000) + ' Kbps';
                if (level.bandwidth) {
                    levelName = (Math.round(level.bandwidth / 1024) + 'kb');
                } else {
                    return null;
                }
            }

            customMenuItem = videojs.getComponent("CustomMenuItem")

            item = new customMenuItem(player, {
                el: videojs.dom.createEl('li', {
                    label: levelName,
                    value: level.bandwidth,
                    id: level.id,
                    class: 'vjs-menu-item vjs-button'
                })
            }, onClickUnselectOthers);


            menuItems.push(item);

            if (level.name === 'Auto') {
                item.el().classList.add('vjs-selected');
            }
            item.el().innerText = levelName;

            return item;
        });
    },
    buildCSSClass: function () {
        return "vjs-icon-cog vjs-menu-button vjs-menu-button-popup vjs-control vjs-button";
    }
});
videojs.registerComponent("hdButton", hdButton);