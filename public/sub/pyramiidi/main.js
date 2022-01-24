import { useAnalyzer, startAnalysing } from './analyser.js'
import { sineWaveGenerator, createWhiteNoise, squareWaveGenerator } from './waveform.js'
import f from './frequencies.js'

const gapEl = document.getElementById('gap')
const layersEl = document.getElementById('amount')
const startEl = document.getElementById('start')
const whitenoiseEl = document.getElementById('whitenoise')
const stopEl = document.getElementById('off')
const resetEl = document.getElementById('reset')
const precisionEl = document.getElementById('precision')

let start = 40,
    amount = 10,
    gap = 90,
    polyphony = 1,
    precision = 14,
    precisionComputed = 2 ** precision,
    visualizerSmoothing = 0,
    gain = 0.1,
    sampleRate = 44100,
    soundGenerator = squareWaveGenerator

const audioContext = new AudioContext({ sampleRate })
const audioAnalyser = audioContext.createAnalyser()
const gainNode = audioContext.createGain()

audioAnalyser.fftSize = precisionComputed
audioAnalyser.smoothingTimeConstant = visualizerSmoothing
gainNode.gain.value = gain
gainNode.gain.max

gainNode.connect(audioAnalyser)
gainNode.connect(audioContext.destination)

useAnalyzer(audioAnalyser)

let sources = []
let additionalSource

function playSound({ audioContext, array }) {
    let source
    if (array[0] instanceof Float32Array) {
        source = []
        for (const innerArray of array) {
            const audioBuffer = audioContext.createBuffer(1, innerArray.length, audioContext.sampleRate)
            audioBuffer.copyToChannel(innerArray, 0)

            const innerSource = audioContext.createBufferSource()
            innerSource.loop = true
            innerSource.buffer = audioBuffer
            innerSource.connect(gainNode)
            innerSource.start()
            source.push(innerSource)
        }
    } else {
        const audioBuffer = audioContext.createBuffer(1, array.length, audioContext.sampleRate)
        audioBuffer.copyToChannel(array, 0)
        source = audioContext.createBufferSource()
        source.loop = true
        source.buffer = audioBuffer
        source.connect(gainNode)
        source.start()
    }

    return source
}

function removeSurplusSources() {
    while (sources[polyphony]) {
        if (sources[0] instanceof Array) {
            for (const source of sources[0]) {
                source.stop()
            }
        } else {
            sources[0].stop()
        }

        sources.shift()
    }
}

function addSource({ audioContext, array, additional }) {
    const source = playSound({ audioContext, array })

    if (additional) {
        additionalSource = source
        return
    }

    sources.push(source)

    removeSurplusSources()
}

function clearSources() {
    while (sources[0]) {
        if (sources[0] instanceof Array) {
            for (const source of sources[0]) {
                source.stop()
            }
        } else {
            sources[0].stop()
        }

        sources.shift()
    }

    if (additionalSource) {
        additionalSource.stop()
        additionalSource = undefined
    }
}

function createHoverSource(hz) {
    const array = new Float32Array(audioContext.sampleRate)
    const sampleGenerator = soundGenerator(audioContext, hz)

    for (let i = 0; i < array.length; i++) {
        array[i] = sampleGenerator.next().value
    }

    addSource({ audioContext, array })
}

function toggleWhiteNoise(audioContext) {
    if (additionalSource) {
        additionalSource.stop()
        additionalSource = undefined
        return
    }
    const array = createWhiteNoise(audioContext)

    addSource({ array, audioContext, additional: true })
}

function getFrequencyArray(hz) {
    const freqArray = new Float32Array(audioContext.sampleRate)

    const sampleGenerator = soundGenerator(audioContext, Math.round(hz))

    for (let i = 0; i < freqArray.length; i++) {
        freqArray[i] = sampleGenerator.next().value
    }
    return freqArray
}

function createButton(audioContext, hz) {
    let array
    if (Array.isArray(hz)) {
        array = []
        for (const freq of hz) {
            array.push(getFrequencyArray(freq))
        }
    } else {
        array = getFrequencyArray(hz)
    }

    const button = document.createElement('button')
    button.textContent = `${hz[0] || hz} hz`
    button.addEventListener('click', () => addSource({ audioContext, array }))
    return button
}

function setUserDefinedValues() {
    gapEl.value = gap
    startEl.value = start
    layersEl.value = polyphony
    precisionEl.value = precision
}

function magicButtons() {
    const mag = [
        [f.FS2, f.CS3, f.FS3, f.AS3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.FS4],
        [f.A2, f.E3, f.G3, f.CS4, f.FS4],
        [f.G2, f.B2, f.D3, f.G3, f.B3, f.G4]
    ].map((hz) => {
        return createButton(audioContext, hz)
    })
    return mag
}

function readUserDefinedValues() {
    gap = Math.max(gapEl.value, 1)
    start = Math.min(Math.max(startEl.value, 40), 22000)
    polyphony = Math.min(Math.max(layersEl.value, 1), 100)
    precision = Math.min(Math.max(precisionEl.value, 5), 15)
    precisionComputed = 2 ** precision
    audioAnalyser.fftSize = precisionComputed
}

function updateButtons(initial) {
    if (!initial) {
        readUserDefinedValues()
    }

    setUserDefinedValues()

    removeSurplusSources()

    const target = document.getElementById('target')

    while (target.firstChild) {
        target.removeChild(target.firstChild)
    }

    const container = document.createElement('div')

    const buttons = magicButtons()
    buttons.forEach((b) => container.appendChild(b))
    target.appendChild(container)

    // let hz = start
    // for (let i = 0; i < amount; i++) {
    //     if (hz + i * gap > 22000) continue
    //     container.appendChild(createButton(audioContext, hz + i * gap, i))
    // }

    // target.appendChild(container)
}

function handleKeydown(e) {
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
}

function handleKeyup(e) {
    if (e.target.nodeName === 'INPUT' || e.repeat) return
    let index = parseInt(e.key)
    if (isNaN(index)) return
    if (index === 0) index = 10
    const button = document.querySelectorAll('#target div button')[index - 1]
    if (!button) return
    button.classList.remove('pressed')
}

whitenoiseEl.addEventListener('click', () => toggleWhiteNoise(audioContext))
stopEl.addEventListener('click', () => clearSources())
resetEl.addEventListener('click', () => updateButtons())
document.addEventListener('keydown', handleKeydown)
document.addEventListener('keyup', handleKeyup)
;[...document.querySelectorAll('input')].forEach((el) => el.addEventListener('change', () => updateButtons()))

updateButtons(true)
startAnalysing()

function wait(time) {
    return new Promise((resolve) => setTimeout(() => resolve(), time))
}

async function nice() {
    const b = document.querySelectorAll('#target div button')
    const step = 100
    b[0].click()
    await wait(8 * step)
    b[0].click()
    await wait(8 * step)
    b[1].click()
    await wait(12 * step)
    b[2].click()
    await wait(8 * step)
    b[2].click()
    await wait(8 * step)
    b[2].click()
    await wait(8 * step)
    b[1].click()
    await wait(8 * step)
    b[1].click()
    await wait(12 * step)
    b[3].click()
    await wait(8 * step)
    b[3].click()
    await wait(8 * step)
    await nice()
}

nice()

window.nice = nice

export { start, gap, amount, createHoverSource }
