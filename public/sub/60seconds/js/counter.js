const setTrueVh = () => document.documentElement.style.setProperty('--vh', `${window.innerHeight * 0.01}px`)

const base = 60

const secondsElement = document.getElementById('time')
const textElement = document.getElementById('text')

const now = Date.now()
const hash = parseFloat(window.location.hash.replace('#', '')) || null

const time = hash ? hash * base : base
const end = new Date(Date.now() + time * 1000)

textElement.textContent = 'seconds.'
secondsElement.textContent = parseInt((end - now) / 1000)

let loop = setInterval(() => {
    const now = Date.now()
    const remaining = parseInt((end - now) / 1000)
    if (end > now) {
        textElement.textContent = 'seconds.'
        secondsElement.textContent = remaining
        document.title = `${remaining}seconds.`
    } else {
        secondsElement.textContent = 'time'
        textElement.textContent = `\xa0ran out.`
        document.title = `0seconds.`
        clearInterval(loop)
    }
}, 80)

setTrueVh()
window.addEventListener('resize', setTrueVh)
