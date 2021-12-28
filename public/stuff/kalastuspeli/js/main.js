var ifirsttime = Cookies.get("firsttime");
var cookiesloaded = 0;

/* Variables */

/*scores*/
var fishscore = 0,
    fishmanscore = 0,
    fishhutscore = 0,
    fishvesselscore = 0,
    fishflockscore = 0,
    fishmagicscore = 0;

/*      */
var fishman = 0,
    fishhut = 0,
    fishvessel = 0,
    fishflock = 0,
    fishmagic = 0,
    totalyield = 0,
    totaltotalyield = 0,
    nE = 0,
    fisknum = 0,
    autoclicker = 0;


/*prices*/
var fishmanprice = parseInt(20),
    fishhutprice = parseInt(500),
    fishvesselprice = parseInt(2000),
    fishflockprice = parseInt(50000),
    fishmagicprice = parseInt(1000000);

/*utility lucrativity*/
var fishmanyield = 0.20,
    fishhutyield = 4,
    fishvesselyield = 40,
    fishflockyield = 100,
    fishmagicyield = 10000,
    fishyield = 1;

/*upgrades*/
var upg1 = 10000,
    upg2 = 10000,
    upg3 = 20000,
    upg4 = 50000,
    upg5 = 10000,
    upg6 = 10000000,
    upg7 = 500000,
    upg8 = 1000;

var upg1b = 1,
    upg2b = 0,
    upg3b = 0,
    upg4b = 0,
    upg5b = 0,
    upg6b = 0,
    upg7b = 1,
    upg8b = 0;
/* ---------------------------- */




$(document).ready(function () {

    $("#resetbtn").click(function () {
        if (confirm('Are you sure?')) {
            abortTimer();
            ckremove();
            Cookies.remove("firsttime");
            setTimeout(function () {
                location.reload();
            }, 3000);
        } else {
            // Do nothing!
        }
    });

    $('#prognappi').click(function () {
        $('#progress').css("opacity", "0");
        $('#stats').css("opacity", "1");
        $('#progress').css("z-index", "-1");
        $('#stats').css("z-index", "auto");
        console.log("stats esiin");
        $('#prognappi').addClass('active');
        $('#statsnappi').removeClass('active');
    });

    $('#statsnappi').click(function () {
        $('#progress').css("opacity", "1");
        $('#stats').css("opacity", "0");
        $('#progress').css("z-index", "auto");
        $('#stats').css("z-index", "-1");
        $('#statsnappi').addClass('active');
        $('#prognappi').removeClass('active');
        console.log("progress esiin");
    });

    $("#fishingbtn").click(function () {
        fisknum++;
        if ($("body").css("background-image").includes("/img/fullartbg.gif")) {
            $("body").css("background-image", "url(./img/catch.gif)");
        }

        addscore('fish');
        floatingnumbers();
        setTimeout(function () {
            var lastfisk = fisknum;
            setTimeout(function () {
                if (lastfisk === fisknum) {
                    $("body").css("background-image", "url(./img/fullartbg.gif)");
                }
            }, 1000);
        }, 1300);
    });

    $(".purchasebtn").click(function (e) {
        purchase($(e.target).attr("value"));

    });

    $(".upgrade").click(function (e) {
        purchase(($(e.target).attr('id')));
    });



    if (ifirsttime === "no") {
        console.log("Cool, you are coming back!");
        ckload();
        setTimeout(function () {
            cookiesloaded = 1;
        }, 2000);
    } else {
        console.log("So a first timer, hmm.");
        Cookies.set("firsttime", "no");
        setTimeout(function () {
            cookiesloaded = 1;
        }, 2000);
        $(".attention").css("opacity", "1");
        setTimeout(function () {
            $(".attention").animate({
                opacity: '0'
            }, 5500);
        }, 10000);
    }

});


function floatingnumbers() {
    nE++;
    $("#clickbtnholder").append("<span class='floatingnumber' id='num" + nE + "'>" +
        (parseInt(fishyield * upg1b * upg7b)) + "</span>");
    $("#num" + nE).animate({
        top: "-=300px",
        opacity: '0'
    }, 1500);
}

