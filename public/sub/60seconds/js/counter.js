const setTrueVh = () => document.documentElement.style.setProperty('--vh', `${window.innerHeight * 0.01}px`)

let timeEl = document.getElementById("time")
let textEl = document.getElementById("text")
let now = Date.now()

let hash = parseFloat(window.location.hash.replace("#", "")) || false;
let base = 60;

let time = hash ? hash * base : base;
let end = new Date(Date.now() + time * 1000)

textEl.textContent = "seconds."
timeEl.textContent = parseInt((end - now) / 1000)

let loop = setInterval(() => {
	let now = Date.now()
	let s = parseInt((end - now) / 1000)
	if (end > now) {
		textEl.textContent = "seconds."
		timeEl.textContent = s
		document.title = `${s}seconds.`
	} else {
		timeEl.textContent = "time"
		textEl.textContent = `\xa0ran out.`
		document.title = `0seconds.`
		clearInterval(loop)
	}
}, 80)

setTrueVh()
window.addEventListener("resize", setTrueVh)