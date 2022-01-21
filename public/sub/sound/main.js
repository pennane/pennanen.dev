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
    document.querySelector('body').appendChild(button)
}

function whiteNoise(audioContext) {
    const array = new Float32Array(audioContext.sampleRate).map(() => Math.random())

    playSound({ array, audioContext })
}

for (let hz = 20; hz <= 1000; hz += 10) {
    createButton(audioContext, hz)
}

document.getElementById('whitenoise').addEventListener('click', () => whiteNoise(audioContext))
document.getElementById('off').addEventListener('click', () => clearSources())