function purchase(val) {
    switch (val) {
    case 'fishman':
        if (fishscore >= fishmanprice) {
            fishscore = fishscore - fishmanprice;
            fishmanprice = parseInt(fishmanprice * 1.2);
            fishmanscore++;
            updatestats();
        }
        break;

    case 'fishhut':
        if (fishscore >= fishhutprice) {
            fishscore = fishscore - fishhutprice;
            fishhutprice = parseInt(fishhutprice * 1.2);
            fishhutscore++;
            updatestats();
        }
        break;

    case 'fishvessel':
        if (fishscore >= fishvesselprice) {
            fishscore = fishscore - fishvesselprice;
            fishvesselprice = parseInt(fishvesselprice * 1.2);
            fishvesselscore++;
            updatestats();
        }
        break;

    case 'fishflock':
        if (fishscore >= fishflockprice) {
            fishscore = fishscore - fishflockprice;
            fishflockprice = parseInt(fishflockprice * 1.2);
            fishflockscore++;
            updatestats();
        }
        break;

    case 'fishmagic':
        if (fishscore >= fishmagicprice) {
            fishscore = fishscore - fishmagicprice;
            fishmagicprice = parseInt(fishmagicprice * 1.2);
            fishmagicscore++;
            updatestats();
        }
        break;
    case 'upg1':

        if ((fishscore >= upg1 && upg1b == 1)) {
            fishscore = fishscore - upg1;
            updatestats();
            upg1b = 2;
        }
        break;
    case 'upg2':

        if ((fishscore >= upg2 && upg2b == 0)) {
            fishscore = fishscore - upg2;
            fishmanyield = fishmanyield * 2;
            updatestats();
            upg2b = 1;
        }
        break;
    case 'upg3':

        if ((fishscore >= upg3 && upg3b == 0)) {
            fishscore = fishscore - upg3;
            fishhutyield = fishhutyield * 2;
            updatestats();
            upg3b = 1;
        }
        break;
    case 'upg4':
        if ((fishscore >= upg4 && upg4b == 0)) {
            fishscore = fishscore - upg4;
            fishvesselyield = fishvesselyield * 2;
            updatestats();
            upg4b = 1;
        }
        break;
    case 'upg5':
        if ((fishscore >= upg5 && upg5b == 0)) {
            fishscore = fishscore - upg5;
            fishflockyield = fishflockyield * 2;
            updatestats();
            upg5b = 1;
        }
        break;
    case 'upg6':
        if ((fishscore >= upg6 && upg6b == 0)) {
            fishscore = fishscore - upg6;
            fishmagicyield = fishmagicyield * 2;
            updatestats();
            upg6b = 1;
        }
        break;
    case 'upg7':
        if ((fishscore >= upg7 && upg7b == 1)) {
            fishscore = fishscore - upg7;
            updatestats();
            upg7b = 4;
        }
        break;
    case 'upg8':
        if ((fishscore >= upg8 && upg8b == 0)) {
            fishscore = fishscore - upg8;
            upg8b = 1;
        }
        break;

    default:
        error("wrongpurchasetype");
    }

}


function addscore(val) {
    switch (val) {
    case 'fish':
        fishscore = parseInt(fishscore + (fishyield * upg1b * upg7b));
        $(".fishstat").text(fishscore);
        break;

    case 'fishman':
        fishmanscore++;
        $("#fishmanstat").text(fishmanscore);
        break;

    case 'fishhut':
        fishhutscore++;
        $("#fishhutstat").text(fishhutscore);
        break;

    case 'fishvessel':
        fishvesselscore++;
        $("#fishvesselstat").text(fishvesselscore);
        break;

    case 'fishflock':
        fishflockscore++;
        $("#fishflockstat").text(fishflockscore);
        break;

    case 'fishmagic':
        fishmagicscore++;
        $("#fishmagicstat").text(fishmagicscore);
        break;


    default:
        error("wrongtype");
    }

}

var tid = setTimeout(fishpersec, 1000);

