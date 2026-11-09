//main.js web portfolio sivulle
// 2018
"use strict";
var sound1 = new Audio("sound/Laser_Shoot15_01.mp3");
var sound2 = new Audio("sound/rocketlaunch.mp3");
var sound3 = new Audio("sound/rocketlaunch.mp3");


window.lintuboolean = 0;
window.rakettiboolean = 0;
window.globtoggle = 0;
window.laivaboolean = 0;

$(document).ready(function () {

    $("img").attr("draggable", false);

    $(window).bind("scroll", function (e) {
        parallaxScroll();
    });

    function parallaxScroll() {

        var rullattu = $(window).scrollTop();

        $("#parallaksi1").css("top", (0 - (rullattu * 0.25)) + "px");
        $("#parallaksi2").css("top", (0 - (rullattu * 0.5)) + "px");
        $("#parallaksipohja").css("top", (0 - (rullattu * 0.42)) + "px");
        $("#parpar2").css("top", (0 - (rullattu * 0.492)) + "px");
        $("#parpar3").css("top", (0 - (rullattu * 0.488)) + "px");

        if (window.lintuboolean !== 3) {
            $("#lintupassing").css("top", (0 - (rullattu * 0.36)) + "px");
        }
        if (window.rakettiboolean === 0) {
            $("#parallaksi3").css("top", (0 - (rullattu * 0.75)) + "px");
        }
    }

    $("a[href='#globdestination']").click(function () {
        $("html, body").animate({
            scrollTop: $(document).height()
        }, 9000);
        return false;
    });

    $("a[href='#wrapper']").click(function () {
        $("html, body").animate({
            scrollTop: 0 - ($(document).height())
        }, 12000);
        return false;
    });

    var fadeStart = 700;
    var fadeUntil = 1900;

    $(document).ready(function () {
        $(window).bind("scroll", function () {
            var offset = $(document).scrollTop();
            var opacity = 1;
            var opacity2 = 0;
            if (offset <= fadeStart) {
                opacity = 0;
                opacity2 = 1;
            } else if (offset <= fadeUntil) {
                opacity = 0 + ((offset - fadeStart) / 500);
                opacity2 = 1 - ((offset - fadeStart) / 500);
            }

            $(".pilvi").css("opacity", opacity);
            $(".avaruus").css("opacity", opacity2);
        });
    });


    var fStart = 1200; // Start transition 100px from top
    var fEnd = 1500; // End at 500px
    var gStart = [9, 21, 57]; // Space BLue
    var gEnd = [63, 166, 251]; // LSky Bluye
    var gDiff = [gEnd[0] - gStart[0], gEnd[1] - gStart[1], gEnd[2] - gStart[2]];


    var korkeus = $(document).height();
    var globkorkeus = korkeus - 200;

    $(document).scroll(function () {
        var p = ($(this).scrollTop() - fStart) / (fEnd - fStart);
        p = Math.min(1, Math.max(0, p));
        var cBg = [Math.round(gStart[0] + gDiff[0] * p), Math.round(gStart[1] + gDiff[1] * p), Math.round(gStart[2] + gDiff[2] * p)];
        $("body").css("background-color", "rgb(" + cBg.join(",") + ")");

        if (($(window).scrollTop() + $(window).height() >= globkorkeus) && (window.globtoggle === 0)) {
            window.globtoggle = 1;
            document.getElementById("globsound").play();
            $("#globsound").animate({
                volume: 0.7
            }, 2000);
            $(".globcave").animate({
                opacity: 1
            }, 1000);
        }

        if (($(window).scrollTop() + $(window).height() < globkorkeus) && (window.globtoggle === 1)) {
            window.globtoggle = 0;
            $("#globsound").animate({
                volume: 0
            }, 1000);
            $(".globcave").animate({
                opacity: 0
            }, 1000);
            setTimeout(function () {
                document.getElementById("globsound").pause();
            }, 1000);
        }
    });
});


function rakettifc() {
    window.rakettiboolean++;
    sound2.play();
    $("#parallaksi3").animate({
        top: "-1000"
    }, 4000);
    setTimeout(function () {
        window.open("http://safeos.net/arttu/AdventureOfNoggerborn", "_self");
    }, 3750);
}

function lintuded() {
    window.lintuboolean++;
    sound1.play();
    if (window.lintuboolean === 3) {
        $("#lintu").attr("src", "img/lintu150_static.gif");
        $("#lintupassingcontent").css("webkitAnimationPlayState", "paused");
        window.lintuboolean = 3;
        $("#lintupassing").animate({
            top: "10000"
        }, 3000);
        setTimeout(function () {
            document.getElementById("lintupassing").remove();
        }, 3000);
    }
}

function laivafc() {
    if (window.laivaboolean === 0) {
        window.laivaboolean++;
        sound3.play();
        $("#p2-laiva").attr("src", "img/pixlaiva.gif");
        $("#parpar2").css("opacity", 0);
        $("#parpar3").css("opacity", 0);
        $("#p2-laiva").animate({
            left: "-1000"
        }, 5000);
        setTimeout(function () {
            $("#p2-laiva").css("left", ($(window).width() + 250));
            $("#p2-laiva").animate({
                left: "300"
            }, 4000);
            setTimeout(function () {
                $("#p2-laiva").attr("src", "img/pixlaiva2.png");
                $("#parpar2").css("opacity", 1);
                $("#parpar3").css("opacity", 1);
                window.laivaboolean = 0;
            }, 5000);
        }, 5100);


    }

}
