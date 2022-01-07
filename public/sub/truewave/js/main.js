let canvas = document.getElementById("canvas");
let ctx = canvas.getContext("2d");

let transX = 0;
let transY = canvas.height * 0.5;

ctx.translate(transX, transY);

let freqSlider = document.getElementById("freq");
let ampSlider = document.getElementById("amp");

let points = [];
let animationRequest;
let accTime = 1;

let amp = canvas.height / 2 - 15;
let freq = 1;
let wavelen = canvas.offsetWidth - 30;

ampSlider.addEventListener("input", event => {
    amp = event.target.value;
});

freqSlider.addEventListener("input", event => {
    freq = event.target.value;
});

const sine = step => Math.sin(step * Math.PI * 2);

const drawFromYpoints = (ctx, points) => {
    ctx.beginPath();
    ctx.lineWidth = 3;
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    points.forEach((y, x) => {
        if (x === 0) ctx.moveTo(x + 10, y);
        else ctx.lineTo(x + 10, y);
    });
    ctx.stroke();
};

const clearCanvas = () =>
    ctx.clearRect(0, -canvas.height / 2, canvas.width, canvas.height);

const background = () => {
    ctx.beginPath();
    ctx.fillStyle = "blue";
    ctx.rect(0, -canvas.height / 2, canvas.width, canvas.height);
    ctx.fill();
};

const animate = (time, lastTime) => {
    background();
    if (!lastTime) return requestAnimationFrame(newTime => animate(newTime, time));
    if (time - lastTime > 100) lastTime = time - 16.77;
    accTime += time - lastTime;
    let y = sine((accTime % (1000 / freq)) / (1000 / freq)) * amp;
    points.unshift(y);
    points.length = wavelen;
    drawFromYpoints(ctx, points);
    animationRequest = requestAnimationFrame(newTime => animate(newTime, time));
};

animationRequest = requestAnimationFrame(animate);