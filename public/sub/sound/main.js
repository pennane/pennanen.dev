import { useAnalyzer, startAnalysing } from './analyser.js'
import { sineWave, squareWave, whiteNoise } from './waveform.js'

const gapEl = document.getElementById('gap')
const layersEl = document.getElementById('amount')
const startEl = document.getElementById('start')
const whitenoiseEl = document.getElementById('whitenoise')
const stopEl = document.getElementById('off')
const resetEl = document.getElementById('reset')

let start = 160,
    amount = 10,
    gap = 80,
    polyphony = 3

const audioContext = new AudioContext({ sampleRate: 44100 })
const audioAnalyser = audioContext.createAnalyser()
const gainNode = audioContext.createGain()
gainNode.gain.value = 0.1

gainNode.connect(audioAnalyser)
gainNode.connect(audioContext.destination)

useAnalyzer(audioAnalyser)

let sources = []
let additionalSource

function playSound({ audioContext, array }) {
    const audioBuffer = audioContext.createBuffer(1, array.length, audioContext.sampleRate)
    audioBuffer.copyToChannel(array, 0)

    const source = audioContext.createBufferSource()
    source.loop = true
    source.buffer = audioBuffer
    source.connect(gainNode)
    source.start()

    return source
}

function addSource({ audioContext, array, additional }) {
    const source = playSound({ audioContext, array })

    if (additional) {
        additionalSource = source
        return
    }

    sources.push(source)

    while (sources[polyphony]) {
        sources[0].stop()
        sources.shift()
    }
}

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

function toggleWhiteNoise(audioContext) {
    if (additionalSource) {
        additionalSource.stop()
        additionalSource = undefined
        return
    }
    const array = whiteNoise(audioContext)

    addSource({ array, audioContext, additional: true })
}

function createButton(audioContext, hz) {
    const array = new Float32Array(audioContext.sampleRate)
    const sampleGenerator = sineWave(audioContext, hz)

    for (let i = 0; i < array.length; i++) {
        array[i] = sampleGenerator.next().value
    }

    const button = document.createElement('button')
    button.textContent = hz + ' hz'
    button.addEventListener('click', () => addSource({ audioContext, array }))
    return button
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

whitenoiseEl.addEventListener('click', () => toggleWhiteNoise(audioContext))
stopEl.addEventListener('click', () => clearSources())
resetEl.addEventListener('click', () => createButtons())
document.addEventListener('keydown', (e) => {
    if (e.target.nodeName === 'INPUT' || e.repeat) return
    if (e.key === 'r' || e.key === 'R') {
        clearSources()
        return
    }
    if (e.key === 'w' || e.key === 'W') {
        toggleWhiteNoise(audioContext)
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

export { start, gap, amount }

createButtons(true)
startAnalysing()
