var width = 50;
var height = 50;
var sizeInterval = 10;
var maxSize = 500;

var canvas = document.getElementById('canvas');
var ctx = document.getElementById('canvas').getContext('2d');
var img = document.getElementById('i3');

var sizeUp = document.getElementById('plusb')
var sizeDown = document.getElementById('minusb')
var downloadButton = document.getElementById('dlbtn')
var sizeValueElem = document.getElementById('sizevalue')
var clearButton = document.getElementById('clrbtn')

var cursorElem = document.createElement('img')
cursorElem.id = 'imageCursor'
cursorElem.src = imgSrc()
cursorElem.style.display = 'none'
cursorElem.style.pointerEvents = 'none'
cursorElem.style.opacity = '0.3'

document.body.appendChild(cursorElem)

canvas.addEventListener('mousedown', event => {
	if (event.button !== 0) return;
	getPosition(event, drawCoordinates)
	window.addEventListener('mousemove', moved)
})

canvas.addEventListener('touchmove', event => {
	event.preventDefault()
	if (event.touches.length === 1) {
		var rect = canvas.getBoundingClientRect();
		var x = event.touches[0].clientX - rect.left;
		var y = event.touches[0].clientY - rect.top;
		drawCoordinates(x,y)
	}
	
})

canvas.addEventListener('wheel', event => {
	if (event.deltaY > 0) {
		width -= sizeInterval;
		height -= sizeInterval;
		if (width < sizeInterval) {
			width = sizeInterval;
			height = sizeInterval;
		}
		sizeValueElem.value = width;
		drawCursor(event.pageX, event.pageY)
	} else if (event.deltaY < 0) {
		width += sizeInterval;
		height += sizeInterval;
		if (width > maxSize) {
			width = maxSize;
			height = maxSize;
		}
		sizeValueElem.value = width;
		drawCursor(event.pageX, event.pageY)
	}
})

window.addEventListener('resize', resizeCanvas)

document.addEventListener('keydown', event => {
	let keys = ["1","2","3","4","5","6","7","8","9","0","+"]
	let ind = keys.indexOf(event.key.toString())
	let useInd = ind +1;
	if (ind !== -1) {
		img = document.getElementById('i' + useInd);
		cursorElem.src = imgSrc()
		 document.querySelector('#t'+useInd).checked = true
	}
})

document.addEventListener('mousemove', event => {
	if (event.target.id === 'canvas') {
		drawCursor(event.pageX, event.pageY)
	} else {
		cursorElem.style.display = 'none'
	}
})

sizeUp.addEventListener('click', event => {
	width += sizeInterval;
	height += sizeInterval;
	if (width > maxSize) {
		width = maxSize;
		height = maxSize;
	}
	sizeValueElem.value = width;
})

sizeDown.addEventListener('click', event => {
	width -= sizeInterval;
	height -= sizeInterval;
	if (width < sizeInterval) {
		width = sizeInterval;
		height = sizeInterval;
	}
	sizeValueElem.value = width;
})

sizeValueElem.addEventListener('change', event => {
	let val = parseInt(event.target.value)
	if (val < sizeInterval) {
		width = sizeInterval;
		height = sizeInterval;
	} else if (val > maxSize) {
		width = maxSize
		height = maxSize
	} else {
		width = val
		height = val
	}
	sizeValueElem.value = width;
})

downloadButton.addEventListener('click', () => {
	var d = canvas.toDataURL("image/png");
	var w = window.open('about:blank', 'Your masterpiece');
	w.document.write("<img src='" + d + "' alt='Masterpiece from canvas'/>");
})

clearButton.addEventListener('click', () => {
	ctx.clearRect(0, 0, 10000, 10000)
})

for (let i = 0; i < document.querySelectorAll('[type=radio]').length; i++) {
	document.querySelectorAll('[type=radio]')[i].addEventListener('click', () => {
		var id = document.querySelector('input[type=radio]:checked').id.replace(/[^0-9]/, '')
		img = document.getElementById('i' + id);
		cursorElem.src = imgSrc()
	})
}

function imgSrc() {
	return img.src
}

function resizeCanvas() {
	canvas.width = parseInt(getComputedStyle(canvas).width)
	canvas.height = parseInt(getComputedStyle(canvas).height)
}

function moved(event) {
	if (event.buttons === 0) return window.removeEventListener('mousemove', moved)
	getPosition(event, drawCoordinates)
}

function getPosition(event, callback) {
	var rect = canvas.getBoundingClientRect();
	var x = event.clientX - rect.left;
	var y = event.clientY - rect.top;

	callback(x, y);
}

function drawCursor(x, y) {
	cursorElem.width = width
	cursorElem.height = height
	x -= cursorElem.width / 2
	y -= cursorElem.height / 2
	cursorElem.style.display = 'inherit'
	cursorElem.style.position = 'absolute'
	cursorElem.style.top = y + 'px'
	cursorElem.style.left = x + 'px'

}

function drawCoordinates(x, y) {
	var canvasWidth = getComputedStyle(canvas).width
	if (canvasWidth === '400px') {
		x *= 2;
		y *= 2;
	}
	x = x - (width / 2);
	y = y - (height / 2);

	ctx.drawImage(img, x, y, width, height);
}

 resizeCanvas()