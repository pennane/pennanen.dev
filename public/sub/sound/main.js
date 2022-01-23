const gapEl = document.getElementById('gap')
const layersEl = document.getElementById('amount')
const startEl = document.getElementById('start')

const canvas = document.getElementById('visualizer')
canvas.width = canvas.offsetWidth
canvas.height = canvas.offsetHeight
const context = canvas.getContext('2d')

function clampBetween(input, min, max) {
    return input < min ? min : input > max ? max : input
}

function mapBetween(current, in_min, in_max, out_min, out_max) {
    const mapped = ((current - in_min) * (out_max - out_min)) / (in_max - in_min) + out_min
    return clampBetween(mapped, out_min, out_max)
}

let start = 160,
    amount = 10,
    gap = 80,
    polyphony = 3

const resizeObserver = new ResizeObserver((entries) => {
    const { width, height } = entries[0].contentRect
    canvas.width = width
    canvas.height = height

    console.log('Size changed')
})

resizeObserver.observe(canvas)

const audioContext = new AudioContext({ sampleRate: 44100 })
const sampleRate = audioContext.sampleRate

const analyser = audioContext.createAnalyser()
analyser.fftSize = 2048

const gainNode = audioContext.createGain()
gainNode.gain.value = 0.1
gainNode.connect(audioContext.destination)
gainNode.connect(analyser)

let sources = []
let additionalSource

let data = new Uint8Array(analyser.frequencyBinCount)

function draw(data) {
    context.clearRect(0, 0, canvas.width, canvas.height) //x,y,width,height
    let startValue = Math.floor(mapBetween(Math.max(start - 2 * gap, 0), 0, 22000, 0, data.length - 1))
    let endValue = Math.floor(mapBetween(Math.min(start + amount * gap + 2 * gap, 22000), 0, 22000, 0, data.length - 1))
    let space = canvas.width / (endValue - startValue)
    context.beginPath()
    context.moveTo(0, canvas.height)
    let line = 1
    for (let i = startValue; i < endValue; i++) {
        const value = data[i]
        context.lineTo(space * line, canvas.height - value)
        line++
    }
    context.lineTo(canvas.width, canvas.height)
    context.stroke()
}

function animation() {
    analyser.getByteFrequencyData(data)
    draw(data)
    requestAnimationFrame(animation)
}

requestAnimationFrame(animation)

function clearSources() {
    for (const source of sources) {
        source.stop()
    }
    sources = []
    if (additionalSource) {
        additionalSource.stop()
        additionalSource = undefined
    }
}

function* sineWave(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.sin(i / (sampleFrequency / (Math.PI * 2)))
    }
}

function playSound({ audioContext, array, additional }) {
    const audioBuffer = audioContext.createBuffer(1, array.length, audioContext.sampleRate)
    audioBuffer.copyToChannel(array, 0)
    // clearSources()
    const source = audioContext.createBufferSource()
    source.loop = true
    source.connect(gainNode)
    source.buffer = audioBuffer
    source.start()
    if (additional) {
        additionalSource = source
    } else {
        sources.push(source)
    }

    while (sources[polyphony]) {
        sources[0].stop()
        sources.shift()
    }
}

function createButton(audioContext, hz) {
    const array = new Float32Array(audioContext.sampleRate)
    const sampleGenerator = sineWave(audioContext, hz)

    for (i = 0; i < array.length; i++) {
        array[i] = sampleGenerator.next().value
    }

    const button = document.createElement('button')
    button.textContent = hz + ' hz'
    button.addEventListener('click', () => playSound({ audioContext, array }))
    return button
}

function whiteNoise(audioContext) {
    if (additionalSource) {
        additionalSource.stop()
        additionalSource = undefined
        return
    }
    const array = new Float32Array(audioContext.sampleRate).map(() => Math.random())

    playSound({ array, audioContext, additional: true })
}

function createButtons(initial) {
    if (!initial) {
        gap = Math.max(gapEl.value, 1)
        start = Math.min(Math.max(startEl.value, 40), 22000)
        polyphony = Math.min(Math.max(layersEl.value, 1), 100)
    }

    gapEl.value = gap
    startEl.value = start
    layersEl.value = polyphony

    const target = document.getElementById('target')
    while (target.firstChild) {
        target.removeChild(target.firstChild)
    }

    const container = document.createElement('div')

    let hz = start
    for (let i = 0; i < amount; i++) {
        if (hz + i * gap > 22000) continue
        container.appendChild(createButton(audioContext, hz + i * gap))
    }

    target.appendChild(container)
}

createButtons(true)

document.getElementById('whitenoise').addEventListener('click', () => whiteNoise(audioContext))
document.getElementById('off').addEventListener('click', () => clearSources())
document.getElementById('reset').addEventListener('click', () => createButtons())

document.addEventListener('keydown', (e) => {
    if (e.target.nodeName === 'INPUT' || e.repeat) return
    if (e.key === 'r' || e.key === 'R') {
        clearSources()
        return
    }
    if (e.key === 'w' || e.key === 'W') {
        whiteNoise(audioContext)
        return
    }
    let index = parseInt(e.key)
    if (isNaN(index)) return
    if (index === 0) index = 10
    const button = document.querySelectorAll('#target div button')[index - 1]
    if (!button) return
    button.click()
    button.classList.add('pressed')
})

document.addEventListener('keyup', (e) => {
    if (e.target.nodeName === 'INPUT' || e.repeat) return
    let index = parseInt(e.key)
    if (isNaN(index)) return
    if (index === 0) index = 10
    const button = document.querySelectorAll('#target div button')[index - 1]
    if (!button) return
    button.classList.remove('pressed')
})