function fishpersec() {
    totalyield = (fishmanyield * fishmanscore) + (fishhutyield * fishhutscore) + (fishvesselyield * fishvesselscore) + (fishflockyield * fishflockscore) + (fishmagicyield * fishmagicscore);

    totaltotalyield = totalyield * upg1b * upg7b;

    totaltotalyield = Math.round(totaltotalyield * 100) / 100;
    fishscore = fishscore + totaltotalyield;
    fishscore = Math.round(fishscore * 100) / 100;
    if (fishscore > 10000000) {
        $(".fishstat").text(fishscore.toExponential());
    } else {
        $(".fishstat").text(fishscore);
    }
    if (totalyield > 1000000) {
        $("#fishpersecstat").text(totaltotalyield.toExponential());
    } else {
        $("#fishpersecstat").text(totaltotalyield);
    }
    updatestats();
    $('span[style$="opacity: 0;"]').remove();

    if ((!document.getElementById("autoclicker")) && upg8b == 1) {
        var autoele = "<div id='autoclicker' class='utility'><span>Autoclicker </span><button id='autoOnOff'>On / Off</button> </div>";
        $("#utilitiescontent").append(autoele);
        console.log("autoclicker added");
        $("#autoOnOff").click(function (e) {
            if (autoclicker === 0) {
                autoclicker = 1;
            } else {
                autoclicker = 0;
            }
            autoclickerloop();
        });
    }

    if (document.getElementById("autoclicker") && autoclicker === 1) {
        $("#autoOnOff").text("On");
        $("#autoOnOff").css("filter", "");
    }
    if (document.getElementById("autoclicker") && autoclicker === 0) {
        $("#autoOnOff").css("filter", "brightness: 40");
        $("#autoOnOff").text("Off");
    }




    tid = setTimeout(fishpersec, 1000);
}

function autoclickerloop() {
    setTimeout(function () {
        $('#fishingbtn').click();
        if (autoclicker === 1) {
            autoclickerloop();
        }
    }, 250);
}





var cookietimer = setTimeout(cktimer, 8000);

function cktimer() {
    if (cookiesloaded === 1) {
        ckremove();
        setTimeout(function () {

        }, 2000);
        ckset();
    }

    cookietimer = setTimeout(cktimer, 8000);
}




function abortTimer() {
    clearTimeout(tid);
    clearTimeout(cookietimer);
}


function updatestats() {


    /* Scores */

    $(".fishstat").text(fishscore);
    $("#fishmanstat").text(fishmanscore);
    $("#fishhutstat").text(fishhutscore);
    $("#fishvesselstat").text(fishvesselscore);
    $("#fishflockstat").text(fishflockscore);
    $("#fishmagicstat").text(fishmagicscore);


    /* Prices */

    $("#fishmanprice").text(fishmanprice);
    $("#fishhutprice").text(fishhutprice);
    $("#fishvesselprice").text(fishvesselprice);
    $("#fishflockprice").text(fishflockprice);
    $("#fishmagicprice").text(fishmagicprice);

    /* Per secs */

    $("#fishmanpersec").text(fishmanyield * fishmanscore);
    $("#fishhutpersec").text(fishhutyield * fishhutscore);
    $("#fishvesselpersec").text(fishvesselyield * fishvesselscore);
    $("#fishflockpersec").text(fishflockyield * fishflockscore);
    $("#fishmagicpersec").text(fishmagicyield * fishmagicscore);

}

function error(val) {
    switch (val) {
    case 'wrongtype':
        console.log("Error: " + "Wrong type of value inserted.");
        break;

    case 'wrongpurchasetype':
        console.log("Error: " + "Tried to purchase unknown type of utility.");
        break;

    default:
        console.log("Error:" + "Unknown error");
    }
}


/* Cookie business */
/* Cookies */

function ckload() {
    /* Cookies load */
    fishscore = parseFloat(Cookies.get('fishscoreck'));
    fishmanscore = parseFloat(Cookies.get('fishmanscoreck'));
    fishhutscore = parseFloat(Cookies.get('fishhutscoreck'));
    fishvesselscore = parseFloat(Cookies.get('fishvesselscoreck'));
    fishflockscore = parseFloat(Cookies.get('fishflockscoreck'));
    fishmagicscore = parseFloat(Cookies.get('fishmagicscoreck'));

    /*      */
    fishman = parseFloat(Cookies.get('fishscoreck'));
    fishhut = parseFloat(Cookies.get('fishshutck'));
    fishvessel = parseFloat(Cookies.get('fishvesselck'));
    fishflock = parseFloat(Cookies.get('fishflockck'));
    fishmagic = parseFloat(Cookies.get('fishmagicck'));
    totalyield = parseFloat(Cookies.get('totalyieldck'));
    nE = parseFloat(Cookies.get('nEck'));

    /*prices*/
    fishmanprice = parseFloat(Cookies.get('fishmanpriceck'));
    fishhutprice = parseFloat(Cookies.get('fishhutpriceck'));
    fishvesselprice = parseFloat(Cookies.get('fishvesselpriceck'));
    fishflockprice = parseFloat(Cookies.get('fishflockpriceck'));
    fishmagicprice = parseFloat(Cookies.get('fishmagicpriceck'));

    /*utility lucrativity*/
    fishmanyield = parseFloat(Cookies.get('fishmanyieldck'));
    fishhutyield = parseFloat(Cookies.get('fishhutyieldck'));
    fishvesselyield = parseFloat(Cookies.get('fishvesselyieldck'));
    fishflockyield = parseFloat(Cookies.get('fishflockyieldck'));
    fishmagicyield = parseFloat(Cookies.get('fishmagicyieldck'));
    fishyield = parseFloat(Cookies.get('fishyieldck'));

    upg1b = Cookies.get('upg1bck');
    upg2b = Cookies.get('upg2bck');
    upg3b = Cookies.get('upg3bck');
    upg4b = Cookies.get('upg4bck');
    upg5b = Cookies.get('upg5bck');
    upg6b = Cookies.get('upg6bck');
    upg7b = Cookies.get('upg7bck');
    upg8b = Cookies.get('upg8bck');


}

