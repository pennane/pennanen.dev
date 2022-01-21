const gapEl = document.getElementById('gap')
const amountEl = document.getElementById('amount')
const startEl = document.getElementById('start')

const audioContext = new AudioContext({ sampleRate: 44100 })
const sampleRate = audioContext.sampleRate
const gainNode = audioContext.createGain()
gainNode.gain.value = 0.1
gainNode.connect(audioContext.destination)

let sources = []

function clearSources() {
    for (const source of sources) {
        source.stop()
    }
    sources = []
}

function* sineWave(audioContext, frequency) {
    const sampleFrequency = audioContext.sampleRate / frequency
    let i = 0
    while (true) {
        i++
        yield Math.sin(i / (sampleFrequency / (Math.PI * 2)))
    }
}

function playSound({ audioContext, array }) {
    const audioBuffer = audioContext.createBuffer(1, array.length, audioContext.sampleRate)
    audioBuffer.copyToChannel(array, 0)
    // clearSources()
    const source = audioContext.createBufferSource()
    source.loop = true
    source.connect(gainNode)
    source.buffer = audioBuffer
    source.start()
    sources.push(source)
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
    const array = new Float32Array(audioContext.sampleRate).map(() => Math.random())

    playSound({ array, audioContext })
}

let start = 80,
    amount = 10,
    gap = 40

function createButtons(initial) {
    if (!initial) {
        gap = Math.max(gapEl.value, 1)
        start = Math.min(Math.max(startEl.value, 1), 22000)
        amount = Math.min(Math.max(amountEl.value, 1), 200)
    }

    gapEl.value = gap
    startEl.value = start
    amountEl.value = amount

    const target = document.getElementById('target')
    while (target.firstChild) {
        target.removeChild(target.firstChild)
    }

    const container = document.createElement('div')

    let hz = start
    for (let i = 0; i < amount; i++) {
        container.appendChild(createButton(audioContext, hz + i * gap))
    }

    target.appendChild(container)
}

createButtons(true)

document.getElementById('whitenoise').addEventListener('click', () => whiteNoise(audioContext))
document.getElementById('off').addEventListener('click', () => clearSources())
document.getElementById('reset').addEventListener('click', () => createButtons())

document.addEventListener('keydown', (e) => {
    if (e.target.nodeName !== 'BODY' || e.repeat) return
    if (e.key === 'r' || e.key === 'R') {
        clearSources()
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
    if (e.target.nodeName !== 'BODY' || e.repeat) return
    let index = parseInt(e.key)
    if (isNaN(index)) return
    if (index === 0) index = 10
    const button = document.querySelectorAll('#target div button')[index - 1]
    if (!button) return
    button.classList.remove('pressed')
})
