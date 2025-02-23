import responsiveCanvas from "./responsiveCanvas.js";

let el = document.getElementById('canvas')
let responsive = new responsiveCanvas(el, {
	mousemove: true,
	translate: true,
	resize: false
})
let canvas = responsive.canvas
let ctx = responsive.ctx


let currentx;
let currenty;
let un1;
let un2;
let userN1 = false;
let userN2 = false;

let a1container = document.getElementById("a1")
let a2container = document.getElementById("a2")
let n1container = document.getElementById("n1")
let n2container = document.getElementById("n2")

class Material {
	constructor(name, n, style) {
		this.n = n <= 0 || n >= 10 ? 1 : n
		this.style = style
		this.name = name
	}
}

let materials = {
	water: new Material("water", 1.331, "#3a53ba"),
	air: new Material("air", 1.0002758, "#b7cddc"),
	vacuum: new Material("vacuum", 1, "white"),
	glass: new Material("glass", 1.508, "lightcyan"),
	diamond: new Material("diamond", 2.454, "steelblue")
}

let materialsEl = document.getElementById("materials")
let materialsEl2 = document.getElementById("materials2")
let nEl1 = document.getElementById("n1")
let nEl2 = document.getElementById("n2")

Object.keys(materials).forEach(material => {
	{
		let option = document.createElement("option")
		option.value = material
		option.id = materials[material].name + 1
		option.textContent = materials[material].name
		materialsEl.appendChild(option)
	}
	{
		let option = document.createElement("option")
		option.value = material
		option.id = materials[material].name + 2
		option.textContent = materials[material].name
		materialsEl2.appendChild(option)
	}
})

document.getElementById("air1").selected = true

let material1 = materials.air
let material2 = materials.water

materialsEl.addEventListener("change", event => {
	material1 = materials[event.target.value]
	draw(currentx | -200, currenty | -200)
	userN1 = false
})

materialsEl2.addEventListener("change", event => {
	material2 = materials[event.target.value]
	draw(currentx | -200, currenty | -200)
	userN2 = false
})

nEl1.addEventListener("change", event => {
	nVal = parseFloat(event.target.value)
	userN1 = true
	if (nVal <= 0 || nVal > 10) {
		un1 = 1
		event.target.value = 1
	} else {
		un1 = nVal
	}
	draw(currentx | -200, currenty | -200)
})

nEl2.addEventListener("change", event => {
	nVal = parseFloat(event.target.value)
	userN2 = true
	if (nVal <= 0 || nVal > 10) {
		un2 = 1
		event.target.value = 1
	} else {
		un2 = nVal
	}

	draw(currentx | -200, currenty | -200)
})

canvas.addEventListener("mousedown", event => {
	if (event.button !== 0) return;
	canvas.className = "grabbing"
	draw(responsive.coordinates[0], responsive.coordinates[1])
	window.addEventListener("mousemove", moved)
})

canvas.addEventListener("mouseup", event => {
	if (event.button !== 0) return
	canvas.className = "grab"
})

function moved(event) {
	if (event.buttons === 0) return window.removeEventListener("mousemove", moved)
	draw(responsive.coordinates[0], responsive.coordinates[1])
}