function ckset() {
    /* Cookies set */
    Cookies.set('fishscoreck', fishscore);
    Cookies.set('fishmanscoreck', fishmanscore);
    Cookies.set('fishhutscoreck', fishhutscore);
    Cookies.set('fishvesselscoreck', fishvesselscore);
    Cookies.set('fishflockscoreck', fishflockscore);
    Cookies.set('fishmagicscoreck', fishmagicscore);

    /*      */
    Cookies.set('fishscoreck', fishscore);
    Cookies.set('fishshutck', fishhut);
    Cookies.set('fishvesselck', fishvessel);
    Cookies.set('fishflockck', fishflock);
    Cookies.set('fishmagicck', fishmagic);
    Cookies.set('totalyieldck', totalyield);
    Cookies.set('nEck', nE);

    /*prices*/
    Cookies.set('fishmanpriceck', fishmanprice);
    Cookies.set('fishhutpriceck', fishhutprice);
    Cookies.set('fishvesselpriceck', fishvesselprice);
    Cookies.set('fishflockpriceck', fishflockprice);
    Cookies.set('fishmagicpriceck', fishmagicprice);

    /*utility lucrativity*/
    Cookies.set('fishmanyieldck', fishmanyield);
    Cookies.set('fishhutyieldck', fishhutyield);
    Cookies.set('fishvesselyieldck', fishvesselyield);
    Cookies.set('fishflockyieldck', fishflockyield);
    Cookies.set('fishmagicyieldck', fishmagicyield);
    Cookies.set('fishyieldck', fishyield);

    Cookies.set('upg1bck', upg1b);
    Cookies.set('upg2bck', upg2b);
    Cookies.set('upg3bck', upg3b);
    Cookies.set('upg4bck', upg4b);
    Cookies.set('upg5bck', upg5b);
    Cookies.set('upg6bck', upg6b);
    Cookies.set('upg7bck', upg7b);
    Cookies.set('upg8bck', upg8b);
    /* ---------------------------- */
}

function ckremove() {
    /* Cookies remove */
    Cookies.remove('fishscoreck');
    Cookies.remove('fishmanscoreck');
    Cookies.remove('fishhutscoreck');
    Cookies.remove('fishvesselscoreck');
    Cookies.remove('fishflockscoreck');
    Cookies.remove('fishmagicscoreck');

    /*      */
    Cookies.remove('fishscoreck');
    Cookies.remove('fishshutck');
    Cookies.remove('fishvesselck');
    Cookies.remove('fishflockck');
    Cookies.remove('fishmagicck');
    Cookies.remove('totalyieldck');
    Cookies.remove('nEck');

    /*prices*/
    Cookies.remove('fishmanpriceck');
    Cookies.remove('fishhutpriceck');
    Cookies.remove('fishvesselpriceck');
    Cookies.remove('fishflockpriceck');
    Cookies.remove('fishmagicpriceck');

    /*utility lucrativity*/
    Cookies.remove('fishmanyieldck');
    Cookies.remove('fishhutyieldck');
    Cookies.remove('fishvesselyieldck');
    Cookies.remove('fishflockyieldck');
    Cookies.remove('fishmagicyieldck');
    Cookies.remove('fishyieldck');

    Cookies.remove('upg1bck');
    Cookies.remove('upg2bck');
    Cookies.remove('upg3bck');
    Cookies.remove('upg4bck');
    Cookies.remove('upg5bck');
    Cookies.remove('upg6bck');
    Cookies.remove('upg7bck');
    Cookies.remove('upg8bck');

}
