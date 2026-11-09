let width = 100,
    height = 100,
    sizeInterval = 10,
    maxSize = 500;

let colors = ["Green", "Red", "Orange", "DodgerBlue", "MediumBlue", "Black", "White", "Yellow", "DarkOrchid", "DeepPink", "LimeGreen"];

let userColor;
let oldColor;

let svg = document.getElementById("svg"),
    wrapper = document.getElementById("wrapper"),
    ball = document.getElementById("path22");

let x = Math.floor(Math.random() * (parseInt(getComputedStyle(wrapper).width) - 10 - width / 2)) + 10 + width / 2,
    y = Math.floor(Math.random() * (parseInt(getComputedStyle(wrapper).height) - 10 - width / 2)) + 10 + width / 2,
    diry = Math.random() > 0.5 ? "down" : "up",
    dirx = Math.random() > 0.5 ? "right" : "left";

let yspeed = 4,
    xspeed = 4;

svg.style.display = "inherit";
svg.style.width = width+"px";
svg.style.height = height+"px";

window.addEventListener('focus', event => {
	for (let i = 0; i < 4; i++) {
		window.cancelAnimationFrame(animationRequest);
	}
	animationRequest = requestAnimationFrame(animate);
});

window.addEventListener('click', event => {
	for (let i = 0; i < 4; i++) {
		window.cancelAnimationFrame(animationRequest);
	}
	animationRequest = requestAnimationFrame(animate);
});

function randomColor(a) {
	let newColor;
	if (!oldColor) {
		newColor = a[Math.floor(Math.random() * a.length)];
		oldColor = newColor;
		return newColor;
	}
	do {
		newColor = a[Math.floor(Math.random() * a.length)];
	} while (oldColor === newColor);
	oldColor = newColor;

	return newColor;
}

function drawCoordinates(x, y) {
	x -= width / 2;
	y -= height / 2;
	svg.style.top = y+"px";
	svg.style.left = x+"px";
}

let animationRequest;

let ox = 0,
    oy = 0;
function animate(time, lastTime) {
	drawCoordinates(x, y);

	if (!lastTime) return requestAnimationFrame(newTime => animate(newTime, time));
	if (time-lastTime > 100) lastTime = time - 16.77

	let outBounds = false;
	let multiplier = 1000 / (time - lastTime);
	let wrap = getComputedStyle(wrapper);

	let yspeedNow = yspeed / multiplier * 60;
	let xspeedNow = xspeed / multiplier * 60;

	if (y + height / 2 - yspeedNow - 2 > parseInt(wrap.height)) {
		if (y + height / 2 - yspeedNow - 2 > parseInt(wrap.height) + width / 2) outBounds = true;
		ball.style.fill = userColor ? userColor : randomColor(colors);
		diry = "up";
	}

	if (y - height / 2 + yspeedNow + 2 < 0) {
		if (y - height / 2 + yspeedNow + 2 < 0 - height / 2) outBounds = true;
		ball.style.fill = userColor ? userColor : randomColor(colors);
		diry = "down";
	}

	if (x + width / 2 - xspeedNow - 2 > parseInt(wrap.width)) {
		if (x + width / 2 - xspeedNow - 2 > parseInt(wrap.width) + width / 2) outBounds = true;
		ball.style.fill = userColor ? userColor : randomColor(colors);
		dirx = "left";
	}

	if (x - width / 2 + xspeedNow + 2 < 0) {
		if (x - width / 2 + xspeedNow + 2 < 0 - width / 2) outBounds = true;
		ball.style.fill = userColor ? userColor : randomColor(colors);
		dirx = "right";
	}

	if (outBounds) {
		x = Math.floor(Math.random() * (parseInt(wrap.width) - 10 - width / 2)) + 10 + width / 2;
		y = Math.floor(Math.random() * (parseInt(wrap.height) - 10 - width / 2)) + 10 + width / 2;
	}
	ox = x, oy = y;
	if (dirx === "right") x = x + xspeedNow;
	if (dirx === "left") x = x - xspeedNow;
	if (diry === "down") y = y + yspeedNow;
	if (diry === "up") y = y - yspeedNow;

	animationRequest = requestAnimationFrame(newTime => animate(newTime, time));
}
animationRequest = requestAnimationFrame(animate);