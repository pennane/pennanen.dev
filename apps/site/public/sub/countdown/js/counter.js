const setTrueVh = () => document.documentElement.style.setProperty('--vh', `${window.innerHeight * 0.01}px`)

const hasPassed = (d) => d < Date.now()
const timeDiff = (d, ad) => {
	let dist = Math.abs(d - ad);
    let days = Math.floor(dist / (1000 * 60 * 60 * 24));
    let hours = Math.floor((dist % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    let minutes = Math.floor((dist % (1000 * 60 * 60)) / (1000 * 60));
    let seconds = Math.floor((dist % (1000 * 60)) / 1000);
	return {d:days, h:hours, m:minutes, s:seconds}
}

const genTimeSpan = ({d = 0,h = 0,m = 0,s = 0}) => {
	let html = ``;
		html += `<span class="days"><span class="number">${d}</span><span class="identifier">Days</span></span>`
		html += `<span class="hours"><span class="number">${h}</span><span class="identifier">Hours</span></span>`
		html += `<span class="minutes"><span class="number">${m}</span><span class="identifier">Minutes</span></span>`
		html += `<span class="seconds"><span class="number">${s}</span><span class="identifier">Seconds</span></span>`
	return html
}


let upper = document.getElementById("uppercontent")
let middle = document.getElementById("middlecontent")
let lower = document.getElementById("lowercontent")

const init = async () => {
    setTrueVh()
    window.addEventListener("resize", setTrueVh)
	
    let response = await fetch('assets/date.json');
    let {date, content} = await response.json();
	let dateTo = new Date(`${date.month} ${date.day}, ${date.year} ${date.time}:00`).getTime()
	
	let loop = setInterval(()=> {
		let dif = timeDiff(dateTo, Date.now())
		if (hasPassed(dateTo)) {
			upper.textContent = `${content.title} was`
			middle.innerHTML = genTimeSpan(dif) + `<span class="identifier> ago.</span>`
			lower.textContent = `ago.`
		} else {
			upper.innerHTML = content.title +" in"
			middle.textContent = ""
			lower.innerHTML =genTimeSpan(dif)
		}
	}, 1000)
}
(() => {init()})()