function draw(x, y) {
	let n1 = userN1 ? un1 : material1.n
	let n2 = userN2 ? un2 : material2.n

	currentx = x;
	currenty = y
	clearCanvas()
	drawBg()

	let angle1 = Math.atan(x / y)
	let angle2 = NaN;
	let inverted = false;
	let xmult = 1;
	let ymult = 1;
	let rad90 = toRadians(90);
	let n12;

	if (x > 0) {
		angle1 = -angle1
		xmult = -1;
	}

	if (y < 0) {
		n12 = (n2 / n1)
	} else {
		n12 = (n1 / n2)
		ymult = -1;
	}

	let angle2a = Math.sin(angle1) / n12
	angle2 = Math.asin(angle2a)

	if (angle2a < 1 && angle2a > -1) {
		drawAngle(1)
		drawAngle(2)
		drawMiddleLine()
		drawLine(
			ymult * xmult * (1000 * Math.cos(rad90 - angle2)),
			ymult * (1000 * Math.sin(rad90 - angle2)),
			0,
			0
		)
		drawLine(x, y, 0, 0)
		if (n12 !== 1) drawLine(ymult * xmult * 1000 * Math.cos(rad90 - angle1), ymult * -1000 * Math.sin(rad90 - angle1), 0, 0, "rgba(255,50,50,0.2)")

	} else {
		drawAngle(1)
		drawAngle(2)
		drawMiddleLine()
		drawLine(x, y, 0, 0)
		if (n12 !== 1) drawLine(ymult * xmult * 1000 * Math.cos(rad90 - angle1), ymult * -1000 * Math.sin(rad90 - angle1), 0, 0, "rgba(255,50,50,0.2)")
	}

	function drawAngle(n) {
		ctx.beginPath()
		ctx.moveTo(0, 0)
		if (n === 1) {

			if (xmult === 1 && ymult === 1) {
				ctx.arc(0, 0, 30, toRadians(-90 - (toDegrees(angle1))), toRadians(-90), false)
			} else if (xmult === -1 && ymult === -1) {
				ctx.arc(0, 0, 30, -toRadians(-90 - (toDegrees(angle1))), -toRadians(-90), false)
			} else if (xmult === -1 && ymult === 1) {
				ctx.arc(0, 0, 30, toRadians(-90 + (toDegrees(angle1))), toRadians(-90), true)
			} else if (xmult === 1 && ymult === -1) {
				ctx.arc(0, 0, 30, -toRadians(-90 - (-toDegrees(angle1))), -toRadians(-90), true)
			}

		} else {
			if (xmult === 1 && ymult === 1) {
				ctx.arc(0, 0, 25, toRadians(90 - (toDegrees(angle2))), toRadians(90), false)
			} else if (xmult === -1 && ymult === -1) {
				ctx.arc(0, 0, 25, -toRadians(90 - (toDegrees(angle2))), -toRadians(90), false)
			} else if (xmult === -1 && ymult === 1) {
				ctx.arc(0, 0, 25, toRadians(90 + (toDegrees(angle2))), toRadians(90), true)
			} else if (xmult === 1 && ymult === -1) {
				ctx.arc(0, 0, 25, -toRadians(90 + (toDegrees(angle2))), -toRadians(90), true)
			}
		}
		ctx.lineWidth = "1"
		ctx.strokeStyle = "black";
		ctx.closePath();
		ctx.stroke();
	}

	function updateTable() {
		a1container.textContent = Math.round(toDegrees(angle1 < 0 ? -angle1 : angle1) * 10) / 10 + `\u02DA`
		a2container.textContent = Math.round(toDegrees(angle2 < 0 ? -angle2 : angle2) * 10) / 10 + `\u02DA`
		n1container.value = userN1 ? un1 : material1.n
		n2container.value = userN2 ? un2 : material2.n
	}
	updateTable();
}

function drawLine(fx, fy, tx, ty, style) {
	ctx.beginPath()
	ctx.strokeStyle = style ? style : "red"
	ctx.moveTo(fx, fy)
	ctx.lineTo(tx, ty)
	ctx.lineWidth = 3
	ctx.stroke()
	ctx.closePath()
}

function drawBg() {
	ctx.beginPath();
	ctx.rect(-canvas.width / 2, -canvas.height / 2, canvas.width, canvas.height / 2)
	ctx.fillStyle = material1.style
	ctx.fill()
	ctx.beginPath();
	ctx.rect(-canvas.width / 2, 0, canvas.width, canvas.height / 2)
	ctx.fillStyle = material2.style
	ctx.fill()
	ctx.closePath()
}

function drawMiddleLine() {
	ctx.beginPath();
	ctx.setLineDash([10, 5]); /*dashes are 5px and spaces are 3px*/
	ctx.moveTo(0, -canvas.height / 2);
	ctx.lineTo(0, canvas.height);
	ctx.strokeStyle = "white"
	ctx.stroke();
	ctx.setLineDash([])
	ctx.closePath()
}

function clearCanvas() {
	ctx.clearRect(-canvas.width / 2, -canvas.height / 2, canvas.width, canvas.height);
}

function toDegrees(angle) {
	return angle * (180 / Math.PI);
}

function toRadians(angle) {
	return angle * (Math.PI / 180);
}

draw(-200, -200)