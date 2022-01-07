let canvas = document.getElementById("canvas");
let ctx = canvas.getContext("2d");

// OPTIONS 
let advx = false;
let advy = false;
let afps = 60
ctx.lineWidth = 2;
let pointlength = 5
let changeAmount = .2
let size = 20
let yval = "sin"
let xval = "cos"
let ymultiplier = 2
let xmultiplier = 1
let caregex = /(script)|console|alert+/gim;
let funcx = (val) => {
    if (!advx) {
         if (xval === "cos") {
        return Math.cos(val) * size * xmultiplier
    } else if (xval === "sin") {
        return Math.sin(val * xmultiplier) * size 
    } else if( xval === "tan") {
        return Math.tan(val) * size * xmultiplier
    }
    
    } else {
        if (!(caregex.test(advx))) {
             return parseInt(eval(advx))
        }
       
    }
   
}
let funcy = (val) => {
    if (!advy) {
        if (yval === "cos") {
        return Math.cos(val) * size * ymultiplier
    } else if (yval === "sin") {
        return Math.sin(val) * size * ymultiplier
    } else if( yval === "tan") {
        return Math.tan(val) * size * ymultiplier
    }
    } else {
         if (!(caregex.test(advy))) {
             return parseInt(eval(advx))
        }
       
    }
   
}

let animate;

$(document).change((event) => {
    console.log("changed")
})


$("#afps").change((event) => {
    clearInterval(animate)
    afps = parseFloat(event.target.value)
    animate = setInterval(function x() {
    drawdot(funcx(i), funcy(i), width / 2, width / 2)
    i += changeAmount
    color = HslGen()
    return x;
}(), 1000 / afps);
});
$("#linewidth").change(function(event){
   ctx.lineWidth =  parseFloat(event.target.value)
});
$("#pointlength").change(function(event){
    pointlength = parseFloat(event.target.value)
});
$("#changeAmount").change(function(event){
   changeAmount =  parseFloat(event.target.value)
});
$("#size").change(function(event){
    size =  parseFloat(event.target.value)
});
$("#xmult").change(function(event){
    xmultiplier =  parseFloat(event.target.value)
});
$("#ymult").change(function(event){
    ymultiplier =  parseFloat(event.target.value)
});
$('.yval div input[type=radio]').click(function () {
    yval = $('.yval input[type=radio]:checked').attr('id');
       console.log(yval)
});

$('.xval div input[type=radio]').click(function () {
    xval = $('.xval input[type=radio]:checked').attr('id');
    console.log(xval)
});

$('#ady').change(function(event){
    let val = event.target.value
    if (val.length > 0) {
      advy = val  
    }
})

$('#adx').change(function(event){
    let val = event.target.value
    if (val.length > 0) {
      advx = val
    }

})


//-------------

    console.log("Started animating")
    // methods
let updateSize = function (canvas, operator) {
    if (operator === "width") {
        return canvas.width;
    } else if (operator === "height") {
        return canvas.height;
    }
}

let clearCanvas = function () {
    ctx.clearRect(0, 0, width, height);
}

let HslGen = function () {
    num += .1;
    if (num > 360) {
        num = num - 360
    }
    return `hsl(${num}, 100%, 40%)`
}

let getFPS = () => {
    return new Promise((resolve) => {
        requestAnimationFrame((timeFrame1) => {
            requestAnimationFrame((timeFrame2) => {
                resolve(1000 / (timeFrame2 - timeFrame1));
            });
        });
    });
}
//-------------------------------------------------


//Static variables, dont touch
let width = updateSize(canvas, "width")
let height = updateSize(canvas, "height")
let rows = 1000;
let columns = 1000;
let rectSize = 4.5;
let color = "rgb(0,0,0)";
let num = Math.floor(Math.random() * (360 - 0 + 1))
let fps = getFPS();
let dx, dy, odx, ody;
let points = []
//-------------------------------

// Drawing method itself
let drawdot = function (gdx, gdy, originx, originy) {
    clearCanvas()
    dx = (gdx) * rectSize;
    dy = (gdy) * rectSize;
    if (points.length + 1 > pointlength) {
        points.shift()
        points.push([dx, dy])
    } else {
        points.push([dx, dy])
    }
    
    if (points.length > pointlength) {
        while (true) {
            points.shift()
            if (!(points.length > pointlength)) {
                break;
            }
        }
    }
    ctx.strokeStyle = color;
    ctx.moveTo(points[0][0], points[0][1])
    ctx.beginPath();
    points.forEach((point, index) => {
        if (index === 0) {
            return
        }
        let tx = point[0] - points[index - 1][0]
        let ty = point[1] - points[index - 1][1]
        if (tx > 1000 || tx < -1000) {
            odx = dx
            return
        }
        if (ty > 1000 || ty < -1000) {
            ody = dy
            return
        }
        ctx.moveTo(originx + points[index - 1][0], originy + points[index - 1][1])
        ctx.lineTo(originx + point[0], originy + point[1])
    })
    ctx.stroke()
    ctx.closePath()
}
// --------------------


// Animator
let i = 0;
animate = setInterval(function x() {
    drawdot(funcx(i), funcy(i), width / 2, width / 2)
    i += changeAmount
    color = HslGen()
    return x;
}(), 1000 / afps);
// -----------------

